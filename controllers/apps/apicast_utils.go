package controllers

import (
	"fmt"
	"reflect"
	"strings"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
)

const (
	APIcastSecretLabelPrefix = "secret.apicast.apps.3scale.net/"
	APIcastSecretLabelValue  = "true"
)

func apicastSecretLabelKey(uid string) string {
	return fmt.Sprintf("%s%s", APIcastSecretLabelPrefix, uid)
}

func replaceAPIcastSecretLabels(apicast *appsv1alpha1.APIcast, desiredSecretUIDs []string) bool {
	existingLabels := apicast.GetLabels()

	if existingLabels == nil {
		existingLabels = map[string]string{}
	}

	existingSecretLabels := map[string]string{}

	// existing Secret UIDs not included in desiredAPIUIDs are deleted
	for k := range existingLabels {
		if strings.HasPrefix(k, APIcastSecretLabelPrefix) {
			existingSecretLabels[k] = APIcastSecretLabelValue
			// it is safe to remove keys while looping in range
			delete(existingLabels, k)
		}
	}

	desiredSecretLabels := map[string]string{}
	for _, uid := range desiredSecretUIDs {
		desiredSecretLabels[apicastSecretLabelKey(uid)] = APIcastSecretLabelValue
		existingLabels[apicastSecretLabelKey(uid)] = APIcastSecretLabelValue
	}

	apicast.SetLabels(existingLabels)

	return !reflect.DeepEqual(existingSecretLabels, desiredSecretLabels)
}
