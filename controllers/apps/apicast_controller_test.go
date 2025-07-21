//go:build integration

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	apicastpkg "github.com/3scale/apicast-operator/pkg/apicast"
	"github.com/3scale/apicast-operator/pkg/k8sutils"
)

const (
	testCustomEnvironmentSecretName            = "custom-env-1"
	testAPIcastEmbeddedConfigurationSecretName = "apicast-embedded-configuration"
)

var _ = Describe("APIcast controller", func() {
	const (
		retryInterval = time.Second * 5
	)
	var testNamespace string
	apicastName := "example-apicast"

	BeforeEach(CreateNamespaceCallback(&testNamespace))
	AfterEach(DeleteNamespaceCallback(&testNamespace))

	Context("Run directly without existing APIcast", func() {
		It("Should create successfully", func() {
			Expect(1).To(Equal(1))
		})
	})

	// Test basic APIcast deployment
	Context("Run with basic APIcast deployment", func() {
		It("Should create successfully", func() {
			start := time.Now()

			// Create a custom environment secret
			err := testCreateCustomEnvironmentSecret(context.Background(), testNamespace)
			Expect(err).ToNot(HaveOccurred())

			// Get the newly created custom environment secret for later
			customEnvSecret := &v1.Secret{}
			customEnvSecretLookupKey := types.NamespacedName{Name: testCustomEnvironmentSecretName, Namespace: testNamespace}
			err = testClient().Get(context.Background(), customEnvSecretLookupKey, customEnvSecret)
			Expect(err).ToNot(HaveOccurred())

			// Create an APIcast embedded configuration secret
			err = testCreateAPIcastEmbeddedConfigurationSecret(context.Background(), testNamespace)
			Expect(err).ToNot(HaveOccurred())

			// Get the newly created embedded configuration secret for later
			configSecret := &v1.Secret{}
			configSecretLookupKey := types.NamespacedName{Name: testAPIcastEmbeddedConfigurationSecretName, Namespace: testNamespace}
			err = testClient().Get(context.Background(), configSecretLookupKey, configSecret)
			Expect(err).ToNot(HaveOccurred())

			// Create an APIcast
			apicast := &appsv1alpha1.APIcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apicastName,
					Namespace: testNamespace,
				},
				Spec: appsv1alpha1.APIcastSpec{
					EmbeddedConfigurationSecretRef: &v1.LocalObjectReference{
						Name: testAPIcastEmbeddedConfigurationSecretName,
					},
					CustomEnvironments: []appsv1alpha1.CustomEnvironmentSpec{
						{
							SecretRef: &v1.LocalObjectReference{
								Name: testCustomEnvironmentSecretName,
							},
						},
					},
				},
			}

			err = testClient().Create(context.Background(), apicast)
			Expect(err).ToNot(HaveOccurred())

			// Check that the APIcast labels contain the secret UIDs
			Eventually(func() bool {
				reconciledApicast := &appsv1alpha1.APIcast{}
				reconciledApicastKey := types.NamespacedName{Name: apicastName, Namespace: testNamespace}
				err = testClient().Get(context.Background(), reconciledApicastKey, reconciledApicast)
				if err != nil {
					return false
				}

				expectedLabels := map[string]string{
					fmt.Sprintf("%s%s", APIcastSecretLabelPrefix, string(configSecret.GetUID())):    "false",
					fmt.Sprintf("%s%s", APIcastSecretLabelPrefix, string(customEnvSecret.GetUID())): "true",
				}

				// Then verify that the hash matches the hashed config secret
				return reflect.DeepEqual(reconciledApicast.Labels, expectedLabels)
			}, 1*time.Minute, retryInterval).Should(BeTrue())

			// Check that the corresponding APIcast hashed Secret has been created and is accurate
			hashedSecretLookupKey := types.NamespacedName{Name: apicastpkg.HashedSecretName, Namespace: testNamespace}
			Eventually(func() bool {
				// First get the master hashed secret
				hashedSecret := &v1.Secret{}
				err := testClient().Get(context.Background(), hashedSecretLookupKey, hashedSecret)
				if err != nil {
					return false
				}

				// Then verify that the hash matches the hashed custom environment secret
				return k8sutils.SecretStringDataFromData(hashedSecret)[testCustomEnvironmentSecretName] == apicastpkg.HashSecret(customEnvSecret.Data)
			}, 1*time.Minute, retryInterval).Should(BeTrue())

			// Check that the corresponding APIcast K8s Deployment has been created
			apicastDeploymentName := apicastpkg.APIcastDeploymentName(apicast)
			apicastDeploymentLookupKey := types.NamespacedName{Name: apicastDeploymentName, Namespace: testNamespace}
			Eventually(func() bool {
				deployment := &appsv1.Deployment{}
				err := testClient().Get(context.Background(), apicastDeploymentLookupKey, deployment)
				if err != nil {
					return false
				}

				return k8sutils.IsStatusConditionTrue(deployment.Status.Conditions, appsv1.DeploymentAvailable)
			}, 5*time.Minute, retryInterval).Should(BeTrue())

			Eventually(func() bool {
				newApicast := &appsv1alpha1.APIcast{}
				key := types.NamespacedName{Name: apicastName, Namespace: testNamespace}
				err := testClient().Get(context.Background(), key, newApicast)
				if err != nil {
					return false
				}

				return newApicast.Status.IsReady()
			}, 5*time.Minute, retryInterval).Should(BeTrue())

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "APIcast creation and availability took '%s'\n", elapsed)
		})
	})

	// Test APIcast deployment with ExposedHost

	Context("Run with APIcast with ExposedHost Deployment", func() {
		It("Should create successfully", func() {
			start := time.Now()

			// Create an APIcast embedded configuration secret
			err := testCreateAPIcastEmbeddedConfigurationSecret(context.Background(), testNamespace)
			Expect(err).ToNot(HaveOccurred())

			// Create an APIcast
			apicast := &appsv1alpha1.APIcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apicastName,
					Namespace: testNamespace,
				},
				Spec: appsv1alpha1.APIcastSpec{
					ExposedHost: &appsv1alpha1.APIcastExposedHost{
						Host:             "apicast.example.com",
						IngressClassName: ptr.To("default-openshift"),
					},
					EmbeddedConfigurationSecretRef: &v1.LocalObjectReference{
						Name: testAPIcastEmbeddedConfigurationSecretName,
					},
				},
			}
			err = testClient().Create(context.Background(), apicast)
			Expect(err).ToNot(HaveOccurred())

			// Check that the correspondig APIcast K8s Deployment has been created
			apicastDeploymentName := "apicast-" + apicastName
			apicastDeploymentLookupKey := types.NamespacedName{Name: apicastDeploymentName, Namespace: testNamespace}
			createdDeployment := &appsv1.Deployment{}
			Eventually(func() bool {
				err := testClient().Get(context.Background(), apicastDeploymentLookupKey, createdDeployment)
				return err == nil
			}, 5*time.Minute, retryInterval).Should(BeTrue())

			// Check that the correspondig IU K8s Ingress has been created
			apicastIngressName := "apicast-" + apicastName
			apicastIngressLookupKey := types.NamespacedName{Name: apicastIngressName, Namespace: testNamespace}
			createdIngress := &networkingv1.Ingress{}
			Eventually(func() bool {
				err := testClient().Get(context.Background(), apicastIngressLookupKey, createdIngress)
				return err == nil
			}, 5*time.Minute, retryInterval).Should(BeTrue())

			Expect(*createdIngress.Spec.IngressClassName).To(Equal("default-openshift"))

			elapsed := time.Since(start)
			By(fmt.Sprintf("APIcast creation and availability took %s seconds", elapsed))
		})
	})

	Context("Run APIcast with custom Affinity settings", func() {
		var apicast *appsv1alpha1.APIcast

		affinity := &v1.Affinity{
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

		BeforeEach(func(ctx SpecContext) {
			// Create an APIcast embedded configuration secret
			err := testCreateAPIcastEmbeddedConfigurationSecret(context.Background(), testNamespace)
			Expect(err).ToNot(HaveOccurred())

			// Create an APIcast
			apicast = &appsv1alpha1.APIcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apicastName,
					Namespace: testNamespace,
				},
				Spec: appsv1alpha1.APIcastSpec{
					EmbeddedConfigurationSecretRef: &v1.LocalObjectReference{
						Name: testAPIcastEmbeddedConfigurationSecretName,
					},
				},
			}
		})

		It("Create a new APIcast with specific affinity", func(ctx SpecContext) {
			apicast.Spec.Affinity = affinity.DeepCopy()

			err := testClient().Create(ctx, apicast)
			Expect(err).ToNot(HaveOccurred())

			// Check that the correspondig APIcast K8s Deployment has been created
			apicastDeploymentName := "apicast-" + apicastName
			apicastDeploymentLookupKey := types.NamespacedName{Name: apicastDeploymentName, Namespace: testNamespace}
			createdDeployment := &appsv1.Deployment{}
			Eventually(func() bool {
				err := testClient().Get(ctx, apicastDeploymentLookupKey, createdDeployment)
				return err == nil
			}, 5*time.Minute, retryInterval).Should(BeTrue())
			Expect(createdDeployment.Spec.Template.Spec.Affinity).To(Equal(affinity))
		})

		It("Should update the deployment with affinity", func(ctx SpecContext) {
			err := testClient().Create(ctx, apicast)
			Expect(err).ToNot(HaveOccurred())

			// Check that the correspondig APIcast K8s Deployment has been created
			apicastDeploymentName := "apicast-" + apicastName
			apicastDeploymentLookupKey := types.NamespacedName{Name: apicastDeploymentName, Namespace: testNamespace}
			createdDeployment := &appsv1.Deployment{}
			Eventually(func() bool {
				err := testClient().Get(ctx, apicastDeploymentLookupKey, createdDeployment)
				return err == nil
			}, 5*time.Minute, retryInterval).Should(BeTrue())

			Expect(createdDeployment.Spec.Template.Spec.Affinity).To(BeNil())

			updatedAPIcast := appsv1alpha1.APIcast{}
			Eventually(func(g Gomega) {
				g.Expect(testClient().Get(ctx, types.NamespacedName{
					Name:      apicast.Name,
					Namespace: testNamespace,
				}, &updatedAPIcast)).To(Succeed())
				updatedAPIcast.Spec.Affinity = affinity.DeepCopy()
				g.Expect(testClient().Update(context.Background(), &updatedAPIcast)).Should(Succeed())
			}, 5*time.Minute, retryInterval).Should(Succeed())

			Eventually(func(g Gomega) {
				newDeployment := &appsv1.Deployment{}
				g.Expect(testClient().Get(context.Background(), apicastDeploymentLookupKey, newDeployment)).To(Succeed())
				g.Expect(newDeployment.Spec.Template.Spec.Affinity).Should(Equal(affinity))
			}, 5*time.Minute, retryInterval).Should(Succeed())
		})
	})
})

