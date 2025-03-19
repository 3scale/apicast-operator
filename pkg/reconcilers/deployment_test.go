package reconcilers

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeploymentConfigTopologySpreadConstraintsMutator(t *testing.T) {
	testTopologySpreadConstraint1 := []corev1.TopologySpreadConstraint{
		corev1.TopologySpreadConstraint{
			TopologyKey:       "topologyKey1",
			WhenUnsatisfiable: "DoNotSchedule",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "3scale-api-management"},
			},
		},
	}
	testTopologySpreadConstraint2 := []corev1.TopologySpreadConstraint{
		corev1.TopologySpreadConstraint{
			TopologyKey:       "topologyKey2",
			WhenUnsatisfiable: "ScheduleAnyway",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "3scale-api-management", "threescale_component": "system"},
			},
		},
	}

	dcFactory := func(topologySpreadConstraint []corev1.TopologySpreadConstraint) *appsv1.Deployment {
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
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						TopologySpreadConstraints: topologySpreadConstraint,
					},
				},
			},
		}
	}

	cases := []struct {
		testName                          string
		existingTopologySpreadConstraints []corev1.TopologySpreadConstraint
		desiredTopologySpreadConstraints  []corev1.TopologySpreadConstraint
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
			update := DeploymentConfigTopologySpreadConstraintsMutator(desired, existing)

			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing.Spec.Template.Spec.TopologySpreadConstraints, desired.Spec.Template.Spec.TopologySpreadConstraints) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.TopologySpreadConstraints, desired.Spec.Template.Spec.TopologySpreadConstraints))
			}
		})
	}

}
