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
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/api"
	akov2 "github.com/mongodb/mongodb-atlas-kubernetes/v2/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/api/v1/common"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/api/v1/status"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/indexer"
)

func TestHasProjectName(t *testing.T) {
	t.Run("HasProjectName returns true when ProjectName is set", func(t *testing.T) {
		p := &ConnectionSecretContext{ProjectName: "my-project"}
		assert.True(t, p.HasProjectName())
	})
	t.Run("HasProjectName returns false when ProjectName is empty", func(t *testing.T) {
		p := &ConnectionSecretContext{}
		assert.False(t, p.HasProjectName())
	})
}

func TestShouldUseSDKResolution(t *testing.T) {
	t.Run("ShouldUseSDKResolution returns true if both have ExternalProjectRef.ID", func(t *testing.T) {
		p := &ConnectionSecretContext{
			Deployment: &akov2.AtlasDeployment{
				Spec: akov2.AtlasDeploymentSpec{
					ProjectDualReference: akov2.ProjectDualReference{
						ExternalProjectRef: &akov2.ExternalProjectReference{
							ID: "abc",
						},
					},
				},
			},
			User: &akov2.AtlasDatabaseUser{
				Spec: akov2.AtlasDatabaseUserSpec{
					ProjectDualReference: akov2.ProjectDualReference{
						ExternalProjectRef: &akov2.ExternalProjectReference{
							ID: "abc",
						},
					},
				},
			},
		}
		assert.True(t, p.ShouldUseSDKResolution())
	})

	t.Run("ShouldUseSDKResolution returns false if one is missing", func(t *testing.T) {
		p := &ConnectionSecretContext{
			Deployment: &akov2.AtlasDeployment{
				Spec: akov2.AtlasDeploymentSpec{
					ProjectDualReference: akov2.ProjectDualReference{
						ExternalProjectRef: &akov2.ExternalProjectReference{
							ID: "abc",
						},
					},
				},
			},
			User: &akov2.AtlasDatabaseUser{
				Spec: akov2.AtlasDatabaseUserSpec{
					ProjectDualReference: akov2.ProjectDualReference{
						ProjectRef: &common.ResourceRefNamespaced{
							Name:      "my-project",
							Namespace: "my-namespace",
						},
					},
				},
			},
		}
		assert.False(t, p.ShouldUseSDKResolution())
	})
}

func TestAreResourcesReady(t *testing.T) {
	t.Run("Both resources ready returns true", func(t *testing.T) {
		p := &ConnectionSecretContext{
			Deployment: &akov2.AtlasDeployment{
				ObjectMeta: metav1.ObjectMeta{Name: "dep"},
				Status: status.AtlasDeploymentStatus{
					Common: api.Common{
						Conditions: []api.Condition{
							{Type: api.DeploymentReadyType, Status: corev1.ConditionTrue},
						},
					},
				},
			},
			User: &akov2.AtlasDatabaseUser{
				ObjectMeta: metav1.ObjectMeta{Name: "user"},
				Status: status.AtlasDatabaseUserStatus{
					Common: api.Common{
						Conditions: []api.Condition{
							{Type: api.DatabaseUserReadyType, Status: corev1.ConditionTrue},
						},
					},
				},
			},
		}
		ok, notReady := p.AreResourcesReady()
		assert.True(t, ok)
		assert.Empty(t, notReady)
	})

	t.Run("One resource not ready returns false with correct message", func(t *testing.T) {
		p := &ConnectionSecretContext{
			Deployment: &akov2.AtlasDeployment{
				ObjectMeta: metav1.ObjectMeta{Name: "dep"},
				Status: status.AtlasDeploymentStatus{
					Common: api.Common{
						Conditions: []api.Condition{
							{Type: api.DeploymentReadyType, Status: corev1.ConditionTrue},
						},
					},
				},
			},
			User: &akov2.AtlasDatabaseUser{
				ObjectMeta: metav1.ObjectMeta{Name: "user"},
				Status: status.AtlasDatabaseUserStatus{
					Common: api.Common{
						Conditions: []api.Condition{
							{Type: api.DatabaseUserReadyType, Status: corev1.ConditionFalse},
						},
					},
				},
			},
		}
		ok, notReady := p.AreResourcesReady()
		assert.False(t, ok)
		assert.Equal(t, []string{"AtlasDatabaseUser/user"}, notReady)
	})
}

