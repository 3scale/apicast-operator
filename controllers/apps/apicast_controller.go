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
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	"github.com/3scale/apicast-operator/pkg/apicast"
	"github.com/3scale/apicast-operator/pkg/reconcilers"
	"github.com/3scale/apicast-operator/version"
)

// APIcastReconciler reconciles a APIcast object
type APIcastReconciler struct {
	reconcilers.BaseControllerReconciler
	Log logr.Logger
}

// blank assignment to verify that ReconcileAPIcast implements reconcile.Reconciler
var _ reconcile.Reconciler = &APIcastReconciler{}

const (
	APIcastOperatorVersionAnnotation = "apicast.apps.3scale.net/operator-version"
)

// +kubebuilder:rbac:groups=apps.3scale.net,resources=apicasts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.3scale.net,resources=apicasts/status,verbs=get;update;patch

func (r *APIcastReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	log := r.Log.WithValues("apicast", req.NamespacedName)

	// your logic here
	log.Info("Reconciling APIcast")

	instance, err := r.getAPIcast(req)
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
		err = r.updateAPIcastOperatorVersionInAnnotations(instance, log)
		if err != nil {
			return ctrl.Result{}, err
		}
		log.Info("APIcast operator version in annotations set. Requeuing request...")
		return ctrl.Result{Requeue: true}, err
	}

	if instance.ObjectMeta.Annotations[APIcastOperatorVersionAnnotation] != version.Version {
		log.Info("APIcast operator version in annotations does not match expected version. Applying upgrade procedure...")
		upgradeReconcileResult, err := r.upgradeAPIcast()
		if err != nil {
			log.Error(err, "Error upgrading APIcast")
			return ctrl.Result{}, err
		}
		if upgradeReconcileResult.Requeue {
			return upgradeReconcileResult, nil
		}
		log.Info("APIcast upgrade procedure applied")
		log.Info("Setting APIcast operator version in annotations...")
		err = r.updateAPIcastOperatorVersionInAnnotations(instance, log)
		if err != nil {
			return ctrl.Result{}, err
		}
		log.Info("APIcast operator version in annotations set. Requeuing request...")
		return ctrl.Result{Requeue: true}, nil
	}

	baseReconciler := reconcilers.NewBaseReconciler(r.Client(), r.APIClientReader(), r.Scheme(), log)
	logicReconciler := NewAPIcastLogicReconciler(baseReconciler, instance)
	result, err := logicReconciler.Reconcile()
	if err != nil {
		// Ignore conflicts, resource might just be outdated.
		if errors.IsConflict(err) {
			log.Info("Resource update conflict error. Requeuing...", "error", err)
			return ctrl.Result{Requeue: true}, nil
		}

		log.Error(err, "Main reconciler")
		return result, err
	}
	if result.Requeue {
		log.Info("Requeuing request...")
		return result, nil
	}
	log.Info("APIcast logic reconciled")

	result, err = r.updateStatus(instance, &logicReconciler)
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.APIcast{}).
		Owns(&corev1.Secret{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&extensionsv1beta1.Ingress{}).
		Complete(r)
}

func (r *APIcastReconciler) updateStatus(instance *appsv1alpha1.APIcast, reconciler *APIcastLogicReconciler) (ctrl.Result, error) {
	apicastFactory, err := apicast.Factory(instance, r.Client())
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredDeployment := apicastFactory.Deployment()
	apicastDeployment := &appsv1.Deployment{}
	err = r.Client().Get(context.TODO(), types.NamespacedName{Name: desiredDeployment.Name, Namespace: desiredDeployment.Namespace}, apicastDeployment)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	deployedImage := apicastDeployment.Spec.Template.Spec.Containers[0].Image
	if instance.Status.Image != deployedImage {
		instance.Status.Image = deployedImage
		err = r.Client().Status().Update(context.TODO(), instance)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func (r *APIcastReconciler) upgradeAPIcast() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *APIcastReconciler) getAPIcast(request ctrl.Request) (*appsv1alpha1.APIcast, error) {
	instance := appsv1alpha1.APIcast{}
	err := r.Client().Get(context.TODO(), request.NamespacedName, &instance)
	return &instance, err
}

func (r *APIcastReconciler) updateAPIcastOperatorVersionInAnnotations(instance *appsv1alpha1.APIcast, logger logr.Logger) error {
	instance.Annotations[APIcastOperatorVersionAnnotation] = version.Version
	err := r.Client().Update(context.TODO(), instance)
	if err != nil {
		logger.Error(err, "Error setting APIcast operator version in annotations")
	}
	return err
}
