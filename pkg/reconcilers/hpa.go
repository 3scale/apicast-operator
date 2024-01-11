package reconcilers

import (
	"fmt"
	"reflect"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	helper "github.com/3scale/apicast-operator/pkg/helper"
	"github.com/3scale/apicast-operator/pkg/k8sutils"
	hpa "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HpaMutateFn is a function which mutates the existing Hpa into it's desired state.
type HpaMutateFn func(desired, existing *hpa.HorizontalPodAutoscaler) bool

func HpaGenericMutator() MutateFn {
	return func(existingObj, desiredObj k8sutils.KubernetesObject) (bool, error) {
		existing, ok := existingObj.(*hpa.HorizontalPodAutoscaler)
		if !ok {
			return false, fmt.Errorf("%T is not a *v2.HorizontalPodAutoscaler", existingObj)
		}
		desired, ok := desiredObj.(*hpa.HorizontalPodAutoscaler)
		if !ok {
			return false, fmt.Errorf("%T is not a *v2.HorizontalPodAutoscaler", desiredObj)
		}

		updated := false
		if !reflect.DeepEqual(desired.Spec, existing.Spec) {
			existing.Spec = desired.Spec
			updated = true
		}

		return updated, nil
	}
}

func HpaCreateOnlyMutator() MutateFn {
	return func(existingObj, desiredObj k8sutils.KubernetesObject) (bool, error) {
		return false, nil
	}
}

func HpaDeleteMutator() MutateFn {
	return func(existingObj, desiredObj k8sutils.KubernetesObject) (bool, error) {
		return false, nil
	}
}

func HpaCR(cr *appsv1alpha1.APIcast) *hpa.HorizontalPodAutoscaler {
	minPods := helper.Int32Ptr(1)
	maxPods := int32(5)
	cpuPercent := helper.Int32Ptr(90)
	memoryPercent := helper.Int32Ptr(90)

	// Override values if they are specified in the APICast CR
	if cr.Spec.Hpa.MinPods != nil {
		minPods = cr.Spec.Hpa.MinPods
	}

	if cr.Spec.Hpa.MaxPods != 0 {
		maxPods = cr.Spec.Hpa.MaxPods
	}

	if cr.Spec.Hpa.CpuPercent != nil {
		cpuPercent = cr.Spec.Hpa.CpuPercent
	}

	if cr.Spec.Hpa.MemoryPercent != nil {
		memoryPercent = cr.Spec.Hpa.MemoryPercent
	}

	return &hpa.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: hpa.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: hpa.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       fmt.Sprintf("apicast-%s", cr.Name),
				APIVersion: "apps/v1",
			},
			MinReplicas: minPods,
			MaxReplicas: maxPods,
			Metrics: []hpa.MetricSpec{
				{
					Type: hpa.ResourceMetricSourceType,
					Resource: &hpa.ResourceMetricSource{
						Name: "memory",
						Target: hpa.MetricTarget{
							Type:               hpa.UtilizationMetricType,
							AverageUtilization: memoryPercent,
						},
					},
				},
				{
					Type: hpa.ResourceMetricSourceType,
					Resource: &hpa.ResourceMetricSource{
						Name: "cpu",
						Target: hpa.MetricTarget{
							Type:               hpa.UtilizationMetricType,
							AverageUtilization: cpuPercent,
						},
					},
				},
			},
		},
	}
}