func TestParseRequestName(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))

	secretValid := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myproj-mycluster-admin",
			Namespace: "default",
			Labels: map[string]string{
				ProjectLabelKey: "proj123",
				ClusterLabelKey: "mycluster",
			},
		},
	}
	secretMissingLabels := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "missing-mycluster-admin",
			Namespace: "default",
			Labels:    map[string]string{},
		},
	}
	secretEmptyLabel := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "emptylabel-mycluster-admin",
			Namespace: "default",
			Labels: map[string]string{
				ProjectLabelKey: "",
				ClusterLabelKey: "mycluster",
			},
		},
	}
	secretBadSplit := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "-mycluster-admin",
			Namespace: "default",
			Labels: map[string]string{
				ProjectLabelKey: "proj123",
				ClusterLabelKey: "mycluster",
			},
		},
	}
	secretInvalidSep := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "invalid-separator",
			Namespace: "default",
			Labels: map[string]string{
				ProjectLabelKey: "proj123",
				ClusterLabelKey: "unknown",
			},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(
			secretValid,
			secretMissingLabels,
			secretEmptyLabel,
			secretBadSplit,
			secretInvalidSep,
		).
		Build()

	tests := map[string]struct {
		name      string
		namespace string
		expected  *ConnectionSecretContext
		errSubstr string
	}{
		"valid internal format": {
			name: "proj123$mycluster$admin",
			expected: &ConnectionSecretContext{
				ProjectID:        "proj123",
				ClusterName:      "mycluster",
				DatabaseUsername: "admin",
			},
		},
		"internal format with too few parts": {
			name:      "proj123$clusterOnly",
			errSubstr: "expected 3 parts",
		},
		"internal format with empty part": {
			name:      "proj123$$admin",
			errSubstr: "empty value in one or more parts",
		},
		"valid k8s format": {
			name:      "myproj-mycluster-admin",
			namespace: "default",
			expected: &ConnectionSecretContext{
				ProjectID:        "proj123",
				ProjectName:      "myproj",
				ClusterName:      "mycluster",
				DatabaseUsername: "admin",
			},
		},
		"k8s format with missing secret": {
			name:      "nonexistent-secret",
			namespace: "default",
			errSubstr: "unable to retrieve Secret",
		},
		"k8s format with missing labels": {
			name:      "missing-mycluster-admin",
			namespace: "default",
			errSubstr: "missing required label",
		},
		"k8s format with empty label value": {
			name:      "emptylabel-mycluster-admin",
			namespace: "default",
			errSubstr: "has empty value for label",
		},
		"k8s format with invalid name separator": {
			name:      "invalid-separator",
			namespace: "default",
			errSubstr: "expected separator",
		},
		"k8s format with empty value after split": {
			name:      "-mycluster-admin",
			namespace: "default",
			errSubstr: "empty value in one or more parts",
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			p := &ConnectionSecretContext{}
			err := p.ParseRequestName(context.Background(), client, types.NamespacedName{
				Name:      tc.name,
				Namespace: tc.namespace,
			})

			if tc.errSubstr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errSubstr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.ProjectID, p.ProjectID)
				assert.Equal(t, tc.expected.ProjectName, p.ProjectName)
				assert.Equal(t, tc.expected.ClusterName, p.ClusterName)
				assert.Equal(t, tc.expected.DatabaseUsername, p.DatabaseUsername)
			}
		})
	}
}

