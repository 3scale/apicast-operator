package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	apicast "github.com/3scale/apicast-operator/pkg/apicast"

	"github.com/3scale/apicast-operator/pkg/k8sutils"
	"github.com/3scale/apicast-operator/pkg/reconcilers"
	hpa "k8s.io/api/autoscaling/v2"
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
	logger, err := logr.FromContext(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	res, err := r.reconcileAPIcastCR(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if res.Requeue {
		return res, nil
	}

	err = r.validateAPicastCR(ctx)
	if err != nil {
		return reconcile.Result{}, err
	}

	apicastFactory, err := apicast.Factory(ctx, r.APIcastCR, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	upgradeDeploymentResult, err := r.upgradeDeploymentSelector(ctx, apicastFactory)
	if err != nil {
		return reconcile.Result{}, err
	}
	if upgradeDeploymentResult.Requeue {
		logger.Info("Upgrade in process. Requeueing request...")
		return upgradeDeploymentResult, nil
	}

	//
	// Gateway deployment
	//
	deploymentMutators := make([]reconcilers.DeploymentMutateFn, 0)
	if r.APIcastCR.Spec.Replicas != nil {
		deploymentMutators = append(deploymentMutators, reconcilers.DeploymentReplicasMutator)
	}

	deploymentMutators = append(deploymentMutators,
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
	err = r.ReconcileResource(ctx, &appsv1.Deployment{}, deployment, reconcilers.DeploymentMutator(deploymentMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	//
	// Gateway service
	//
	serviceMutators := []reconcilers.ServiceMutateFn{
		reconcilers.ServicePortMutator,
		reconcilers.ServiceSelectorMutator,
	}

	service := apicastFactory.Service()
	err = r.ReconcileResource(ctx, &v1.Service{}, service, reconcilers.ServiceMutator(serviceMutators...))
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

	// Hpa
	if r.APIcastCR.Spec.Hpa.Enabled {
		// If HPA is enabled and any of the fields are set, reconcile it
		if r.APIcastCR.Spec.Hpa.MinPods != nil || r.APIcastCR.Spec.Hpa.MaxPods != 0 || r.APIcastCR.Spec.Hpa.CpuPercent != nil || r.APIcastCR.Spec.Hpa.MemoryPercent != nil {
			err = r.ReconcileResource(ctx, &hpa.HorizontalPodAutoscaler{}, reconcilers.HpaCR(r.APIcastCR), reconcilers.HpaGenericMutator())
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		// If HPA is enabled but only the enable field is set, create the initial HPA object and do not reconcile further
		if r.APIcastCR.Spec.Hpa.MinPods == nil || r.APIcastCR.Spec.Hpa.MaxPods == 0 || r.APIcastCR.Spec.Hpa.CpuPercent == nil || r.APIcastCR.Spec.Hpa.MemoryPercent == nil {
			err = r.ReconcileResource(ctx, &hpa.HorizontalPodAutoscaler{}, reconcilers.HpaCR(r.APIcastCR), reconcilers.HpaCreateOnlyMutator())
			if err != nil {
				return reconcile.Result{}, err
			}

		}
	} else {
		// Check if HPA CR exists, if it does, delete it because HPA is set to false
		hpaDesired := reconcilers.HpaCR(r.APIcastCR)
		k8sutils.TagObjectToDelete(hpaDesired)
		err = r.ReconcileResource(ctx, &hpa.HorizontalPodAutoscaler{}, hpaDesired, reconcilers.HpaDeleteMutator())
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *APIcastLogicReconciler) reconcileAPIcastCR(ctx context.Context) (ctrl.Result, error) {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	changed := false

	tmpChanged := r.APIcastCR.UpdateOperatorVersion()
	changed = changed || tmpChanged

	tmpChanged, err = r.reconcileApicastSecretLabels(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	changed = changed || tmpChanged

	if changed {
		err = r.Client().Update(ctx, r.APIcastCR)
		logger.Info("reconciling", "error", err)
	}

	return ctrl.Result{Requeue: changed}, err
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
	// opentracing tracing config secret (deprecated)
	// opentelemetry tracing config secret

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

	if r.APIcastCR.OpenTelemetryEnabled() && r.APIcastCR.Spec.OpenTelemetry.TracingConfigSecretRef != nil {
		secretKeys = append(secretKeys, client.ObjectKey{
			Name:      r.APIcastCR.Spec.OpenTelemetry.TracingConfigSecretRef.Name,
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

func (r *APIcastLogicReconciler) validateAPicastCR(ctx context.Context) error {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		return err
	}

	if r.APIcastCR.OpenTracingIsEnabled() {
		logger.Info("[WARNING] opentracing use is DEPRECATED. Use Opentelemetry instead.")
	}

	errors := field.ErrorList{}
	// internal validation
	errors = append(errors, r.APIcastCR.Validate()...)

	if len(errors) == 0 {
		return nil
	}

	return errors.ToAggregate()
}
