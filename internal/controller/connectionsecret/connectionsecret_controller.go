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

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	akov2 "github.com/mongodb/mongodb-atlas-kubernetes/v2/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/controller/atlas"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/controller/reconciler"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/controller/workflow"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/indexer"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/pointer"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/stringutil"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/translation/project"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/ratelimit"
)

type ConnectionSecretReconciler struct {
	reconciler.AtlasReconciler
	GlobalPredicates []predicate.Predicate
	EventRecorder    record.EventRecorder
}

func (r *ConnectionSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Parses the request name and fills up the identifiers: ProjectID, ClusterName, DatabaseUsername
	strRequest := req.NamespacedName.String()
	ids, err := LoadRequestIdentifiers(ctx, r.Client, req.NamespacedName)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			r.Log.Warnf("ConnectionSecret not found; assuming it was deleted %s", strRequest)
			return workflow.TerminateSilently(err).WithoutRetry().ReconcileResult()
		}

		r.Log.Errorf("Failed to parse ConnectionSecret request with %s: %e", strRequest, err)
		return workflow.Terminate("InvalidConnectionSecretName", err).ReconcileResult()
	}

	// Loads the pair of AtlasDeployment and AtlasDatabaseUser via the indexers
	pair, err := LoadPairedResources(ctx, r.Client, ids, req.Namespace)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoDeploymentFound),
			errors.Is(err, ErrNoUserFound):

			r.Log.Info("Paired resource missing for ConnectionSecret request with %s — scheduling deletion", strRequest)
			r.handleDelete(ctx, req)
			return workflow.TerminateSilently(nil).WithoutRetry().ReconcileResult()

		case errors.Is(err, ErrManyDeployments),
			errors.Is(err, ErrManyUsers):
			r.Log.Errorf("Ambiguous pairing (more than one) for ConnectionSecret request with %s", strRequest)
			return workflow.Terminate("AmbiguousConnectionResources", err).ReconcileResult()

		default:
			r.Log.Errorf("Failed to get paired resources ConnectionSecret request with %s: %v", strRequest, err)
			return workflow.Terminate("InvalidConnectionResources", err).ReconcileResult()
		}
	}

	// If either AtlasDeployment or AtlasDatabaseUser have a deletion timestamp, delete the connection secret as well
	if pair.IsDeleting() {
		r.Log.Infof("Paired resource marked for deletion for ConnectionSecret with %s — scheduling deletion", strRequest)
		r.handleDelete(ctx, req)
		return workflow.TerminateSilently(nil).WithoutRetry().ReconcileResult()
	}

	// Checks that AtlasDeployment and AtlasDatabaseUser are ready before proceeding
	if ready, notReady := pair.IsReady(); !ready {
		r.Log.Infof("Waiting till paired resources are ready for ConnectionSecret request with %s", strRequest)
		return workflow.InProgress("ConnectionSecretNotReady", fmt.Sprintf("Not ready: %s", strings.Join(notReady, ", "))).ReconcileResult()
	}

	// ProjectName is required for ConnectionSecret metadata.name
	if ids.ProjectName == "" {
		// If both AtlasDeployment and AtlasDatabaseUser use externalRef, resolve via SDK
		if pair.NeedsSDKProjectResolution() {
			connectionConfig, err := r.ResolveConnectionConfig(ctx, pair.Deployment)
			if err != nil {
				r.Log.Errorf("Failed to resolve Atlas connection config for ConnectionSecret request with %s: %e", strRequest, err)
				return workflow.Terminate("FailedToResolveConnectionConfig", err).ReconcileResult()
			}

			sdkClientSet, err := r.AtlasProvider.SdkClientSet(ctx, connectionConfig.Credentials, r.Log)
			if err != nil {
				r.Log.Errorw("Failed to create SDK client for ConnectionSecret request with %s: %e", strRequest, err)
				return workflow.Terminate("FailedToCreateAtlasClient", err).ReconcileResult()
			}

			projectService := project.NewProjectAPIService(sdkClientSet.SdkClient20250312002.ProjectsApi)
			ids.ProjectName, err = pair.ResolveProjectNameSDK(ctx, projectService)
			if err != nil {
				r.Log.Errorw("Failed to resolve ProjectName for ConnectionSecret request with %s: %e", strRequest, err)
				return workflow.Terminate("FailedToFetchProjectFromAtlas", err).ReconcileResult()
			}
		} else { // Otherwise, fetch AtlasProject from K8s
			var resolveErr error
			ids.ProjectName, resolveErr = pair.ResolveProjectNameK8s(ctx, r.Client, req.Namespace)
			if resolveErr != nil {
				r.Log.Errorw("Failed to resolve ProjectName for ConnectionSecret request with %s: %e", strRequest, err)
				return workflow.Terminate("FailedToResolveProjectName", resolveErr).ReconcileResult()
			}
		}
	}

	// Compute the connection data
	data, err := pair.BuildConnectionData(ctx, r.Client)
	if err != nil {
		r.Log.Errorw("Failed to build connection data for ConnectionSecret request with %s: %e", strRequest, err)
		return workflow.Terminate("FailedToBuildConnectionData", err).ReconcileResult()
	}

	// Create or update the connection secret in K8s cluster
	if err := pair.HandleSecret(ctx, r.Client, ids, data); err != nil {
		r.Log.Errorw("Failed to create/update the ConnectionSecret for request with %s: %e", "error", err)
		return workflow.Terminate("FailedToEnsureConnectionSecret", err).ReconcileResult()
	}

	return workflow.OK().ReconcileResult()
}