func testAPIcastEmbeddedConfigurationContent() string {
	return `{
  "services": [
    {
      "proxy": {
        "policy_chain": [
          { "name": "apicast.policy.upstream",
            "configuration": {
              "rules": [{
                "regex": "/",
                "url": "http://echo-api.3scale.net"
              }]
            }
          }
        ]
      }
    }
  ]
}`
}

func testCreateAPIcastEmbeddedConfigurationSecret(ctx context.Context, namespace string) error {
	embeddedConfigurationContent := testAPIcastEmbeddedConfigurationContent()
	embeddedConfigSecret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testAPIcastEmbeddedConfigurationSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			"config.json": embeddedConfigurationContent,
		},
	}

	return testClient().Create(ctx, &embeddedConfigSecret)
}

func testCustomEnvironmentContent() string {
	return `
    local cjson = require('cjson')
    local PolicyChain = require('apicast.policy_chain')
    local policy_chain = context.policy_chain
    
    local logging_policy_config = cjson.decode([[
    {
      "enable_access_logs": false,
      "custom_logging": "\"{{request}}\" to service {{service.name}} and {{service.id}}"
    }
    ]])
    
    policy_chain:insert( PolicyChain.load_policy('logging', 'builtin', logging_policy_config), 1)
    
    return {
      policy_chain = policy_chain,
      port = { metrics = 9421 },
    }
`
}

func testCreateCustomEnvironmentSecret(ctx context.Context, namespace string) error {
	customEnvironmentSecret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testCustomEnvironmentSecretName,
			Namespace: namespace,
			Labels: map[string]string{
				"apicast.apps.3scale.net/watched-by": "apicast",
			},
		},
		StringData: map[string]string{
			"custom_env.lua":  testCustomEnvironmentContent(),
			"custom_env2.lua": testCustomEnvironmentContent(),
		},
	}
	return testClient().Create(ctx, &customEnvironmentSecret)
}
