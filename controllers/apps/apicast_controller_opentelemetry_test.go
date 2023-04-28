//go:build integration

package controllers

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	apicastpkg "github.com/3scale/apicast-operator/pkg/apicast"
	"github.com/3scale/apicast-operator/pkg/k8sutils"
)

const testAPIcastOtelSecretName = "apicast-otel-configuration"

var _ = Describe("APIcast controller Opentelemetry feature", func() {
	var testNamespace string

	const (
		retryInterval = time.Second * 5
	)

	BeforeEach(CreateNamespaceCallback(&testNamespace))
	AfterEach(DeleteNamespaceCallback(&testNamespace))

	Context("Basic usage", func() {
		It("APIcast is ready", func() {
			// Create an APIcast embedded configuration secret
			err := testCreateAPIcastEmbeddedConfigurationSecret(context.Background(), testNamespace)
			Expect(err).ToNot(HaveOccurred())

			err = testCreateAPIcastOtelConfigurationSecret(context.Background(), testNamespace)
			Expect(err).ToNot(HaveOccurred())

			// Create an APIcast
			apicastName := "example-apicast"
			apicast := &appsv1alpha1.APIcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apicastName,
					Namespace: testNamespace,
				},
				Spec: appsv1alpha1.APIcastSpec{
					EmbeddedConfigurationSecretRef: &v1.LocalObjectReference{
						Name: testAPIcastEmbeddedConfigurationSecretName,
					},
					OpenTelemetry: &appsv1alpha1.OpenTelemetrySpec{
						Enabled: &[]bool{true}[0],
						TracingConfigSecretRef: &v1.LocalObjectReference{
							Name: testAPIcastOtelSecretName,
						},
					},
				},
			}

			err = testClient().Create(context.Background(), apicast)
			Expect(err).ToNot(HaveOccurred())

			newApicast := &appsv1alpha1.APIcast{}
			// APIcast CR should report READY
			Eventually(func() bool {
				key := types.NamespacedName{Name: apicastName, Namespace: testNamespace}
				err := testClient().Get(context.Background(), key, newApicast)
				if err != nil {
					return false
				}

				return newApicast.Status.IsReady()
			}, 5*time.Minute, retryInterval).Should(BeTrue())

			// APIcast CR should have otel secret UID for watching secrets
			otelSecret := &v1.Secret{}
			key := types.NamespacedName{Name: testAPIcastOtelSecretName, Namespace: testNamespace}
			err = testClient().Get(context.Background(), key, otelSecret)
			Expect(err).ToNot(HaveOccurred())
			labelKeys := make([]string, 0, len(newApicast.GetLabels()))
			for k := range newApicast.GetLabels() {
				labelKeys = append(labelKeys, k)
			}
			Expect(labelKeys).To(ContainElement(apicastSecretLabelKey(string(otelSecret.GetUID()))))

			// Deployment should have expected env vars
			deployment := &appsv1.Deployment{}
			deploymentName := apicastpkg.APIcastDeploymentName(newApicast)
			key = types.NamespacedName{Name: deploymentName, Namespace: testNamespace}
			err = testClient().Get(context.Background(), key, deployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(deployment.Spec.Template.Spec.Containers).ToNot(BeEmpty())
			envVars := deployment.Spec.Template.Spec.Containers[0].Env
			Expect(envVars).To(ContainElement(k8sutils.EnvVarFromValue("OPENTELEMETRY", "1")))
		})
	})

	Context("When secret is not found", func() {
		It("Deployment should not be available", func() {
			// Create an APIcast embedded configuration secret
			err := testCreateAPIcastEmbeddedConfigurationSecret(context.Background(), testNamespace)
			Expect(err).ToNot(HaveOccurred())

			// Create an APIcast
			apicastName := "example-apicast"
			apicast := &appsv1alpha1.APIcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apicastName,
					Namespace: testNamespace,
				},
				Spec: appsv1alpha1.APIcastSpec{
					EmbeddedConfigurationSecretRef: &v1.LocalObjectReference{
						Name: testAPIcastEmbeddedConfigurationSecretName,
					},
					OpenTelemetry: &appsv1alpha1.OpenTelemetrySpec{
						Enabled: &[]bool{true}[0],
						TracingConfigSecretRef: &v1.LocalObjectReference{
							Name: "notExistingSecretName",
						},
					},
				},
			}

			err = testClient().Create(context.Background(), apicast)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				newApicast := &appsv1alpha1.APIcast{}
				key := types.NamespacedName{Name: apicastName, Namespace: testNamespace}
				err := testClient().Get(context.Background(), key, newApicast)
				if err != nil {
					return false
				}

				for i := range newApicast.Status.Conditions {
					if newApicast.Status.Conditions[i].Type == appsv1alpha1.ReadyConditionType {
						return strings.Contains(newApicast.Status.Conditions[i].Message, "Secret \"notExistingSecretName\" not found")
					}
				}

				return false
			}, 5*time.Minute, retryInterval).Should(BeTrue())

		})
	})
})

func testAPIcastOtelConfig() string {
	return `
exporter = "otlp"
processor = "simple"
[exporters.otlp]
host = "jaeger"
port = 4317
[processors.batch]
max_queue_size = 2048
schedule_delay_millis = 5000
max_export_batch_size = 512
[service]
name = "apicast" # Opentelemetry resource name
`
}

func testCreateAPIcastOtelConfigurationSecret(ctx context.Context, namespace string) error {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testAPIcastOtelSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			"otel.json": testAPIcastOtelConfig(),
		},
	}

	return testClient().Create(ctx, secret)
}
