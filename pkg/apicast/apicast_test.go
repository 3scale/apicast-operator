//go:build unit

package apicast

import (
	"context"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testNamespace                              = "my-namespace"
	testAPIcastEmbeddedConfigurationSecretName = "apicast-embedded-configuration"
	testCustomEnvironmentSecretName            = "custom-env-1"
)

func testDefaultOpts() *APIcastOptions {
	opts := NewAPIcastOptions()
	opts.Namespace = testNamespace
	opts.DeploymentName = "apicast-apicast1"
	opts.Owner = &metav1.OwnerReference{}
	opts.ServiceName = "apicast-apicast1"
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
	deployment, err := apicastFactory.Deployment(context.TODO(), fake.NewFakeClient())
	if err != nil {
		t.Errorf("error getting deployment: %v", err)
	}
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

func TestAPIcast_HashedSecret(t *testing.T) {
	secrets := []runtime.Object{testAPIcastEmbeddedConfigurationSecret(), testCustomEnvironmentSecret()}
	secretRefs := []*v1.LocalObjectReference{
		{
			Name: testAPIcastEmbeddedConfigurationSecretName,
		},
		{
			Name: testCustomEnvironmentSecretName,
		},
	}

	type args struct {
		ctx        context.Context
		k8sclient  client.Client
		secretRefs []*v1.LocalObjectReference
	}
	tests := []struct {
		name    string
		args    args
		want    *v1.Secret
		wantErr bool
	}{
		{
			name: "succesfully create empty hashed secret",
			args: args{
				ctx:        context.TODO(),
				k8sclient:  fake.NewFakeClient(secrets...),
				secretRefs: make([]*v1.LocalObjectReference, 0),
			},
			want: &v1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      HashedSecretName,
					Namespace: testDefaultOpts().Namespace,
					Labels:    testDefaultOpts().CommonLabels,
				},
				StringData: map[string]string{},
				Type:       v1.SecretTypeOpaque,
			},
			wantErr: false,
		},
		{
			name: "successfully create filled hashed secret",
			args: args{
				ctx:        context.TODO(),
				k8sclient:  fake.NewFakeClient(secrets...),
				secretRefs: secretRefs,
			},
			want: &v1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      HashedSecretName,
					Namespace: testDefaultOpts().Namespace,
					Labels:    testDefaultOpts().CommonLabels,
				},
				StringData: map[string]string{
					testAPIcastEmbeddedConfigurationSecretName: "aeefb14e2600c4b611b71955d84f3f79db8a399ea092a7baaed922ab37f95012",
					testCustomEnvironmentSecretName:            "a37c48fa15cb1fe1d8b656fc385899664c7ff23b99d95d53c7784daefef9656b",
				},
				Type: v1.SecretTypeOpaque,
			},
			wantErr: false,
		},
		{
			name: "fail to create hashed secret if missing source secrets",
			args: args{
				ctx:        context.TODO(),
				k8sclient:  fake.NewFakeClient(),
				secretRefs: secretRefs,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &APIcast{
				options: testDefaultOpts(),
			}
			got, err := a.HashedSecret(tt.args.ctx, tt.args.k8sclient, tt.args.secretRefs)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashedSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HashedSecret() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func testAPIcastEmbeddedConfigurationSecret() *v1.Secret {
	embeddedConfigurationContent := `{
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
	embeddedConfigSecret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testAPIcastEmbeddedConfigurationSecretName,
			Namespace: testNamespace,
			Labels: map[string]string{
				"apicast.apps.3scale.net/watched-by": "apicast",
			},
		},
		Data: map[string][]byte{
			"config.json": []byte(embeddedConfigurationContent),
		},
	}

	return &embeddedConfigSecret
}

func testCustomEnvironmentSecret() *v1.Secret {
	customEnvironmentContent := `
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
	customEnvironmentSecret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testCustomEnvironmentSecretName,
			Namespace: testNamespace,
			Labels: map[string]string{
				"apicast.apps.3scale.net/watched-by": "apicast",
			},
		},
		Data: map[string][]byte{
			"custom_env.lua": []byte(customEnvironmentContent),
		},
	}

	return &customEnvironmentSecret
}
