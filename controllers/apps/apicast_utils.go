package controllers

import (
	"fmt"
	"reflect"
	"strings"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
)

const (
	APIcastSecretLabelPrefix = "secret.apicast.apps.3scale.net/"
)

func apicastSecretLabelKey(uid string) string {
	return fmt.Sprintf("%s%s", APIcastSecretLabelPrefix, uid)
}

func replaceAPIcastSecretLabels(apicast *appsv1alpha1.APIcast, desiredSecretUIDs map[string]string) bool {
	existingLabels := apicast.GetLabels()

	if existingLabels == nil {
		existingLabels = map[string]string{}
	}

	existingSecretLabels := map[string]string{}

	// existing Secret UIDs not included in desiredAPIUIDs are deleted
	for key, value := range existingLabels {
		if strings.HasPrefix(key, APIcastSecretLabelPrefix) {
			existingSecretLabels[key] = value
			// it is safe to remove keys while looping in range
			delete(existingLabels, key)
		}
	}

	desiredSecretLabels := map[string]string{}
	for uid, watchedByStatus := range desiredSecretUIDs {
		desiredSecretLabels[apicastSecretLabelKey(uid)] = watchedByStatus
		existingLabels[apicastSecretLabelKey(uid)] = watchedByStatus
	}

	apicast.SetLabels(existingLabels)

	return !reflect.DeepEqual(existingSecretLabels, desiredSecretLabels)
}
