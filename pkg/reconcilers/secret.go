package reconcilers

import (
	"fmt"
	"reflect"

	"github.com/3scale/apicast-operator/pkg/k8sutils"
	v1 "k8s.io/api/core/v1"
)

// SecretMutateFn is a function which mutates the existing Secret into it's desired state.
type SecretMutateFn func(desired, existing *v1.Secret) bool

func SecretMutator(opts ...SecretMutateFn) MutateFn {
	return func(existingObj, desiredObj k8sutils.KubernetesObject) (bool, error) {
		existing, ok := existingObj.(*v1.Secret)
		if !ok {
			return false, fmt.Errorf("%T is not a *v1.Secret", existingObj)
		}
		desired, ok := desiredObj.(*v1.Secret)
		if !ok {
			return false, fmt.Errorf("%T is not a *v1.Secret", desiredObj)
		}

		update := false

		// Loop through each option
		for _, opt := range opts {
			tmpUpdate := opt(desired, existing)
			update = update || tmpUpdate
		}

		return update, nil
	}
}

func SecretStringDataMutator(desired, existing *v1.Secret) bool {
	updated := false

	// StringData is merged to Data on write, so we need to compare the existing Data to the desired StringData
	// Before we can do this we need to convert the existing Data to StringData
	existingStringData := make(map[string]string)
	for key, bytes := range existing.Data {
		existingStringData[key] = string(bytes)
	}
	if !reflect.DeepEqual(existingStringData, desired.StringData) {
		updated = true
		existing.Data = nil // Need to clear the existing.Data because of how StringData is converted to Data
		existing.StringData = desired.StringData
	}

	return updated
}
