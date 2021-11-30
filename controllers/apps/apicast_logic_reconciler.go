package controllers

import (
	"context"
	"fmt"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	apicast "github.com/3scale/apicast-operator/pkg/apicast"
	"github.com/3scale/apicast-operator/pkg/k8sutils"
	"github.com/3scale/apicast-operator/pkg/reconcilers"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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

func (r *APIcastLogicReconciler) Reconcile(ctx context.Context) (reconcile.Result, error) {
	r.Logger().WithValues("Name", r.APIcastCR.Name, "Namespace", r.APIcastCR.Namespace)

	appliedInitialization, err := r.initialize(ctx)
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

	apicastFactory, err := apicast.Factory(ctx, r.APIcastCR, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	//
	// Admin Portal Credentials secret
	//
	adminPortalSecret := apicastFactory.AdminPortalCredentialsSecret()
	err = r.reconcileApicastSecret(ctx, adminPortalSecret)
	if err != nil {
		return reconcile.Result{}, err
	}

	//
	// Gateway configuration secret
	//
	confSecret := apicastFactory.GatewayConfigurationSecret()
	err = r.reconcileApicastSecret(ctx, confSecret)
	if err != nil {
		return reconcile.Result{}, err
	}

	//
	// Gateway deployment
	//
	deploymentMutator := reconcilers.DeploymentMutator(
		reconcilers.DeploymentReplicasMutator,
		reconcilers.DeploymentImageMutator,
		reconcilers.DeploymentServiceAccountNameMutator,
		reconcilers.DeploymentEnvVarsMutator,
		reconcilers.DeploymentResourceMutator,
		reconcilers.DeploymentPodTemplateAnnotationsMutator,
		reconcilers.DeploymentVolumesMutator,
		reconcilers.DeploymentVolumeMountsMutator,
		reconcilers.DeploymentPortsMutator,
		reconcilers.DeploymentTemplateLabelsMutator,
	)
	deployment := apicastFactory.Deployment()
	err = r.ReconcileResource(ctx, &appsv1.Deployment{}, deployment, deploymentMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	//
	// Gateway service
	//
	service := apicastFactory.Service()
	err = r.ReconcileResource(ctx, &v1.Service{}, service, reconcilers.ServicePortMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	//
	// Gateway ingress
	//
	ingress := apicastFactory.Ingress()
	err = r.reconcileIngress(ctx, ingress)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *APIcastLogicReconciler) reconcileApicastSecret(ctx context.Context, secret *v1.Secret) error {
	if secret == nil {
		return nil
	}

	return r.ReconcileResource(ctx, &v1.Secret{}, secret, r.ensureOwnerReferenceMutator)
}

func (r *APIcastLogicReconciler) initialize(ctx context.Context) (bool, error) {
	if appliedSomeInitialization := r.applyInitialization(); appliedSomeInitialization {
		r.Logger().Info(fmt.Sprintf("Updating %s", k8sutils.ObjectInfo(r.APIcastCR)))
		err := r.Client().Update(ctx, r.APIcastCR)
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

func (r *APIcastLogicReconciler) reconcileIngress(ctx context.Context, desired *networkingv1.Ingress) error {
	if r.APIcastCR.Spec.ExposedHost == nil {
		k8sutils.TagObjectToDelete(desired)
	}

	return r.ReconcileResource(ctx, &networkingv1.Ingress{}, desired, reconcilers.IngressMutator)
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
