//go:build unit

package apicast

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testDefaultOpts() *APIcastOptions {
	opts := NewAPIcastOptions()
	opts.Namespace = "my-namespace"
	opts.DeploymentName = "apicast-apicast1"
	opts.Owner = &metav1.OwnerReference{}
	opts.ServiceName = "apicast-apicast1"
	opts.AdditionalPodAnnotations = map[string]string{}
	opts.ServiceAccountName = "my-sa"
	opts.Image = "example.com/my-registry/apicast-operator:latest"
	opts.AdminPortalCredentialsSecret = &v1.Secret{}
	opts.CommonLabels = map[string]string{}
	opts.PodTemplateLabels = map[string]string{}
	opts.PodLabelSelector = map[string]string{}

	return opts
}

func TestAPIcastDeploymentSelector(t *testing.T) {
	// This test must pass to ensure the upgrade procedure in picast_controller_deployment_upgrade.go
	// works as expected

	podLabelSelector := map[string]string{"a": "a1", "b": "b1"}

	opts := testDefaultOpts()
	opts.PodLabelSelector = podLabelSelector
	err := opts.Validate()
	if err != nil {
		t.Errorf("validation error: %v", err)
	}
	apicastFactory := NewAPIcast(opts)
	deployment := apicastFactory.Deployment()
	if deployment == nil {
		t.Error("deployment is nil")
	}
	if deployment.Spec.Selector == nil {
		t.Error("deployment selector is nil")
	}

	if !reflect.DeepEqual(podLabelSelector, deployment.Spec.Selector.MatchLabels) {
		t.Error("deployment selector does not match podlabelselector")
	}
}
