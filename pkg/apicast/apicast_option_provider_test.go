//go:build unit

package apicast

import (
	"context"
	"path"
	"reflect"
	"strings"
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

func TestOpentelemetryOptions(t *testing.T) {
	namespace := "my-ns"
	apicastConfigSecretName := "my-secret"
	embeddedConfigSecret := GetTestSecret(namespace, apicastConfigSecretName,
		map[string]string{"config.json": "{}"},
	)

	t.Run("Secret ref not set", func(subT *testing.T) {
		apicastCR := &appsv1alpha1.APIcast{
			ObjectMeta: metav1.ObjectMeta{
				Name: "instance1", Namespace: namespace,
			},
			Spec: appsv1alpha1.APIcastSpec{
				EmbeddedConfigurationSecretRef: &v1.LocalObjectReference{
					Name: apicastConfigSecretName,
				},
				OpenTelemetry: &appsv1alpha1.OpenTelemetrySpec{
					Enabled: &[]bool{true}[0],
				},
			},
		}

		objs := []runtime.Object{embeddedConfigSecret}
		cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		optsProvider := NewApicastOptionsProvider(apicastCR, cl)
		_, err := optsProvider.GetApicastOptions(context.TODO())
		if err == nil {
			subT.Fatal("get options should fail")
		}

		if !strings.Contains(err.Error(), "spec.openTelemetry.tracingConfigSecretRef: Invalid value") {
			subT.Fatalf("error unexpected: %s", err)
		}
	})

	t.Run("Secret key provided", func(subT *testing.T) {
		apicastCR := &appsv1alpha1.APIcast{
			ObjectMeta: metav1.ObjectMeta{
				Name: "instance1", Namespace: namespace,
			},
			Spec: appsv1alpha1.APIcastSpec{
				EmbeddedConfigurationSecretRef: &v1.LocalObjectReference{
					Name: apicastConfigSecretName,
				},
				OpenTelemetry: &appsv1alpha1.OpenTelemetrySpec{
					Enabled: &[]bool{true}[0],
					TracingConfigSecretRef: &v1.LocalObjectReference{
						Name: "secretName",
					},
					TracingConfigSecretKey: &[]string{"file1"}[0],
				},
			},
		}

		objs := []runtime.Object{embeddedConfigSecret}
		cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		optsProvider := NewApicastOptionsProvider(apicastCR, cl)
		opts, err := optsProvider.GetApicastOptions(context.TODO())
		if err != nil {
			subT.Fatalf("get options should not fail: %s", err)
		}

		if opts == nil {
			subT.Fatal("options should not be nil")
		}

		expectedOtelOptions := OpentelemetryConfig{
			Enabled:    true,
			SecretName: "secretName",
			ConfigFile: path.Join(OpentelemetryConfigMountBasePath, "file1"),
		}

		if !reflect.DeepEqual(expectedOtelOptions, opts.Opentelemetry) {
			subT.Fatalf("opentelemetry object not expected: %s",
				cmp.Diff(expectedOtelOptions, opts.Opentelemetry))
		}
	})

	t.Run("Secret key not provided", func(subT *testing.T) {
		tracingConfigSecret := GetTestSecret(namespace, "otelSecret",
			map[string]string{
				"c.json": "{}",
				"b.json": "{}",
				"a.json": "{}",
			},
		)
		apicastCR := &appsv1alpha1.APIcast{
			ObjectMeta: metav1.ObjectMeta{
				Name: "instance1", Namespace: namespace,
			},
			Spec: appsv1alpha1.APIcastSpec{
				EmbeddedConfigurationSecretRef: &v1.LocalObjectReference{
					Name: apicastConfigSecretName,
				},
				OpenTelemetry: &appsv1alpha1.OpenTelemetrySpec{
					Enabled: &[]bool{true}[0],
					TracingConfigSecretRef: &v1.LocalObjectReference{
						Name: "otelSecret",
					},
				},
			},
		}

		objs := []runtime.Object{embeddedConfigSecret, tracingConfigSecret}
		cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		optsProvider := NewApicastOptionsProvider(apicastCR, cl)
		opts, err := optsProvider.GetApicastOptions(context.TODO())
		if err != nil {
			subT.Fatalf("get options should not fail: %s", err)
		}

		if opts == nil {
			subT.Fatal("options should not be nil")
		}

		expectedOtelOptions := OpentelemetryConfig{
			Enabled:    true,
			SecretName: "otelSecret",
			ConfigFile: path.Join(OpentelemetryConfigMountBasePath, "a.json"),
		}

		if !reflect.DeepEqual(expectedOtelOptions, opts.Opentelemetry) {
			subT.Fatalf("opentelemetry object not expected: %s",
				cmp.Diff(expectedOtelOptions, opts.Opentelemetry))
		}
	})
}