func (r *ConnectionSecretReconciler) DeploymentWatcherPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool {
			d, ok := e.ObjectNew.(*akov2.AtlasDeployment)
			if !ok {
				return false
			}
			return IsDeploymentReady(d) && len(d.GetFinalizers()) > 0
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			d, ok := e.Object.(*akov2.AtlasDeployment)
			if !ok {
				return false
			}
			return len(d.GetFinalizers()) > 0
		},
	}
}

func (r *ConnectionSecretReconciler) DatabaseUserWatcherPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool {
			u, ok := e.ObjectNew.(*akov2.AtlasDatabaseUser)
			if !ok {
				return false
			}

			return IsDatabaseUserReady(u)
		},
	}
}

func (r *ConnectionSecretReconciler) For() (client.Object, builder.Predicates) {
	// Filter out connection secrets based on the required labels
	labelPredicates := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		labels := obj.GetLabels()
		_, hasType := labels[TypeLabelKey]
		_, hasProject := labels[ProjectLabelKey]
		_, hasCluster := labels[ClusterLabelKey]
		return hasType && hasProject && hasCluster
	})

	predicates := append(r.GlobalPredicates, labelPredicates)
	return &corev1.Secret{}, builder.WithPredicates(predicates...)
}

func (r *ConnectionSecretReconciler) SetupWithManager(mgr ctrl.Manager, skipNameValidation bool) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("ConnectionSecret").
		For(r.For()).
		Watches(
			&akov2.AtlasDeployment{},
			handler.EnqueueRequestsFromMapFunc(r.newDeploymentMapFunc),
			builder.WithPredicates(r.DeploymentWatcherPredicate()),
		).
		Watches(
			&akov2.AtlasDatabaseUser{},
			handler.EnqueueRequestsFromMapFunc(r.newDatabaseUserMapFunc),
			builder.WithPredicates(r.DatabaseUserWatcherPredicate()),
		).
		WithOptions(controller.TypedOptions[reconcile.Request]{
			RateLimiter:        ratelimit.NewRateLimiter[reconcile.Request](),
			SkipNameValidation: pointer.MakePtr(skipNameValidation),
		}).
		Complete(r)
}

func (r *ConnectionSecretReconciler) generateConnectionSecretRequests(
	projectID string,
	deployments []akov2.AtlasDeployment,
	users []akov2.AtlasDatabaseUser,
) []reconcile.Request {
	var requests []reconcile.Request
	for _, d := range deployments {
		for _, u := range users {
			scopes := u.GetScopes(akov2.DeploymentScopeType)
			if len(scopes) != 0 && !stringutil.Contains(scopes, d.GetDeploymentName()) {
				continue
			}

			requestName := CreateInternalFormat(projectID, d.GetDeploymentName(), u.Spec.Username)
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: u.Namespace,
					Name:      requestName,
				},
			})
		}
	}
	return requests
}

func (r *ConnectionSecretReconciler) newDeploymentMapFunc(ctx context.Context, obj client.Object) []reconcile.Request {
	deployment, ok := obj.(*akov2.AtlasDeployment)
	if !ok {
		r.Log.Warnf("watching AtlasDeployment but got %T", obj)
		return nil
	}

	projectID, err := ResolveProjectIDFromDeployment(ctx, r.Client, deployment)
	if err != nil {
		r.Log.Errorw("Unable to resolve projectID for deployment", "error", err)
		return nil
	}

	users := &akov2.AtlasDatabaseUserList{}
	if err := r.Client.List(ctx, users, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(indexer.AtlasDatabaseUserByProject, projectID),
	}); err != nil {
		r.Log.Errorf("failed to list AtlasDatabaseUsers: %e", err)
		return nil
	}

	return r.generateConnectionSecretRequests(projectID, []akov2.AtlasDeployment{*deployment}, users.Items)
}

func (r *ConnectionSecretReconciler) newDatabaseUserMapFunc(ctx context.Context, obj client.Object) []reconcile.Request {
	user, ok := obj.(*akov2.AtlasDatabaseUser)
	if !ok {
		r.Log.Warnf("watching AtlasDatabaseUser but got %T", obj)
		return nil
	}

	projectID, err := ResolveProjectIDFromDatabaseUser(ctx, r.Client, user)
	if err != nil {
		r.Log.Errorw("Unable to resolve projectID for user", "error", err)
		return nil
	}

	deployments := &akov2.AtlasDeploymentList{}
	if err := r.Client.List(ctx, deployments, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(indexer.AtlasDeploymentByProject, projectID),
	}); err != nil {
		r.Log.Errorf("failed to list AtlasDeployments: %e", err)
		return nil
	}

	return r.generateConnectionSecretRequests(projectID, deployments.Items, []akov2.AtlasDatabaseUser{*user})
}

func NewConnectionSecretReconciler(
	c cluster.Cluster,
	predicates []predicate.Predicate,
	atlasProvider atlas.Provider,
	logger *zap.Logger,
	globalSecretRef types.NamespacedName,
) *ConnectionSecretReconciler {
	return &ConnectionSecretReconciler{
		AtlasReconciler: reconciler.AtlasReconciler{
			Client:          c.GetClient(),
			Log:             logger.Named("controllers").Named("ConnectionSecret").Sugar(),
			GlobalSecretRef: globalSecretRef,
			AtlasProvider:   atlasProvider,
		},
		EventRecorder:    c.GetEventRecorderFor("ConnectionSecret"),
		GlobalPredicates: predicates,
	}
}
