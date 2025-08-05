// Copyright 2025 MongoDB Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package connectionsecret

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/api"
	akov2 "github.com/mongodb/mongodb-atlas-kubernetes/v2/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/indexer"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/kube"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/translation/project"
)

const InternalSeparator = "$"

// ConnectionSecretContext represents the pairing of an AtlasDeployment and an AtlasDatabaseUser
// required to construct a ConnectionSecret. It holds resolved identifiers and the corresponding resources.
type ConnectionSecretContext struct {
	ProjectID        string
	ProjectName      string
	ClusterName      string
	DatabaseUsername string

	Deployment *akov2.AtlasDeployment
	User       *akov2.AtlasDatabaseUser
}

// ParseRequestName parses the named resource sent for reconcilation and always extracts ProjectID, ClusterName, and DatabaseUsername
func (p *ConnectionSecretContext) ParseRequestName(ctx context.Context, c client.Client, req types.NamespacedName) error {

	// === Format Internal: <ProjectID>$<ClusterName>$<DatabaseUserName>
	// Split request name around `$` to parse the ProjectID, ClusterName, DatabaseUserName
	// ProjectName will be set to empty and filled-in later
	if strings.Contains(req.Name, InternalSeparator) {
		parts := strings.SplitN(req.Name, InternalSeparator, 3)

		// Error out on improper format
		if len(parts) != 3 {
			return fmt.Errorf(
				"invalid internal name format for %q: expected 3 parts separated by '%s'",
				req.Name, InternalSeparator,
			)
		}
		if parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return fmt.Errorf(
				"invalid internal name format for %q: empty value in one or more parts "+
					"(projectID=%q, clusterName=%q, databaseUsername=%q)",
				req.Name, parts[0], parts[1], parts[2],
			)
		}

		p.ProjectID = parts[0]
		p.ClusterName = parts[1]
		p.DatabaseUsername = parts[2]
		return nil
	}

	// === Format K8s: <ProjectName>-<ClusterName>-<DatabaseUserName>
	// Parse ProjectID and ClusterName from secret labels
	// Parse ProjectName and DatabaseUsername from request name by splitting around -<ClusterName>-
	var secret corev1.Secret
	if err := c.Get(ctx, req, &secret); err != nil {
		return fmt.Errorf(
			"unable to retrieve Secret %q in namespace %q: %w",
			req.Name, req.Namespace, err,
		)
	}

	labels := secret.GetLabels()
	projectID, hasProject := labels[ProjectLabelKey]
	clusterName, hasCluster := labels[ClusterLabelKey]

	// Error out on missing labels, or empty fields
	var missing []string
	if !hasProject {
		missing = append(missing, ProjectLabelKey)
	}
	if !hasCluster {
		missing = append(missing, ClusterLabelKey)
	}
	if len(missing) > 0 {
		return fmt.Errorf(
			"secret %q is missing required label(s): %v. Current labels: %v",
			req.Name, missing, labels,
		)
	}

	if projectID == "" {
		return fmt.Errorf(
			"secret %q has empty value for label %q",
			req.Name, ProjectLabelKey,
		)
	}
	if clusterName == "" {
		return fmt.Errorf(
			"secret %q has empty value for label %q",
			req.Name, ClusterLabelKey,
		)
	}

	// Error out on improper format
	sep := fmt.Sprintf("-%s-", clusterName)
	parts := strings.SplitN(req.Name, sep, 2)
	if len(parts) != 2 {
		return fmt.Errorf(
			"invalid K8s name format for Secret %q: expected separator '-%s-' not found",
			req.Name, clusterName,
		)
	}

	if parts[0] == "" || parts[1] == "" {
		return fmt.Errorf(
			"invalid K8s name format for %q: empty value in one or more parts "+
				"(projectName=%q, databaseUsername=%q)",
			req.Name, parts[0], parts[1],
		)
	}

	p.ProjectID = projectID
	p.ProjectName = parts[0]
	p.ClusterName = clusterName
	p.DatabaseUsername = parts[1]
	return nil
}

