//go:build integration

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	"github.com/3scale/apicast-operator/pkg/apicast"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("APIcast v0.6.0 deployment upgrade procedure", func() {
	var testNamespace string

	BeforeEach(CreateNamespaceCallback(&testNamespace))
	AfterEach(DeleteNamespaceCallback(&testNamespace))

	Context("APIcast deployment v0.6.0 exists", func() {
		It("Should upgrade deployment pod selector", func() {
			// Create v0.6.0 deployment
			// Create v0.6.0 service
			// Create APIcast CR
			// Wait for a deployment with the new pod selector
			// Wait for temporary deployment to be deleted
			// Wait for service with the new pod selector

			fmt.Fprintf(GinkgoWriter, "create secret for namespace '%s'\n", testNamespace)
			err := testCreateAPIcastEmbeddedConfigurationSecret(context.TODO(), testNamespace)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				return testClient().Get(context.TODO(), client.ObjectKey{
					Name: testAPIcastEmbeddedConfigurationSecretName, Namespace: testNamespace},
					&v1.Secret{},
				) == nil
			}, 5*time.Minute, time.Second*5).Should(BeTrue())

			apicastName := "instance1"
			apicastCR := &appsv1alpha1.APIcast{
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

			apicastFactory, err := apicast.Factory(context.TODO(), apicastCR, testClient())
			Expect(err).ToNot(HaveOccurred())

			// v0.6.0 deployment selector
			fmt.Fprintf(GinkgoWriter, "create old deployment for namespace '%s'\n", testNamespace)
			apicastDeployment, err := apicastFactory.Deployment(context.TODO(), testClient())
			Expect(err).ToNot(HaveOccurred())
			apicastDeployment = apicastDeployment.DeepCopy()
			apicastDeployment.OwnerReferences = nil
			apicastDeployment.Spec.Selector.MatchLabels["rht.comp_ver"] = "v0.6.0"
			apicastDeployment.Spec.Template.Labels["rht.comp_ver"] = "v0.6.0"
			err = testClient().Create(context.TODO(), apicastDeployment)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				return testClient().Get(context.TODO(), client.ObjectKey{
					Name: apicastDeployment.Name, Namespace: testNamespace},
					&appsv1.Deployment{},
				) == nil
			}, 5*time.Minute, time.Second*5).Should(BeTrue())

			// v0.6.0 service
			fmt.Fprintf(GinkgoWriter, "create old service for namespace '%s'\n", testNamespace)
			apicastService := apicastFactory.Service().DeepCopy()
			apicastService.Spec.Selector["rht.comp_ver"] = "v0.6.0"
			apicastService.OwnerReferences = nil
			err = testClient().Create(context.TODO(), apicastService)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				return testClient().Get(context.TODO(), client.ObjectKey{
					Name: apicastService.Name, Namespace: testNamespace},
					&v1.Service{},
				) == nil
			}, 5*time.Minute, time.Second*5).Should(BeTrue())

			// apicast CR
			fmt.Fprintf(GinkgoWriter, "create apicast CR for namespace '%s'\n", testNamespace)
			err = testClient().Create(context.TODO(), apicastCR)
			Expect(err).ToNot(HaveOccurred())

			// deployment selector should be the current one
			newApicastDeployment, err := apicastFactory.Deployment(context.TODO(), testClient())
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				existingDeployment := &appsv1.Deployment{}
				err := testClient().Get(context.TODO(), client.ObjectKey{
					Name: newApicastDeployment.Name, Namespace: testNamespace},
					existingDeployment,
				)
				if err != nil {
					return false
				}

				return reflect.DeepEqual(newApicastDeployment.Spec.Selector, existingDeployment.Spec.Selector)
			}, 5*time.Minute, time.Second*5).Should(BeTrue())
			fmt.Fprintf(GinkgoWriter, "deployment upgraded for namespace '%s'\n", testNamespace)

			// temp deployment should be gone
			Eventually(func() bool {
				existingDeployment := &appsv1.Deployment{}
				err := testClient().Get(context.TODO(), client.ObjectKey{
					Name: TMP_DEPLOYMENT_NAME, Namespace: testNamespace},
					existingDeployment,
				)
				return errors.IsNotFound(err)
			}, 5*time.Minute, time.Second*5).Should(BeTrue())
			fmt.Fprintf(GinkgoWriter, "temp deployment deleted for namespace '%s'\n", testNamespace)

			// service selector should be the current one
			newApicastService := apicastFactory.Service()
			Eventually(func() bool {
				existingService := &v1.Service{}
				err := testClient().Get(context.TODO(), client.ObjectKey{
					Name: newApicastService.Name, Namespace: testNamespace},
					existingService,
				)
				if err != nil {
					return false
				}

				return reflect.DeepEqual(newApicastService.Spec.Selector, existingService.Spec.Selector)
			}, 5*time.Minute, time.Second*5).Should(BeTrue())
			fmt.Fprintf(GinkgoWriter, "service upgraded for namespace '%s'\n", testNamespace)
		})
	})
})