func TestLoadDeploymentAndUser(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(akov2.AddToScheme(scheme))

	const (
		ns             = "default"
		projectID      = "proj123"
		otherprojectID = "proj456"
	)

	tests := map[string]struct {
		clusterName      string
		databaseUsername string
		deployments      []client.Object
		users            []client.Object
		expectedErr      string
	}{
		"successfully finds one deployment and one user": {
			clusterName:      "clusterA",
			databaseUsername: "admin",
			deployments: []client.Object{
				&akov2.AtlasDeployment{
					ObjectMeta: metav1.ObjectMeta{Name: "dep1", Namespace: ns},
					Spec: akov2.AtlasDeploymentSpec{
						DeploymentSpec: &akov2.AdvancedDeploymentSpec{
							Name: "clusterA",
						},
					},
				},
			},
			users: []client.Object{
				&akov2.AtlasDatabaseUser{
					ObjectMeta: metav1.ObjectMeta{Name: "user1", Namespace: ns},
					Spec: akov2.AtlasDatabaseUserSpec{
						Username: "admin",
					},
				},
			},
		},
		"no deployments found overall": {
			clusterName:      "clusterA",
			databaseUsername: "admin",
			users: []client.Object{
				&akov2.AtlasDatabaseUser{
					ObjectMeta: metav1.ObjectMeta{Name: "user1", Namespace: ns},
					Spec: akov2.AtlasDatabaseUserSpec{
						Username: "admin",
					},
				},
			},
			expectedErr: `expected 1 AtlasDeployment for "proj123-clusterA", found 0`,
		},
		"no deployments found due to missing index": {
			clusterName:      "clusterB",
			databaseUsername: "admin",
			deployments: []client.Object{
				&akov2.AtlasDeployment{
					ObjectMeta: metav1.ObjectMeta{Name: "dep1", Namespace: ns},
					Spec: akov2.AtlasDeploymentSpec{
						DeploymentSpec: &akov2.AdvancedDeploymentSpec{
							Name: "clusterB",
						},
					},
				},
			},
			users: []client.Object{
				&akov2.AtlasDatabaseUser{
					ObjectMeta: metav1.ObjectMeta{Name: "user1", Namespace: ns},
					Spec: akov2.AtlasDatabaseUserSpec{
						Username: "admin",
					},
				},
			},
			expectedErr: `expected 1 AtlasDeployment for "proj123-clusterB", found 0`,
		},
		"multiple users found": {
			clusterName:      "clusterA",
			databaseUsername: "admin",
			deployments: []client.Object{
				&akov2.AtlasDeployment{
					ObjectMeta: metav1.ObjectMeta{Name: "dep1", Namespace: ns},
					Spec: akov2.AtlasDeploymentSpec{
						DeploymentSpec: &akov2.AdvancedDeploymentSpec{
							Name: "clusterA",
						},
					},
				},
			},
			users: []client.Object{
				&akov2.AtlasDatabaseUser{
					ObjectMeta: metav1.ObjectMeta{Name: "user1", Namespace: ns},
					Spec: akov2.AtlasDatabaseUserSpec{
						Username: "admin",
					},
				},
				&akov2.AtlasDatabaseUser{
					ObjectMeta: metav1.ObjectMeta{Name: "user2", Namespace: ns},
					Spec: akov2.AtlasDatabaseUserSpec{
						Username: "admin",
					},
				},
			},
			expectedErr: `expected 1 AtlasDatabaseUser for "proj123-admin", found 2`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			allObjects := append(tt.deployments, tt.users...)

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(allObjects...).
				WithIndex(&akov2.AtlasDeployment{}, indexer.AtlasDeploymentBySpecNameAndProjectID, func(obj client.Object) []string {
					return []string{projectID + "-clusterA"}
				}).
				WithIndex(&akov2.AtlasDatabaseUser{}, indexer.AtlasDatabaseUserBySpecUsernameAndProjectID, func(obj client.Object) []string {
					return []string{projectID + "-admin"}
				}).
				Build()

			p := &ConnectionSecretContext{
				ProjectID:        projectID,
				ClusterName:      tt.clusterName,
				DatabaseUsername: tt.databaseUsername,
			}

			err := p.LoadDeploymentAndUser(context.Background(), client, ns)

			if tt.expectedErr == "" {
				assert.NoError(t, err)
				assert.NotNil(t, p.Deployment)
				assert.NotNil(t, p.User)
				assert.Equal(t, tt.clusterName, p.Deployment.GetDeploymentName())
				assert.Equal(t, tt.databaseUsername, p.User.Spec.Username)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			}

			failp := &ConnectionSecretContext{
				ProjectID:        otherprojectID,
				ClusterName:      tt.clusterName,
				DatabaseUsername: tt.databaseUsername,
			}

			failErr := failp.LoadDeploymentAndUser(context.Background(), client, ns)
			assert.Error(t, failErr)
			assert.Contains(t, failErr.Error(), fmt.Sprintf(`expected 1 AtlasDeployment for "%s-%s"`, otherprojectID, tt.clusterName))
		})
	}
}