// LoadDeploymentAndUser retrives the paired AtlasDeployment and AtlasDatabaseUser forming the ConnectionSecret
func (p *ConnectionSecretContext) LoadDeploymentAndUser(ctx context.Context, c client.Client, namespace string) error {
	// Retrives the AtlasDeployment by composite key <ProjectID>-<ClusterName> of indexer
	compositeDeploymentKey := p.ProjectID + "-" + p.ClusterName
	deployments := &akov2.AtlasDeploymentList{}
	if err := c.List(ctx, deployments, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(indexer.AtlasDeploymentBySpecNameAndProjectID, compositeDeploymentKey),
		Namespace:     namespace,
	}); err != nil {
		return fmt.Errorf("failed to list AtlasDeployments: %w", err)
	}
	if len(deployments.Items) != 1 {
		return fmt.Errorf("expected 1 AtlasDeployment for %q, found %d", compositeDeploymentKey, len(deployments.Items))
	}

	// Retrives the AtlasDeployment by composite key <ProjectID>-<DatabaseUsername> of indexer
	compositeUserKey := p.ProjectID + "-" + p.DatabaseUsername
	users := &akov2.AtlasDatabaseUserList{}
	if err := c.List(ctx, users, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(indexer.AtlasDatabaseUserBySpecUsernameAndProjectID, compositeUserKey),
		Namespace:     namespace,
	}); err != nil {
		return fmt.Errorf("failed to list AtlasDatabaseUsers: %w", err)
	}
	if len(users.Items) != 1 {
		return fmt.Errorf("expected 1 AtlasDatabaseUser for %q, found %d", compositeUserKey, len(users.Items))
	}

	p.Deployment = &deployments.Items[0]
	p.User = &users.Items[0]
	return nil
}

// AreResourcesReady checks that both resources are ready
func (p *ConnectionSecretContext) AreResourcesReady() (bool, []string) {
	notReady := []string{}

	if !IsDeploymentReady(p.Deployment) {
		notReady = append(notReady, fmt.Sprintf("AtlasDeployment/%s", p.Deployment.GetName()))
	}
	if !IsDatabaseUserReady(p.User) {
		notReady = append(notReady, fmt.Sprintf("AtlasDatabaseUser/%s", p.User.GetName()))
	}

	return len(notReady) == 0, notReady
}

// HasProjectName returns true if ProjectName is populated (required for naming the Secret)
func (p *ConnectionSecretContext) HasProjectName() bool {
	return p.ProjectName != ""
}

// ShouldUseSDKResolution checks if we need to use SDK to retrive the projectName
func (p *ConnectionSecretContext) ShouldUseSDKResolution() bool {
	return p.Deployment.Spec.ExternalProjectRef != nil &&
		p.Deployment.Spec.ExternalProjectRef.ID != "" &&
		p.User.Spec.ExternalProjectRef != nil &&
		p.User.Spec.ExternalProjectRef.ID != ""
}

// ResolveProjectNameK8s retrives the ProjectName by K8s AtlasProject resource
func (p *ConnectionSecretContext) ResolveProjectNameK8s(ctx context.Context, c client.Client, namespace string) error {
	hasExternalRefDeployment := p.Deployment.Spec.ExternalProjectRef != nil && p.Deployment.Spec.ExternalProjectRef.ID != ""

	var project *akov2.AtlasProject
	var name string

	if !hasExternalRefDeployment {
		name = p.Deployment.Spec.ProjectRef.Name
	} else {
		name = p.User.Spec.ProjectRef.Name
	}

	project = &akov2.AtlasProject{}
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, project); err != nil {
		return fmt.Errorf("failed to retrieve AtlasProject %q: %w", name, err)
	}

	p.ProjectName = kube.NormalizeIdentifier(project.Spec.Name)
	return nil
}

// ResolveProjectNameSDK resolves the ProjectName by calling the Atlas SDK
func (p *ConnectionSecretContext) ResolveProjectNameSDK(ctx context.Context, projectService project.ProjectService) error {
	atlasProject, err := projectService.GetProject(ctx, p.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to fetch project from Atlas API for %q: %w", p.ProjectID, err)
	}

	p.ProjectName = kube.NormalizeIdentifier(atlasProject.Name)
	return nil
}

