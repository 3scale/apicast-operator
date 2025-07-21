package reconcilers

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAffinityMutator(t *testing.T) {
	deploymentFactory := func(affinitiy *v1.Affinity) *appsv1.Deployment {
		return &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.PodSpec{
						Affinity: affinitiy,
					},
				},
			},
		}
	}

	affinitiy1 := &v1.Affinity{
		PodAntiAffinity: &v1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: v1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"pod": "label",
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}

	affinitiy2 := &v1.Affinity{
		PodAntiAffinity: &v1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
				{
					Weight: 999,
					PodAffinityTerm: v1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"pod": "label",
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		existing *appsv1.Deployment
		desired  *appsv1.Deployment
		expected bool
	}{
		{
			"test false when desired and existing are the same",
			deploymentFactory(affinitiy1),
			deploymentFactory(affinitiy1),
			false,
		},
		{
			"test true when desired and existing do not match",
			deploymentFactory(affinitiy1),
			deploymentFactory(affinitiy2),
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changed := DeploymentAffinityMutator(tc.desired, tc.existing)
			if changed != tc.expected {
				t.Error("expected mutator return ", tc.expected, " but got: ", changed)
			}
		})
	}
}
