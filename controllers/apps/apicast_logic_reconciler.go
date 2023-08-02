package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	apicast "github.com/3scale/apicast-operator/pkg/apicast"
	"github.com/3scale/apicast-operator/pkg/k8sutils"
	"github.com/3scale/apicast-operator/pkg/reconcilers"
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

func (r *APIcastLogicReconciler) initialize(ctx context.Context) (bool, error) {
	appliedSomeInitialization, err := r.applyInitialization(ctx)
	if err != nil {
		return false, err
	}

	if appliedSomeInitialization {
		r.Logger().Info(fmt.Sprintf("Updating %s", k8sutils.ObjectInfo(r.APIcastCR)))
		err := r.Client().Update(ctx, r.APIcastCR)
		if err != nil {
			return false, err
		}
		r.Logger().Info("APIcast resource missed some fields. Updated CR which triggered a new reconciliation event")
		return true, nil
	}
	return false, nil
}

func (r *APIcastLogicReconciler) applyInitialization(ctx context.Context) (bool, error) {
	var defaultAPIcastReplicas int64 = 1
	appliedInitialization := false

	if r.APIcastCR.Spec.Replicas == nil {
		r.APIcastCR.Spec.Replicas = &defaultAPIcastReplicas
		appliedInitialization = true
	}

	changed, err := r.reconcileApicastSecretLabels(ctx)
	if err != nil {
		return false, err
	}
	appliedInitialization = appliedInitialization || changed

	return appliedInitialization, nil
}

func (r *APIcastLogicReconciler) reconcileApicastSecretLabels(ctx context.Context) (bool, error) {
	secretUIDs, err := r.getSecretUIDs(ctx)
	if err != nil {
		return false, err
	}

	return replaceAPIcastSecretLabels(r.APIcastCR, secretUIDs), nil
}

func (r *APIcastLogicReconciler) getSecretUIDs(ctx context.Context) ([]string, error) {
	// https certificate secret
	// admin portal secret
	// gateway conf secret
	// custom policy secret(s)
	// custom env secret(s)
	// tracing config secret

	secretKeys := []client.ObjectKey{}
	if r.APIcastCR.Spec.HTTPSCertificateSecretRef != nil {
		secretKeys = append(secretKeys, client.ObjectKey{
			Name:      r.APIcastCR.Spec.HTTPSCertificateSecretRef.Name,
			Namespace: r.APIcastCR.Namespace, // review when operator is also cluster scoped
		})
	}
	if r.APIcastCR.Spec.AdminPortalCredentialsRef != nil {
		secretKeys = append(secretKeys, client.ObjectKey{
			Name:      r.APIcastCR.Spec.AdminPortalCredentialsRef.Name,
			Namespace: r.APIcastCR.Namespace, // review when operator is also cluster scoped
		})
	}
	if r.APIcastCR.Spec.EmbeddedConfigurationSecretRef != nil {
		secretKeys = append(secretKeys, client.ObjectKey{
			Name:      r.APIcastCR.Spec.EmbeddedConfigurationSecretRef.Name,
			Namespace: r.APIcastCR.Namespace, // review when operator is also cluster scoped
		})
	}

	for idx := range r.APIcastCR.Spec.CustomPolicies {
		secretKeys = append(secretKeys, client.ObjectKey{
			Name:      r.APIcastCR.Spec.CustomPolicies[idx].SecretRef.Name, // CR validation ensures not nil
			Namespace: r.APIcastCR.Namespace,                               // review when operator is also cluster scoped
		})
	}

	for idx := range r.APIcastCR.Spec.CustomEnvironments {
		secretKeys = append(secretKeys, client.ObjectKey{
			Name:      r.APIcastCR.Spec.CustomEnvironments[idx].SecretRef.Name, // CR validation ensures not nil
			Namespace: r.APIcastCR.Namespace,                                   // review when operator is also cluster scoped
		})
	}

	if r.APIcastCR.OpenTracingIsEnabled() && r.APIcastCR.Spec.OpenTracing.TracingConfigSecretRef != nil {
		secretKeys = append(secretKeys, client.ObjectKey{
			Name:      r.APIcastCR.Spec.OpenTracing.TracingConfigSecretRef.Name,
			Namespace: r.APIcastCR.Namespace, // review when operator is also cluster scoped
		})
	}

	uids := []string{}
	for idx := range secretKeys {
		secret := &v1.Secret{}
		secretKey := secretKeys[idx]
		err := r.Client().Get(ctx, secretKey, secret)
		r.Logger().V(1).Info("read secret", "objectKey", secretKey, "error", err)
		if err != nil {
			return nil, err
		}
		uids = append(uids, string(secret.GetUID()))
	}

	return uids, nil
}

func (r *APIcastLogicReconciler) reconcileIngress(ctx context.Context, desired *networkingv1.Ingress) error {
	if r.APIcastCR.Spec.ExposedHost == nil {
		k8sutils.TagObjectToDelete(desired)
	}

	return r.ReconcileResource(ctx, &networkingv1.Ingress{}, desired, reconcilers.IngressMutator)
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
