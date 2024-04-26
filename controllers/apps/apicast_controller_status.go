package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	"github.com/3scale/apicast-operator/pkg/apicast"
	"github.com/3scale/apicast-operator/pkg/k8sutils"
)

func (r *APIcastReconciler) reconcileStatus(ctx context.Context, cr *appsv1alpha1.APIcast, specErr error) (ctrl.Result, error) {
	logger, _ := logr.FromContext(ctx)
	newStatus, err := r.calculateStatus(ctx, cr, specErr)
	if err != nil {
		return reconcile.Result{}, err
	}

	equalStatus := cr.Status.Equals(newStatus, logger)
	logger.V(1).Info("Status", "status is different", !equalStatus)
	logger.V(1).Info("Status", "generation is different", cr.Generation != cr.Status.ObservedGeneration)
	if equalStatus && cr.Generation == cr.Status.ObservedGeneration {
		// Steady state
		logger.V(1).Info("Status was not updated")
		return reconcile.Result{}, nil
	}

	// Save the generation number we acted on, otherwise we might wrongfully indicate
	// that we've seen a spec update when we retry.
	// TODO: This can clobber an update if we allow multiple agents to write to the
	// same status.
	newStatus.ObservedGeneration = cr.Generation

	logger.V(1).Info("Updating Status", "sequence no:", fmt.Sprintf("sequence No: %v->%v", cr.Status.ObservedGeneration, newStatus.ObservedGeneration))

	cr.Status = *newStatus
	updateErr := r.Client().Status().Update(ctx, cr)
	if updateErr != nil {
		// Ignore conflicts, resource might just be outdated.
		if errors.IsConflict(updateErr) {
			logger.Info("Failed to update status: resource might just be outdated")
			return reconcile.Result{Requeue: true}, nil
		}

		return reconcile.Result{}, fmt.Errorf("failed to update status: %w", updateErr)
	}
	return ctrl.Result{}, nil
}

func (r *APIcastReconciler) calculateStatus(ctx context.Context, cr *appsv1alpha1.APIcast, specErr error) (*appsv1alpha1.APIcastStatus, error) {
	newStatus := &appsv1alpha1.APIcastStatus{
		// Copy initial conditions. Otherwise, status will always be updated
		Conditions:         k8sutils.CopyConditions(cr.Status.Conditions),
		ObservedGeneration: cr.Status.ObservedGeneration,
	}

	availableCond, err := r.readyCondition(ctx, cr, specErr)
	if err != nil {
		return nil, err
	}

	meta.SetStatusCondition(&newStatus.Conditions, *availableCond)

	r.reconcileHpaWarningMessage(&newStatus.Conditions, cr)

	image, err := r.deploymentImage(ctx, cr)
	if err != nil {
		return nil, err
	}
	newStatus.Image = image

	return newStatus, nil
}

func (r *APIcastReconciler) deploymentImage(ctx context.Context, cr *appsv1alpha1.APIcast) (string, error) {
	dKey := client.ObjectKey{Name: apicast.APIcastDeploymentName(cr), Namespace: cr.Namespace}
	deployment := &appsv1.Deployment{}
	err := r.Client().Get(ctx, dKey, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			return "", nil
		}

		return "", err
	}

	return deployment.Spec.Template.Spec.Containers[0].Image, nil
}

func (r *APIcastReconciler) reconcileHpaWarningMessage(conditions *[]metav1.Condition, cr *appsv1alpha1.APIcast) {
	cond := &metav1.Condition{
		Type:    appsv1alpha1.WarningConditionType,
		Status:  metav1.ConditionTrue,
		Reason:  "HPA",
		Message: "HorizontalPodAutoscaling (Hpa) enabled overrides values applied to replicas",
	}

	// check if condition is already present
	foundCondition := meta.FindStatusCondition(*conditions, "Warning")

	// If hpa is enabled but the condition is not found, add it
	if cr.Spec.Hpa && foundCondition == nil {
		meta.SetStatusCondition(conditions, *cond)
	}

	// if hpa is disabled and condition is found, remove it
	if !cr.Spec.Hpa && foundCondition != nil {
		meta.RemoveStatusCondition(conditions, "Warning")
	}
}

func (r *APIcastReconciler) readyCondition(ctx context.Context, cr *appsv1alpha1.APIcast, specErr error) (*metav1.Condition, error) {
	cond := &metav1.Condition{
		Type:    appsv1alpha1.ReadyConditionType,
		Status:  metav1.ConditionTrue,
		Reason:  "Ready",
		Message: "APIcast is ready",
	}

	if specErr != nil {
		cond.Status = metav1.ConditionFalse
		cond.Reason = "ReconcilliationError"
		cond.Message = specErr.Error()
		return cond, nil
	}

	reason, err := r.checkDeploymentAvailable(ctx, cr)
	if err != nil {
		return nil, err
	}
	if reason != nil {
		cond.Status = metav1.ConditionFalse
		cond.Reason = "DeploymentNotReady"
		cond.Message = *reason
		return cond, nil
	}

	return cond, nil
}

func (r *APIcastReconciler) checkDeploymentAvailable(ctx context.Context, cr *appsv1alpha1.APIcast) (*string, error) {
	dKey := client.ObjectKey{Name: apicast.APIcastDeploymentName(cr), Namespace: cr.Namespace}
	deployment := &appsv1.Deployment{}
	err := r.Client().Get(ctx, dKey, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	if err != nil && errors.IsNotFound(err) {
		tmp := err.Error()
		return &tmp, nil
	}

	availableCondition := k8sutils.FindDeploymentStatusCondition(deployment.Status.Conditions, appsv1.DeploymentAvailable)
	if availableCondition == nil {
		tmp := "Available condition not found"
		return &tmp, nil
	}

	if availableCondition.Status != corev1.ConditionTrue {
		return &availableCondition.Message, nil
	}

	return nil, nil
}
