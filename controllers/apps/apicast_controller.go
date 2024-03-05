/*
Copyright 2020 Red Hat.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimachinerymetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	"github.com/3scale/apicast-operator/pkg/reconcilers"
)

// APIcastReconciler reconciles a APIcast object
type APIcastReconciler struct {
	reconcilers.BaseControllerReconciler
	Log                 logr.Logger
	SecretLabelSelector apimachinerymetav1.LabelSelector
	WatchedNamespace    string
}

// blank assignment to verify that ReconcileAPIcast implements reconcile.Reconciler
var _ reconcile.Reconciler = &APIcastReconciler{}

// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apicasts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apicasts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apicasts/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,namespace=placeholder,resources=pods;services;services/finalizers;endpoints;persistentvolumeclaims;events;configmaps;secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,namespace=placeholder,resources=deployments;daemonsets;replicasets;statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,namespace=placeholder,resources=servicemonitors,verbs=get;create
// TODO the permission to update deployments/finalizer originally was limited
// to the 'apicast-operator' resource name. It seems it is not possible anymore
// with kubebuilder markers???
// +kubebuilder:rbac:groups=apps,namespace=placeholder,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=networking.k8s.io,namespace=placeholder,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,namespace=placeholder,resources=routes/custom-host,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,namespace=placeholder,resources=horizontalpodautoscalers,verbs=create;update;delete;get;list;watch

func (r *APIcastReconciler) Reconcile(eventCtx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("apicast", req.NamespacedName)
	ctx := logr.NewContext(eventCtx, log)

	// your logic here
	log.Info("Reconciling APIcast")

	instance := &appsv1alpha1.APIcast{}
	err := r.Client().Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("APIcast not found")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Error getting APIcast")
		return ctrl.Result{}, err
	}

	if log.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(instance, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		log.V(1).Info(string(jsonData))
	}

	baseReconciler := reconcilers.NewBaseReconciler(r.Client(), r.APIClientReader(), r.Scheme(), log)
	logicReconciler := NewAPIcastLogicReconciler(baseReconciler, instance)
	specResult, specErr := logicReconciler.Reconcile(ctx)
	if specErr == nil && specResult.Requeue {
		log.V(1).Info("Reconciling spec not finished. Requeueing.")
		return specResult, nil
	}

	statusResult, statusErr := r.reconcileStatus(ctx, instance, specErr)

	if specErr != nil {
		// Ignore conflicts, resource might just be outdated.
		if errors.IsConflict(specErr) {
			log.Info("Resource update conflict error. Requeuing...", "error", specErr)
			return ctrl.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, specErr
	}

	if statusErr != nil {
		return ctrl.Result{}, statusErr
	}

	if statusResult.Requeue {
		log.V(1).Info("Reconciling status not finished. Requeueing.")
		return statusResult, nil
	}

	log.Info("Successfully reconciled")
	return ctrl.Result{}, nil
}

func (r *APIcastReconciler) SetupWithManager(mgr ctrl.Manager) error {
	secretToApicastEventMapper := &SecretToApicastEventMapper{
		K8sClient: r.Client(),
		Logger:    r.Log.WithName("secretToApicastEventMapper"),
		Namespace: r.WatchedNamespace,
	}

	// LabelSelectorPredicate only applies to the new object in update events
	// Thus, if the secret label is removed, reconciliation capability will be lost
	// Thus, if the secret label is updated (no longer matching), reconciliation capability will be lost
	// If the controller would want to react when the label is removed or updated, a custom predicate
	// would be needed.
	labelSelectorPredicate, err := predicate.LabelSelectorPredicate(r.SecretLabelSelector)
	if err != nil {
		return nil
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.APIcast{}).
		Watches(
			&source.Kind{Type: &v1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(secretToApicastEventMapper.Map),
			builder.WithPredicates(labelSelectorPredicate),
		).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
