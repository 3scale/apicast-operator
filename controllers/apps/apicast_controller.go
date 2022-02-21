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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	"github.com/3scale/apicast-operator/pkg/apicast"
	"github.com/3scale/apicast-operator/pkg/reconcilers"
	"github.com/3scale/apicast-operator/version"
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

const (
	APIcastOperatorVersionAnnotation = "apicast.apps.3scale.net/operator-version"
)

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

func (r *APIcastReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("apicast", req.NamespacedName)

	// your logic here
	log.Info("Reconciling APIcast")

	instance, err := r.getAPIcast(ctx, req)
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

	if instance.ObjectMeta.Annotations == nil || instance.ObjectMeta.Annotations[APIcastOperatorVersionAnnotation] == "" {
		log.Info("APIcast operator version not set in annotations. Setting it...")
		if instance.ObjectMeta.Annotations == nil {
			instance.ObjectMeta.Annotations = map[string]string{}
		}
		err = r.updateAPIcastOperatorVersionInAnnotations(ctx, instance, log)
		if err != nil {
			return ctrl.Result{}, err
		}
		log.Info("APIcast operator version in annotations set. Requeuing request...")
		return ctrl.Result{Requeue: true}, err
	}

	if instance.ObjectMeta.Annotations[APIcastOperatorVersionAnnotation] != version.Version {
		log.Info("APIcast operator version in annotations does not match expected version. Applying upgrade procedure...")
		upgradeReconcileResult, err := r.upgradeAPIcast(ctx, instance, log)
		if err != nil {
			log.Error(err, "Error upgrading APIcast")
			return ctrl.Result{}, err
		}
		if upgradeReconcileResult.Requeue {
			return upgradeReconcileResult, nil
		}
		log.Info("APIcast upgrade procedure applied")
		log.Info("Setting APIcast operator version in annotations...")
		err = r.updateAPIcastOperatorVersionInAnnotations(ctx, instance, log)
		if err != nil {
			return ctrl.Result{}, err
		}
		log.Info("APIcast operator version in annotations set. Requeuing request...")
		return ctrl.Result{Requeue: true}, nil
	}

	baseReconciler := reconcilers.NewBaseReconciler(r.Client(), r.APIClientReader(), r.Scheme(), log)
	logicReconciler := NewAPIcastLogicReconciler(baseReconciler, instance)
	result, err := logicReconciler.Reconcile(ctx)
	if err != nil {
		// Ignore conflicts, resource might just be outdated.
		if errors.IsConflict(err) {
			log.Info("Resource update conflict error. Requeuing...", "error", err)
			return ctrl.Result{Requeue: true}, nil
		}

		return result, err
	}
	if result.Requeue {
		log.Info("Requeuing request...")
		return result, nil
	}
	log.Info("APIcast logic reconciled")

	result, err = r.updateStatus(ctx, instance, &logicReconciler)
	if err != nil {
		// Ignore conflicts, resource might just be outdated.
		if errors.IsConflict(err) {
			log.Info("Resource update conflict error. Requeuing...", "error", err)
			return ctrl.Result{Requeue: true}, nil
		}

		log.Error(err, "Status reconciler")
		return result, err
	}
	if result.Requeue {
		log.Info("Requeuing request...")
		return result, nil
	}
	log.Info("APIcast status reconciled")
	return ctrl.Result{}, nil
}

