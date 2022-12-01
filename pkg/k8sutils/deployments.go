package k8sutils

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func FindDeploymentStatusCondition(conditions []appsv1.DeploymentCondition, condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}

	return nil
}

func IsStatusConditionTrue(conditions []appsv1.DeploymentCondition, condType appsv1.DeploymentConditionType) bool {
	cond := FindDeploymentStatusCondition(conditions, condType)
	if cond == nil {
		return false
	}

	return cond.Status == v1.ConditionTrue
}
