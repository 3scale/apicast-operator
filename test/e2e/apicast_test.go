package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/3scale/apicast-operator/test/e2e/e2eutil"

	"github.com/3scale/apicast-operator/pkg/apis"
	appsgroup "github.com/3scale/apicast-operator/pkg/apis/apps"
	appsv1alpha1 "github.com/3scale/apicast-operator/pkg/apis/apps/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	frameworke2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	APIcastEmbeddedConfigurationSecretName = "apicast-embedded-configuration"
)

func TestAPIcastBasicDeployment(t *testing.T) {
	err := registerAPIcastTypeInTestFramework()
	if err != nil {
		t.Fatalf(err.Error())
	}

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	f := framework.Global
	operatorName := "apicast-operator"
	err = initializeClusterResources(t, ctx, f, namespace, operatorName)
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("initialized cluster resources")

	apicastName := "example-apicast"
	apicast := &appsv1alpha1.APIcast{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-apicast",
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIcastSpec{
			EmbeddedConfigurationSecretRef: &v1.LocalObjectReference{
				Name: APIcastEmbeddedConfigurationSecretName,
			},
		},
	}

	var start time.Time
	var elapsed time.Duration

	start = time.Now()

	err = f.Client.Create(context.TODO(), apicast, &framework.CleanupOptions{TestContext: ctx, Timeout: 5 * time.Minute, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	elapsed = time.Since(start)

	err = createAPIcastEmbeddedConfigurationSecret(t, ctx, f, namespace)
	if err != nil {
		t.Fatal(err)
	}

	apicastDeploymentName := "apicast-" + apicastName
	err = frameworke2eutil.WaitForDeployment(t, f.KubeClient, namespace, apicastDeploymentName, 1, retryInterval, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("APIcast creation and availability took %s seconds", elapsed)
}

func apicastEmbeddedConfigurationContent() string {
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

func createAPIcastEmbeddedConfigurationSecret(t *testing.T, ctx *framework.TestCtx, f *framework.Framework, namespace string) error {
	embeddedConfigurationContent := apicastEmbeddedConfigurationContent()
	embeddedConfigSecret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      APIcastEmbeddedConfigurationSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			"config.json": embeddedConfigurationContent,
		},
	}
	err := f.Client.Create(context.TODO(), &embeddedConfigSecret, &framework.CleanupOptions{TestContext: ctx, Timeout: 5 * time.Minute, RetryInterval: retryInterval})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForSecret(t, f.KubeClient, namespace, APIcastEmbeddedConfigurationSecretName, retryInterval, 5*time.Minute)
	return err
}

func registerAPIcastTypeInTestFramework() error {
	apicastList := &appsv1alpha1.APIcastList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1alpha1.SchemeGroupVersion.Version,
			Kind:       appsgroup.APIcastKind,
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, apicastList)
	if err != nil {
		return fmt.Errorf("failed to add custom resource scheme to framework: %v", err)
	}
	return nil
}

func initializeClusterResources(t *testing.T, ctx *framework.TestCtx, f *framework.Framework, namespace, operatorName string) error {
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: 5 * time.Minute, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	err = frameworke2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, operatorName, 1, retryInterval, timeout)
	return err
}
