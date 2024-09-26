package k8sutils

import (
	"context"

	v1 "k8s.io/api/core/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ApicastSecretLabel is the label that secrets need to have in order to reconcile changes
	ApicastSecretLabel = "apicast.apps.3scale.net/watched-by=apicast"
)

func SecretStringDataFromData(secret *v1.Secret) map[string]string {
	stringData := map[string]string{}

	for k, v := range secret.Data {
		stringData[k] = string(v)
	}
	return stringData
}

func IsSecretWatchedByApicast(client k8sclient.Client, secretName, secretNamespace string) bool {
	secret := &v1.Secret{}
	objKey := k8sclient.ObjectKey{
		Name:      secretName,
		Namespace: secretNamespace,
	}

	err := client.Get(context.TODO(), objKey, secret)
	if err != nil {
		return false
	}

	existingLabels := secret.Labels

	if existingLabels != nil {
		if _, ok := existingLabels["apicast.apps.3scale.net/watched-by"]; ok {
			return true
		}
	}

	return false
}
