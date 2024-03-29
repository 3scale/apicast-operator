package reconcilers

import (
	"fmt"
	"reflect"

	"github.com/3scale/apicast-operator/pkg/k8sutils"
	v1 "k8s.io/api/core/v1"
)

// ServiceMutateFn is a function which mutates the existing Service into it's desired state.
type ServiceMutateFn func(desired, existing *v1.Service) bool

func ServiceMutator(opts ...ServiceMutateFn) MutateFn {
	return func(existingObj, desiredObj k8sutils.KubernetesObject) (bool, error) {
		existing, ok := existingObj.(*v1.Service)
		if !ok {
			return false, fmt.Errorf("%T is not a *v1.Service", existingObj)
		}
		desired, ok := desiredObj.(*v1.Service)
		if !ok {
			return false, fmt.Errorf("%T is not a *appsv1.Service", desiredObj)
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

func ServicePortMutator(desired, existing *v1.Service) bool {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Ports, desired.Spec.Ports) {
		updated = true
		existing.Spec.Ports = desired.Spec.Ports
	}

	return updated
}

func ServiceSelectorMutator(desired, existing *v1.Service) bool {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Selector, desired.Spec.Selector) {
		updated = true
		existing.Spec.Selector = desired.Spec.Selector
	}

	return updated
}
