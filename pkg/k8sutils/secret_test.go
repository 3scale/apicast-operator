package k8sutils

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIsSecretWatchedByApicast(t *testing.T) {
	labeledSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "labeled-secret",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"apicast.apps.3scale.net/watched-by": "apicast",
			},
		},
	}
	unlabeledSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unlabeled-secret",
			Namespace: "test-namespace",
		},
	}

	type args struct {
		client          k8sclient.Client
		secretName      string
		secretNamespace string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Secret doesn't have watched-by label",
			args: args{
				client:          fake.NewFakeClient(unlabeledSecret),
				secretName:      "unlabeled-secret",
				secretNamespace: "test-namespace",
			},
			want: false,
		},
		{
			name: "Secret has watched-by label",
			args: args{
				client:          fake.NewFakeClient(labeledSecret),
				secretName:      "labeled-secret",
				secretNamespace: "test-namespace",
			},
			want: true,
		},
		{
			name: "Secret doesn't exist",
			args: args{
				client:          fake.NewFakeClient(),
				secretName:      "unlabeled-secret",
				secretNamespace: "test-namespace",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSecretWatchedByApicast(tt.args.client, tt.args.secretName, tt.args.secretNamespace); got != tt.want {
				t.Errorf("IsSecretWatchedByApicast() = %v, want %v", got, tt.want)
			}
		})
	}
}
