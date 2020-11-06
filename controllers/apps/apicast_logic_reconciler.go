package controllers

import (
	"context"
	"fmt"
	"reflect"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	apicast "github.com/3scale/apicast-operator/pkg/apicast"
	"github.com/3scale/apicast-operator/pkg/k8sutils"
	"github.com/3scale/apicast-operator/pkg/reconcilers"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type APIcastLogicReconciler struct {
	reconcilers.BaseReconciler
	APIcastCR *appsv1alpha1.APIcast
}

func NewAPIcastLogicReconciler(b reconcilers.BaseReconciler, cr *appsv1alpha1.APIcast) APIcastLogicReconciler {
	return APIcastLogicReconciler{
		BaseReconciler: b,
		APIcastCR:      cr,
	}
}

func (r *APIcastLogicReconciler) Reconcile() (reconcile.Result, error) {
	r.Logger().WithValues("Name", r.APIcastCR.Name, "Namespace", r.APIcastCR.Namespace)

	appliedInitialization, err := r.initialize()
	if err != nil {
		return reconcile.Result{}, err
	}
	if appliedInitialization {
		// Stop the reconciliation cycle and order requeue to stop processing
		// of reconciliation
		return reconcile.Result{Requeue: true}, nil
	}

	err = r.validateAPicastCR()
	if err != nil {
		return reconcile.Result{}, err
	}

	apicastFactory, err := apicast.Factory(r.APIcastCR, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	//
	// Admin Portal Credentials secret
	//
	adminPortalSecret := apicastFactory.AdminPortalCredentialsSecret()
	err = r.reconcileApicastSecret(adminPortalSecret)
	if err != nil {
		return reconcile.Result{}, err
	}

	//
	// Gateway configuration secret
	//
	confSecret := apicastFactory.GatewayConfigurationSecret()
	err = r.reconcileApicastSecret(confSecret)
	if err != nil {
		return reconcile.Result{}, err
	}

	//
	// Gateway deployment
	//
	deployment := apicastFactory.Deployment()
	err = r.ReconcileResource(&appsv1.Deployment{}, deployment, DeploymentMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	//
	// Gateway service
	//
	service := apicastFactory.Service()
	err = r.ReconcileResource(&v1.Service{}, service, reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	//
	// Gateway ingress
	//
	ingress := apicastFactory.Ingress()
	err = r.reconcileIngress(ingress)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *APIcastLogicReconciler) reconcileApicastSecret(secret *v1.Secret) error {
	if secret == nil {
		return nil
	}

	return r.ReconcileResource(&v1.Secret{}, secret, r.ensureOwnerReferenceMutator)
}

func (r *APIcastLogicReconciler) initialize() (bool, error) {
	if appliedSomeInitialization := r.applyInitialization(); appliedSomeInitialization {
		r.Logger().Info(fmt.Sprintf("Updating %s", k8sutils.ObjectInfo(r.APIcastCR)))
		err := r.Client().Update(context.TODO(), r.APIcastCR)
		if err != nil {
			return false, err
		}
		r.Logger().Info("APIcast resource missed optional fields. Updated CR which triggered a new reconciliation event")
		return true, nil
	}
	return false, nil
}

func (r *APIcastLogicReconciler) applyInitialization() bool {
	var defaultAPIcastReplicas int64 = 1
	appliedInitialization := false

	if r.APIcastCR.Spec.Replicas == nil {
		r.APIcastCR.Spec.Replicas = &defaultAPIcastReplicas
		appliedInitialization = true
	}

	return appliedInitialization
}

func (r *APIcastLogicReconciler) reconcileIngress(desired *extensions.Ingress) error {
	if r.APIcastCR.Spec.ExposedHost == nil {
		k8sutils.TagObjectToDelete(desired)
	}

	return r.ReconcileResource(&extensions.Ingress{}, desired, IngressMutator)
}

func (r *APIcastLogicReconciler) ensureOwnerReferenceMutator(existing, desired k8sutils.KubernetesObject) (bool, error) {
	changed := false

	originalSize := len(existing.GetOwnerReferences())

	if err := controllerutil.SetControllerReference(r.APIcastCR, existing, r.Scheme()); err != nil {
		return false, err
	}

	newSize := len(existing.GetOwnerReferences())
	if originalSize != newSize {
		changed = true
	}

	return changed, nil
}

func (r *APIcastLogicReconciler) validateAPicastCR() error {
	errors := field.ErrorList{}
	// internal validation
	errors = append(errors, r.APIcastCR.Validate()...)

	if len(errors) == 0 {
		return nil
	}

	return errors.ToAggregate()
}

func DeploymentResourcesReconciler(desired, existing *appsv1.Deployment) bool {
	desiredName := k8sutils.ObjectInfo(desired)
	update := false

	//
	// Check container resource requirements
	//
	if len(desired.Spec.Template.Spec.Containers) != 1 {
		panic(fmt.Sprintf("%s desired spec.template.spec.containers length changed to '%d', should be 1", desiredName, len(desired.Spec.Template.Spec.Containers)))
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		existing.Spec.Template.Spec.Containers = desired.Spec.Template.Spec.Containers
		update = true
	}

	if !k8sutils.CmpResources(&existing.Spec.Template.Spec.Containers[0].Resources, &desired.Spec.Template.Spec.Containers[0].Resources) {
		existing.Spec.Template.Spec.Containers[0].Resources = desired.Spec.Template.Spec.Containers[0].Resources
		update = true
	}

	return update
}

func DeploymentMutator(existingObj, desiredObj k8sutils.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*appsv1.Deployment)
	if !ok {
		return false, fmt.Errorf("%T is not a *appsv1.Deployment", existingObj)
	}
	desired, ok := desiredObj.(*appsv1.Deployment)
	if !ok {
		return false, fmt.Errorf("%T is not a *appsv1.Deployment", desiredObj)
	}

	changed := false

	if existing.Spec.Replicas != desired.Spec.Replicas {
		existing.Spec.Replicas = desired.Spec.Replicas
		changed = true
	}
	if existing.Spec.Template.Spec.Containers[0].Image != desired.Spec.Template.Spec.Containers[0].Image {
		existing.Spec.Template.Spec.Containers[0].Image = desired.Spec.Template.Spec.Containers[0].Image
		changed = true

	}
	if existing.Spec.Template.Spec.ServiceAccountName != desired.Spec.Template.Spec.ServiceAccountName {
		changed = true
		existing.Spec.Template.Spec.ServiceAccountName = desired.Spec.Template.Spec.ServiceAccountName
	}

	updatedTmp := reconcilers.ReconcileEnvVar(&existing.Spec.Template.Spec.Containers[0].Env, desired.Spec.Template.Spec.Containers[0].Env)
	changed = changed || updatedTmp

	updatedTmp = DeploymentResourcesReconciler(desired, existing)
	changed = changed || updatedTmp

	// They are annotations of the PodTemplate, part of the Spec, not part of the meta info of the Pod or Environment object itself
	// It is not expected any controller to update them, so we use "set" approach, instead of merge.
	// This way any removed annotation from desired (due to change in CR) will be removed in existing too.
	if !reflect.DeepEqual(existing.Spec.Template.Annotations, desired.Spec.Template.Annotations) {
		changed = true
		existing.Spec.Template.Annotations = desired.Spec.Template.Annotations
	}

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Volumes, desired.Spec.Template.Spec.Volumes) {
		changed = true
		existing.Spec.Template.Spec.Volumes = desired.Spec.Template.Spec.Volumes
	}

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[0].VolumeMounts, desired.Spec.Template.Spec.Containers[0].VolumeMounts) {
		changed = true
		existing.Spec.Template.Spec.Containers[0].VolumeMounts = desired.Spec.Template.Spec.Containers[0].VolumeMounts
	}

	return changed, nil
}

func IngressMutator(existingObj, desiredObj k8sutils.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*extensions.Ingress)
	if !ok {
		return false, fmt.Errorf("%T is not a *extensions.Ingress", existingObj)
	}
	desired, ok := desiredObj.(*extensions.Ingress)
	if !ok {
		return false, fmt.Errorf("%T is not a *extensions.Ingress", desiredObj)
	}

	exposedHostIdx := -1
	exposedHost := desired.Spec.Rules[0].Host
	for idx, rule := range existing.Spec.Rules {
		if rule.Host == exposedHost {
			exposedHostIdx = idx
		}
	}

	update := false

	if exposedHostIdx == -1 {
		existing.Spec.Rules = desired.Spec.Rules
		update = true
	}

	if !reflect.DeepEqual(existing.Spec.TLS, desired.Spec.TLS) {
		existing.Spec.TLS = desired.Spec.TLS
		update = true
	}

	return update, nil
}