func (r *APIcastReconciler) SetupWithManager(mgr ctrl.Manager) error {
	secretToApicastEventMapper := &SecretToApicastEventMapper{
		K8sClient: r.Client(),
		Logger:    r.Log.WithName("secretToApicastEventMapper"),
		Namespace: r.WatchedNamespace,
	}

	// LabelSelectorPredicate only applies to the new object in update events
	// Thus, if the kuadrant secret label is removed, reconciliation capability will be lost
	// Thus, if the kuadrant secret label is updated (no longer matching), reconciliation capability will be lost
	// If the controller would want to react when the label is removed or updated, a custom predicate
	// would be needed. Like it is implemented in the service controller of the kuadrant controller
	// https://github.com/Kuadrant/kuadrant-controller/blob/356fb4d7abce66ef2ad5d93bad6461ee6c254e02/controllers/service_controller.go#L349
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

func (r *APIcastReconciler) updateStatus(ctx context.Context, instance *appsv1alpha1.APIcast, reconciler *APIcastLogicReconciler) (ctrl.Result, error) {
	apicastFactory, err := apicast.Factory(ctx, instance, r.Client())
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredDeployment := apicastFactory.Deployment()
	apicastDeployment := &appsv1.Deployment{}
	err = r.Client().Get(ctx, types.NamespacedName{Name: desiredDeployment.Name, Namespace: desiredDeployment.Namespace}, apicastDeployment)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if err != nil && errors.IsNotFound(err) {
		return ctrl.Result{Requeue: true}, nil
	}

	deployedImage := apicastDeployment.Spec.Template.Spec.Containers[0].Image
	if instance.Status.Image != deployedImage {
		instance.Status.Image = deployedImage
		err = r.Client().Status().Update(ctx, instance)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func (r *APIcastReconciler) upgradeAPIcast(ctx context.Context, apicastCR *appsv1alpha1.APIcast, logger logr.Logger) (ctrl.Result, error) {
	err := r.removeApicastSecretOwnership(ctx, apicastCR, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *APIcastReconciler) getAPIcast(ctx context.Context, request ctrl.Request) (*appsv1alpha1.APIcast, error) {
	instance := appsv1alpha1.APIcast{}
	err := r.Client().Get(ctx, request.NamespacedName, &instance)
	return &instance, err
}

func (r *APIcastReconciler) updateAPIcastOperatorVersionInAnnotations(ctx context.Context, instance *appsv1alpha1.APIcast, logger logr.Logger) error {
	instance.Annotations[APIcastOperatorVersionAnnotation] = version.Version
	err := r.Client().Update(ctx, instance)
	if err != nil {
		logger.Error(err, "Error setting APIcast operator version in annotations")
	}
	return err
}

func (r *APIcastReconciler) removeApicastSecretOwnership(ctx context.Context, apicastCR *appsv1alpha1.APIcast, logger logr.Logger) error {
	// ownership only for:
	// admin portal secret
	// gateway conf secret

	secretKeys := []client.ObjectKey{}
	if apicastCR.Spec.AdminPortalCredentialsRef != nil {
		secretKeys = append(secretKeys, client.ObjectKey{
			Name:      apicastCR.Spec.AdminPortalCredentialsRef.Name,
			Namespace: apicastCR.Namespace, // review when operator is also cluster scoped
		})
	}

	if apicastCR.Spec.EmbeddedConfigurationSecretRef != nil {
		secretKeys = append(secretKeys, client.ObjectKey{
			Name:      apicastCR.Spec.EmbeddedConfigurationSecretRef.Name,
			Namespace: apicastCR.Namespace, // review when operator is also cluster scoped
		})
	}

	for idx := range secretKeys {
		secret := &v1.Secret{}
		secretKey := secretKeys[idx]
		err := r.Client().Get(ctx, secretKey, secret)
		logger.V(1).Info("read secret", "objectKey", secretKey, "error", err)
		if err != nil {
			return err
		}
		updated := removeSecretOwnership(secret, apicastCR)
		if updated {
			err = r.Client().Update(ctx, secret)
			logger.V(1).Info("remove secret ownership", "objectKey", secretKey, "error", err)
			if err != nil {
				logger.Error(err, "Error setting APIcast operator version in annotations")
				return err
			}
		}
	}
	return nil
}

func removeSecretOwnership(secret *v1.Secret, apicastCR *appsv1alpha1.APIcast) bool {
	changed := false

	originalSize := len(secret.GetOwnerReferences())
	apicastOwnerRef := apicastCR.GetOwnerRefence()
	newOwners := []apimachinerymetav1.OwnerReference{}
	for idx := range secret.GetOwnerReferences() {

		aGV, err := schema.ParseGroupVersion(apicastOwnerRef.APIVersion)
		if err != nil {
			continue
		}

		bGV, err := schema.ParseGroupVersion(secret.GetOwnerReferences()[idx].APIVersion)
		if err != nil {
			continue
		}

		if aGV.Group != bGV.Group || apicastOwnerRef.Kind != secret.GetOwnerReferences()[idx].Kind || apicastOwnerRef.Name != secret.GetOwnerReferences()[idx].Name {
			newOwners = append(newOwners, secret.GetOwnerReferences()[idx])
		}
	}

	if originalSize != len(newOwners) {
		secret.SetOwnerReferences(newOwners)
	}

	newSize := len(secret.GetOwnerReferences())
	if originalSize != newSize {
		changed = true
	}

	return changed
}
