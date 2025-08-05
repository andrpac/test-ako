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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/api"
	akov2 "github.com/mongodb/mongodb-atlas-kubernetes/v2/api/v1"
)

const (
	InternalNameSeparator = "$"
)

type Identifiers struct {
	ProjectID        string
	ProjectName      string
	ClusterName      string
	DatabaseUsername string
}

// ExtractIdentifiers tries both internal and K8s name formats.
func ExtractIdentifiers(ctx context.Context, c client.Client, req types.NamespacedName) (Identifiers, error) {
	if strings.Contains(req.Name, InternalNameSeparator) {
		return parseIdentifiersInternalFormat(req.Name)
	}
	return parseIdentifiersK8sFormat(ctx, c, req)
}

// === Internal Format ===
// Format: $<projectID>$<clusterName>$<databaseUsername>
func parseIdentifiersInternalFormat(name string) (Identifiers, error) {
	parts := strings.Split(name, InternalNameSeparator)
	if len(parts) != 3 {
		return Identifiers{}, fmt.Errorf("invalid internal connection secret name: %q", name)
	}
	return Identifiers{
		ProjectID:        parts[0],
		ClusterName:      parts[1],
		DatabaseUsername: parts[2],
	}, nil
}

// === K8s Format ===
// Format: <projectName>-<clusterName>-<databaseUsername>
// Requires looking up labels on the Secret.
func parseIdentifiersK8sFormat(ctx context.Context, c client.Client, req types.NamespacedName) (Identifiers, error) {
	var secret corev1.Secret
	if err := c.Get(ctx, req, &secret); err != nil {
		return Identifiers{}, fmt.Errorf("failed to retrieve ConnectionSecret: %w", err)
	}

	labels := secret.GetLabels()
	projectID, hasProject := labels[ProjectLabelKey]
	clusterName, hasCluster := labels[ClusterLabelKey]

	if !hasProject {
		return Identifiers{}, fmt.Errorf("missing label %q on ConnectionSecret", ProjectLabelKey)
	}
	if !hasCluster {
		return Identifiers{}, fmt.Errorf("missing label %q on ConnectionSecret", ClusterLabelKey)
	}

	parts := strings.Split(req.Name, "-"+clusterName+"-")
	if len(parts) < 2 {
		return Identifiers{}, fmt.Errorf("failed to parse databaseUsername from legacy format: %q", req.Name)
	}

	return Identifiers{
		ProjectID:        projectID,
		ProjectName:      parts[0],
		ClusterName:      clusterName,
		DatabaseUsername: parts[1],
	}, nil
}

func IdentifierToK8sFormat(ids Identifiers) string {
	if ids.ProjectName != "" {
		return ids.ProjectName + "-" + ids.ClusterName + "-" + ids.DatabaseUsername
	}

	return ""
}

func IdentifiersToInternalFormat(ids Identifiers) string {
	return ids.ProjectID + InternalNameSeparator + ids.ClusterName + InternalNameSeparator + ids.DatabaseUsername
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
