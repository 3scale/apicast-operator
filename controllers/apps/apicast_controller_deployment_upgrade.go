package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	apicast "github.com/3scale/apicast-operator/pkg/apicast"
	"github.com/3scale/apicast-operator/pkg/k8sutils"
)

// this name cannot be another apicast deployment which has the pattern "apicast-<CR_NAME>"
const TMP_DEPLOYMENT_NAME = "tmp-upgrade-apicast"

func (r *APIcastLogicReconciler) upgradeDeploymentSelector(ctx context.Context, apicastFactory *apicast.APIcast) (ctrl.Result, error) {
	// Some previous released apicast operator added labels with release version in the
	// deployment selector, which happens to be immutable
	// https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#label-selector-updates
	//
	// This upgrading specific procedure changes the inmutable deployment selector.
	// Since the deployment selector is immutable, the deployment needs to be deleted first.
	// To provide zero downtime(tm), the high level workflow will be as follow:
	// 1) Existing service will be updated. The selector will be simplified to select pods from the old deployment, temporary deployment and the new deployment
	// 2) Create a new temporary deployment to deploy new pods
	// 3) Old ("corrupted") deployment is deleted. Service is working with the temporary deployment.
	// 4) New deployment is created with the fixed deployment selector.
	// 5) Delete temporary deployment

	// To understand the code, a diagram with the implemented workflow has been generated
	// in apicast_controller_deployment_upgrade.md

	logger, err := logr.FromContext(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	expectedDeployment := apicastFactory.Deployment()
	existingDeployment := &appsv1.Deployment{}
	err = r.Client().Get(ctx, client.ObjectKeyFromObject(expectedDeployment), existingDeployment)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	deploymentExists := !errors.IsNotFound(err)

	err = r.Client().Get(ctx, client.ObjectKey{Name: TMP_DEPLOYMENT_NAME, Namespace: expectedDeployment.Namespace}, &appsv1.Deployment{})
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	tempDeploymentExists := !errors.IsNotFound(err)

	if !deploymentExists {
		if !tempDeploymentExists {
			// Nothing to do
			return ctrl.Result{}, nil
		} else {
			// temp deployment exists, workload was not interrupted, create upgraded deployment
			err = r.Client().Create(ctx, expectedDeployment)
			logger.Info("Upgrade deployment: creating new deployment", "key", client.ObjectKeyFromObject(expectedDeployment), "error", err)
			return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, err
		}
	} else {
		// deployment exists. Temp deployment may or may not exist
		// Check the deployment is "corrupted"
		if existingDeployment.Spec.Selector == nil || existingDeployment.Spec.Selector.MatchLabels == nil {
			// nothing to do
			return ctrl.Result{}, nil
		}

		if _, ok := existingDeployment.Spec.Selector.MatchLabels["rht.comp_ver"]; !ok {
			return r.upgradedDeploymentWorkflow(ctx, existingDeployment)
		}

		return r.notUpgradedDeploymentWorkflow(ctx, apicastFactory, existingDeployment)
	}
}

// Check temporary deployment
// if temp deployment does not exist, nothing to be done.
// if temp deployment exists, then
//
//	Check deployment whether status is available or not
//	if available, delete temp deployment
//	if not available, requeue
func (r *APIcastLogicReconciler) upgradedDeploymentWorkflow(ctx context.Context, existingDeployment *appsv1.Deployment) (ctrl.Result, error) {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	tempExistingDeployment := &appsv1.Deployment{}
	err = r.Client().Get(ctx, client.ObjectKey{Name: TMP_DEPLOYMENT_NAME, Namespace: existingDeployment.Namespace}, tempExistingDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			// Nothing to be done
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// check existing deployment (not temp) is ready
	if !k8sutils.IsStatusConditionTrue(existingDeployment.Status.Conditions, appsv1.DeploymentAvailable) {
		logger.Info("Upgrade deployment: new deployment not available")
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	_ = r.Client().Delete(ctx, tempExistingDeployment)
	logger.Info("Upgrade deployment: delete temp deployment")
	return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
}

func (r *APIcastLogicReconciler) reconcileUpgradeService(ctx context.Context, apicastFactory *apicast.APIcast, existingDeployment *appsv1.Deployment) error {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		return err
	}

	expectedService := apicastFactory.Service()
	existingService := &v1.Service{}
	err = r.Client().Get(ctx, client.ObjectKeyFromObject(expectedService), existingService)
	if err != nil {
		if errors.IsNotFound(err) {
			// Not found, nothing to upgrade
			return nil
		}
		return err
	}

	if _, ok := existingService.Spec.Selector["deployment"]; !ok {
		logger.Info("[ERROR] apicast service does not have required 'deployment' label.")
		return nil
	}

	if len(existingService.Spec.Selector) != 1 {
		existingService.Spec.Selector = map[string]string{
			"deployment": existingService.Spec.Selector["deployment"],
		}
		err = r.Client().Update(ctx, existingService)
		logger.Info("Upgrade deployment: updating service", "key", client.ObjectKeyFromObject(existingService), "error", err)
		return err
	}

	return nil
}

// The workflow:
//   - Existing service will be updated. The selector will be simplified to select pods from the old deployment, temporary deployment and the new deployment
//   - Check temp deployment
//     if temp deployment does not exist, create and requeue
//     if temp deployment exists, delete the apicast deployment
func (r *APIcastLogicReconciler) notUpgradedDeploymentWorkflow(ctx context.Context, apicastFactory *apicast.APIcast, existingDeployment *appsv1.Deployment) (ctrl.Result, error) {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.reconcileUpgradeService(ctx, apicastFactory, existingDeployment)
	if err != nil {
		return ctrl.Result{}, err
	}

	tempExistingDeployment := &appsv1.Deployment{}
	err = r.Client().Get(ctx, client.ObjectKey{Name: TMP_DEPLOYMENT_NAME, Namespace: existingDeployment.Namespace}, tempExistingDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			tempDesiredDeployment := newTempDeployment(existingDeployment, r.APIcastCR)
			err = r.Client().Create(ctx, tempDesiredDeployment)
			logger.Info("Upgrade deployment: creating tmp deployment", "key", client.ObjectKeyFromObject(tempDesiredDeployment), "error", err)
			return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, err
		}
		return ctrl.Result{}, err
	}

	// check temp deployment is ready
	if !k8sutils.IsStatusConditionTrue(tempExistingDeployment.Status.Conditions, appsv1.DeploymentAvailable) {
		logger.Info("Upgrade deployment: tmp deployment not available")
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	// delete old deployment
	// if error is returned, it needs to be raised. Deletion must succeed to accomplish the upgrade
	err = r.Client().Delete(ctx, existingDeployment)
	logger.Info("Upgrade deployment: delete old deployment", "key", client.ObjectKeyFromObject(existingDeployment), "error", err)
	return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, err
}

func newTempDeployment(oldDeployment *appsv1.Deployment, apicastCR *appsv1alpha1.APIcast) *appsv1.Deployment {
	tempDeployment := oldDeployment.DeepCopy()

	tempDeployment.ResourceVersion = ""
	tempDeployment.OwnerReferences = nil
	tempDeployment.Name = TMP_DEPLOYMENT_NAME
	tempDeployment.Spec.Selector.MatchLabels["3scale.io/temp"] = "true"
	tempDeployment.Spec.Template.Labels["3scale.io/temp"] = "true"

	tempDeployment.SetOwnerReferences(append(tempDeployment.GetOwnerReferences(), *apicastCR.GetOwnerReference()))

	return tempDeployment
}
