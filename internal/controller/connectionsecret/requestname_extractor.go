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
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	akov2 "github.com/mongodb/mongodb-atlas-kubernetes/v2/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/indexer"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/kube"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/translation/project"
)

const InternalSeparator = "$"

var (
	// Parsing & format
	ErrInternalFormatPartsInvalid = errors.New("internal format expected 3 parts separated by $")
	ErrInternalFormatPartEmpty    = errors.New("internal format got empty value in one or more parts")
	ErrK8sLabelsMissing           = errors.New("k8s format got a missing required label(s)")
	ErrK8sLabelEmpty              = errors.New("k8s format got label present but empty")
	ErrK8sNameSplitInvalid        = errors.New("k8s format expected to separate across -<clusterName>-")
	ErrK8sNameSplitEmpty          = errors.New("k8s format got empty value in one or more parts")

	// Index lookups
	ErrNoDeploymentFound = errors.New("no AtlasDeployment found")
	ErrManyDeployments   = errors.New("multiple AtlasDeployments found")
	ErrNoUserFound       = errors.New("no AtlasDatabaseUser found")
	ErrManyUsers         = errors.New("multiple AtlasDatabaseUsers found")
)

// ConnSecretIdentifiers holds the values extracted from a reconcile request name.
type ConnSecretIdentifiers struct {
	ProjectID        string
	ProjectName      string
	ClusterName      string
	DatabaseUsername string
}

// ConnSecretPair represents the pairing of an AtlasDeployment and an AtlasDatabaseUser
// required to construct a ConnectionSecret. It holds resolved identifiers and the corresponding resources.
// NOTE: this struct intentionally stores only ProjectID (not all identifiers) to keep only necessary information.
type ConnSecretPair struct {
	ProjectID  string
	Deployment *akov2.AtlasDeployment
	User       *akov2.AtlasDatabaseUser
}

// CreateK8sFormat returns the Secret name in the Kubernetes naming format: <projectName>-<clusterName>-<username>
func CreateK8sFormat(projectName string, clusterName string, databaseUsername string) string {
	return strings.Join([]string{
		kube.NormalizeIdentifier(projectName),
		kube.NormalizeIdentifier(clusterName),
		kube.NormalizeIdentifier(databaseUsername),
	}, "-")
}

// CreateInternalFormat returns the Secret name in the internal format used by watchers: <projectID>$<clusterName>$<username>
func CreateInternalFormat(projectID string, clusterName string, databaseUsername string) string {
	return strings.Join([]string{
		projectID,
		kube.NormalizeIdentifier(clusterName),
		kube.NormalizeIdentifier(databaseUsername),
	}, InternalSeparator)
}

// LoadRequestIdentifiers determines whether the request name is internal or K8s format
// and extracts ProjectID, ClusterName, and DatabaseUsername.
func LoadRequestIdentifiers(ctx context.Context, c client.Client, req types.NamespacedName) (ConnSecretIdentifiers, error) {
	var ids ConnSecretIdentifiers

	// === Internal format: <ProjectID>$<ClusterName>$<DatabaseUserName>
	if strings.Contains(req.Name, InternalSeparator) {
		parts := strings.SplitN(req.Name, InternalSeparator, 3)
		if len(parts) != 3 {
			return ids, ErrInternalFormatPartsInvalid
		}
		if parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return ids, ErrInternalFormatPartEmpty
		}
		return ConnSecretIdentifiers{
			ProjectID:        parts[0],
			ClusterName:      parts[1],
			DatabaseUsername: parts[2],
		}, nil
	}

	// === K8s format: <ProjectName>-<ClusterName>-<DatabaseUserName>
	var secret corev1.Secret
	if err := c.Get(ctx, req, &secret); err != nil {
		return ids, err
	}

	labels := secret.GetLabels()
	projectID, hasProject := labels[ProjectLabelKey]
	clusterName, hasCluster := labels[ClusterLabelKey]

	// Missing labels or values
	if !hasProject || !hasCluster {
		return ids, ErrK8sLabelsMissing
	}
	if projectID == "" || clusterName == "" {
		return ids, ErrK8sLabelEmpty
	}

	sep := fmt.Sprintf("-%s-", clusterName)
	parts := strings.SplitN(req.Name, sep, 2)
	if len(parts) != 2 {
		return ids, ErrK8sNameSplitInvalid
	}
	if parts[0] == "" || parts[1] == "" {
		return ids, ErrK8sNameSplitEmpty
	}

	return ConnSecretIdentifiers{
		ProjectID:        projectID,
		ProjectName:      parts[0],
		ClusterName:      clusterName,
		DatabaseUsername: parts[1],
	}, nil
}