// BuildConnectionData constructs the secret data that will be passed in the secret
func (p *ConnectionSecretContext) BuildConnectionData(ctx context.Context, c client.Client) (ConnectionData, error) {
	password, err := p.User.ReadPassword(ctx, c)
	if err != nil {
		return ConnectionData{}, fmt.Errorf("failed to read password for user %q: %w", p.User.Spec.Username, err)
	}

	conn := p.Deployment.Status.ConnectionStrings

	data := ConnectionData{
		DBUserName: p.User.Spec.Username,
		Password:   password,
		ConnURL:    conn.Standard,
		SrvConnURL: conn.StandardSrv,
	}

	if conn.Private != "" {
		data.PrivateConnURLs = append(data.PrivateConnURLs, PrivateLinkConnURLs{
			PvtConnURL:    conn.Private,
			PvtSrvConnURL: conn.PrivateSrv,
		})
	}

	for _, pe := range conn.PrivateEndpoint {
		data.PrivateConnURLs = append(data.PrivateConnURLs, PrivateLinkConnURLs{
			PvtConnURL:      pe.ConnectionString,
			PvtSrvConnURL:   pe.SRVConnectionString,
			PvtShardConnURL: pe.SRVShardOptimizedConnectionString,
		})
	}

	return data, nil
}

// handleSecret ensures the ConnectionSecret resource is created or updated with the given connection data.
// It wraps the Ensure(...) helper and constructs the secret name from the internal pairing.
func (p *ConnectionSecretContext) handleSecret(ctx context.Context, c client.Client, data ConnectionData) error {
	_, err := Ensure(ctx, c, p.User.Namespace, p.ProjectName, p.ProjectID, p.ClusterName, data)
	if err != nil {
		return fmt.Errorf("ensure failed for secret (projectName=%q, clusterName=%q, user=%q): %w",
			p.ProjectName, p.ClusterName, p.User.Spec.Username, err)
	}
	return nil
}

// CreateK8sFormat returns the Secret name in the Kubernetes naming format: <projectName>-<clusterName>-<username>
func CreateK8sFormat(ProjectName string, ClusterName string, DatabaseUsername string) string {
	ProjectName = kube.NormalizeIdentifier(ProjectName)
	ClusterName = kube.NormalizeIdentifier(ClusterName)
	DatabaseUsername = kube.NormalizeIdentifier(DatabaseUsername)
	return strings.Join([]string{ProjectName, ClusterName, DatabaseUsername}, "-")
}

// CreateInternalFormat returns the Secret name in the internal format used by watchers: <projectID>$<clusterName>$<username>
func CreateInternalFormat(ProjectID string, ClusterName string, DatabaseUsername string) string {
	ClusterName = kube.NormalizeIdentifier(ClusterName)
	DatabaseUsername = kube.NormalizeIdentifier(DatabaseUsername)
	return strings.Join([]string{ProjectID, ClusterName, DatabaseUsername}, InternalSeparator)
}

func IsDeploymentReady(d *akov2.AtlasDeployment) bool {
	for _, c := range d.Status.Conditions {
		if c.Type == api.DeploymentReadyType && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsDatabaseUserReady(u *akov2.AtlasDatabaseUser) bool {
	for _, c := range u.Status.Conditions {
		if c.Type == api.DatabaseUserReadyType && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func ResolveProjectIDFromDeployment(ctx context.Context, c client.Client, d *akov2.AtlasDeployment) (string, error) {
	if d.Spec.ExternalProjectRef != nil && d.Spec.ExternalProjectRef.ID != "" {
		return d.Spec.ExternalProjectRef.ID, nil
	}

	if d.Spec.ProjectRef != nil && d.Spec.ProjectRef.Name != "" {
		project := &akov2.AtlasProject{}
		if err := c.Get(ctx, *d.Spec.ProjectRef.GetObject(d.Namespace), project); err != nil {
			return "", fmt.Errorf("failed to resolve projectRef from AtlasDeployment: %w", err)
		}
		return project.ID(), nil
	}

	return "", fmt.Errorf("missing both external and internal project references in AtlasDeployment")
}

func ResolveProjectIDFromDatabaseUser(ctx context.Context, c client.Client, u *akov2.AtlasDatabaseUser) (string, error) {
	if u.Spec.ExternalProjectRef != nil && u.Spec.ExternalProjectRef.ID != "" {
		return u.Spec.ExternalProjectRef.ID, nil
	}

	if u.Spec.ProjectRef != nil && u.Spec.ProjectRef.Name != "" {
		project := &akov2.AtlasProject{}
		if err := c.Get(ctx, *u.Spec.ProjectRef.GetObject(u.Namespace), project); err != nil {
			return "", fmt.Errorf("failed to resolve projectRef from AtlasDatabaseUser: %w", err)
		}
		return project.ID(), nil
	}

	return "", fmt.Errorf("missing both external and internal project references in AtlasDatabaseUser")
}
