package reconcilers

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func TestTolerationsMutator(t *testing.T) {
	deploymentFactory := func(toleration []v1.Toleration) *appsv1.Deployment {
		return &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.PodSpec{
						Tolerations: toleration,
					},
				},
			},
		}
	}

	testTolerations1 := []v1.Toleration{
		{
			Key:      "key1",
			Effect:   v1.TaintEffectNoExecute,
			Operator: v1.TolerationOpEqual,
			Value:    "val1",
		},
		{
			Key:      "key2",
			Effect:   v1.TaintEffectNoExecute,
			Operator: v1.TolerationOpEqual,
			Value:    "val2",
		},
	}

	testTolerations2 := []v1.Toleration{
		{
			Key:      "key3",
			Effect:   v1.TaintEffectNoExecute,
			Operator: v1.TolerationOpEqual,
			Value:    "val3",
		},
		{
			Key:      "key4",
			Effect:   v1.TaintEffectNoExecute,
			Operator: v1.TolerationOpEqual,
			Value:    "val4",
		},
	}

	cases := []struct {
		testName       string
		existing       *appsv1.Deployment
		desired        *appsv1.Deployment
		expectedResult bool
	}{
		{"NothingToReconcile", deploymentFactory(nil), deploymentFactory(nil), false},
		{"EqualTolerations", deploymentFactory(testTolerations1), deploymentFactory(testTolerations1), false},
		{"DifferentTolerations", deploymentFactory(testTolerations1), deploymentFactory(testTolerations2), true},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			update := DeploymentTolerationsMutator(tc.desired, tc.existing)
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}

			if !reflect.DeepEqual(tc.existing.Spec.Template.Spec.Tolerations, tc.desired.Spec.Template.Spec.Tolerations) {
				subT.Fatalf("mismatch values: expected: %v, got %v", tc.desired.Spec.Template.Spec.Tolerations, tc.existing.Spec.Template.Spec.Tolerations)
			}
		})
	}
}

func TestDeploymentConfigTopologySpreadConstraintsMutator(t *testing.T) {
	testTopologySpreadConstraint1 := []v1.TopologySpreadConstraint{
		{
			TopologyKey:       "topologyKey1",
			WhenUnsatisfiable: "DoNotSchedule",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "3scale-api-management"},
			},
		},
	}
	testTopologySpreadConstraint2 := []v1.TopologySpreadConstraint{
		{
			TopologyKey:       "topologyKey2",
			WhenUnsatisfiable: "ScheduleAnyway",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "3scale-api-management", "threescale_component": "system"},
			},
		},
	}

	dcFactory := func(topologySpreadConstraint []v1.TopologySpreadConstraint) *appsv1.Deployment {
		return &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDC",
				Namespace: "myNS",
			},
			Spec: appsv1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						TopologySpreadConstraints: topologySpreadConstraint,
					},
				},
			},
		}
	}

	cases := []struct {
		testName                          string
		existingTopologySpreadConstraints []v1.TopologySpreadConstraint
		desiredTopologySpreadConstraints  []v1.TopologySpreadConstraint
		expectedResult                    bool
	}{
		{"NothingToReconcile", nil, nil, false},
		{"EqualTopologies", testTopologySpreadConstraint1, testTopologySpreadConstraint1, false},
		{"DifferentTopologie", testTopologySpreadConstraint1, testTopologySpreadConstraint2, true},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dcFactory(tc.existingTopologySpreadConstraints)
			desired := dcFactory(tc.desiredTopologySpreadConstraints)
			update := DeploymentTopologySpreadConstraintsMutator(desired, existing)

			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing.Spec.Template.Spec.TopologySpreadConstraints, desired.Spec.Template.Spec.TopologySpreadConstraints) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.TopologySpreadConstraints, desired.Spec.Template.Spec.TopologySpreadConstraints))
			}
		})
	}
}
