//go:build unit

package apicast

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
)

func TestPodLabelSelector(t *testing.T) {
	// This test must pass to ensure the upgrade procedure in picast_controller_deployment_upgrade.go
	// works as expected

	apicastConfigSecretName := "my-secret"
	namespace := "my-ns"

	embeddedConfigSecret := GetTestSecret(namespace, apicastConfigSecretName,
		map[string]string{"config.json": "{}"},
	)

	apicastCR := &appsv1alpha1.APIcast{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance1",
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIcastSpec{
			EmbeddedConfigurationSecretRef: &v1.LocalObjectReference{
				Name: apicastConfigSecretName,
			},
		},
	}

	objs := []runtime.Object{embeddedConfigSecret}
	cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	optsProvider := NewApicastOptionsProvider(apicastCR, cl)
	opts, err := optsProvider.GetApicastOptions(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	expectedPodLabelSelectors := map[string]string{"deployment": "apicast-instance1"}
	if !reflect.DeepEqual(expectedPodLabelSelectors, opts.PodLabelSelector) {
		t.Fatalf("PodLabelSelector not expected: %s",
			cmp.Diff(expectedPodLabelSelectors, opts.PodLabelSelector))
	}
}