// LoadPairedResources fetches the paired AtlasDeployment and AtlasDatabaseUser forming the ConnectionSecret
// using the registered indexers
func LoadPairedResources(ctx context.Context, c client.Client, ids ConnSecretIdentifiers, namespace string) (*ConnSecretPair, error) {
	compositeDeploymentKey := ids.ProjectID + "-" + ids.ClusterName
	deployments := &akov2.AtlasDeploymentList{}
	if err := c.List(ctx, deployments, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(indexer.AtlasDeploymentBySpecNameAndProjectID, compositeDeploymentKey),
		Namespace:     namespace,
	}); err != nil {
		return nil, err
	}
	switch len(deployments.Items) {
	case 0:
		return nil, ErrNoDeploymentFound
	case 1:
		// ok
	default:
		return nil, ErrManyDeployments
	}

	compositeUserKey := ids.ProjectID + "-" + ids.DatabaseUsername
	users := &akov2.AtlasDatabaseUserList{}
	if err := c.List(ctx, users, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(indexer.AtlasDatabaseUserBySpecUsernameAndProjectID, compositeUserKey),
		Namespace:     namespace,
	}); err != nil {
		return nil, err
	}
	switch len(users.Items) {
	case 0:
		return nil, ErrNoUserFound
	case 1:
		// ok
	default:
		return nil, ErrManyUsers
	}

	return &ConnSecretPair{
		ProjectID:  ids.ProjectID,
		Deployment: &deployments.Items[0],
		User:       &users.Items[0],
	}, nil
}

// IsDeleting checks if either the AtlasDeployment or AtlasDatabaseUser are getting deleted
func (p *ConnSecretPair) IsDeleting() bool {
	isDeploymentedDeleted := !p.Deployment.GetDeletionTimestamp().IsZero()
	isUserDeleted := !p.User.GetDeletionTimestamp().IsZero()

	if isDeploymentedDeleted || isUserDeleted {
		return true
	}
	return false
}

// IsReady checks that both AtlasDeployment and AtlasDatabaseUser are ready
func (p *ConnSecretPair) IsReady() (bool, []string) {
	notReady := []string{}

	if !IsDeploymentReady(p.Deployment) {
		notReady = append(notReady, fmt.Sprintf("AtlasDeployment/%s", p.Deployment.GetName()))
	}
	if !IsDatabaseUserReady(p.User) {
		notReady = append(notReady, fmt.Sprintf("AtlasDatabaseUser/%s", p.User.GetName()))
	}

	return len(notReady) == 0, notReady
}

// NeedsSDKProjectResolution checks if we need to use SDK to retrieve the projectName
func (p *ConnSecretPair) NeedsSDKProjectResolution() bool {
	return p.Deployment.Spec.ExternalProjectRef != nil &&
		p.Deployment.Spec.ExternalProjectRef.ID != "" &&
		p.User.Spec.ExternalProjectRef != nil &&
		p.User.Spec.ExternalProjectRef.ID != ""
}

// ResolveProjectNameK8s retrieves the ProjectName by K8s AtlasProject resource
func (p *ConnSecretPair) ResolveProjectNameK8s(ctx context.Context, c client.Client, namespace string) (string, error) {
	var name string
	if p.Deployment.Spec.ExternalProjectRef == nil {
		name = p.Deployment.Spec.ProjectRef.Name
	} else {
		name = p.User.Spec.ProjectRef.Name
	}

	proj := &akov2.AtlasProject{}
	if err := c.Get(ctx, kube.ObjectKey(namespace, name), proj); err != nil {
		return "", fmt.Errorf("failed to retrieve AtlasProject %q: %w", name, err)
	}

	return kube.NormalizeIdentifier(proj.Spec.Name), nil
}

// ResolveProjectNameSDK resolves the ProjectName by calling the Atlas SDK
func (p *ConnSecretPair) ResolveProjectNameSDK(ctx context.Context, projectService project.ProjectService) (string, error) {
	atlasProject, err := projectService.GetProject(ctx, p.ProjectID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch project from Atlas API for %q: %w", p.ProjectID, err)
	}

	return kube.NormalizeIdentifier(atlasProject.Name), nil
}

// BuildConnectionData constructs the secret data that will be passed in the secret
func (p *ConnSecretPair) BuildConnectionData(ctx context.Context, c client.Client) (ConnectionData, error) {
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

// HandleSecret ensures the ConnectionSecret resource is created or updated with the given connection data.
// It wraps the Ensure(...) helper and constructs the secret name from the provided parameters.
// NOTE: since ConnSecretPair no longer stores all identifiers, caller must pass projectName and clusterName.
func (p *ConnSecretPair) HandleSecret(ctx context.Context, c client.Client, ids ConnSecretIdentifiers, data ConnectionData) error {
	_, err := Ensure(ctx, c, p.User.Namespace, ids.ProjectName, p.ProjectID, ids.ClusterName, data)
	if err != nil {
		return fmt.Errorf("ensure failed for secret (projectName=%q, clusterName=%q, user=%q): %w",
			ids.ProjectName, ids.ClusterName, p.User.Spec.Username, err)
	}
	return nil
}
