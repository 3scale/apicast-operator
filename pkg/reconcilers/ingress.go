package reconcilers

import (
	"fmt"
	"reflect"

	networkingv1 "k8s.io/api/networking/v1"

	"github.com/3scale/apicast-operator/pkg/k8sutils"
)

func IngressMutator(existingObj, desiredObj k8sutils.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*networkingv1.Ingress)
	if !ok {
		return false, fmt.Errorf("%T is not a *networkingv1.Ingress", existingObj)
	}
	desired, ok := desiredObj.(*networkingv1.Ingress)
	if !ok {
		return false, fmt.Errorf("%T is not a *networkingv1.Ingress", desiredObj)
	}

	exposedHostIdx := -1
	exposedHost := desired.Spec.Rules[0].Host
	for idx, rule := range existing.Spec.Rules {
		if rule.Host == exposedHost {
			exposedHostIdx = idx
		}
	}

	update := false

	if !reflect.DeepEqual(existing.Spec.IngressClassName, desired.Spec.IngressClassName) {
		existing.Spec.IngressClassName = desired.Spec.IngressClassName
		update = true
	}

	if exposedHostIdx == -1 {
		existing.Spec.Rules = desired.Spec.Rules
		update = true
	}

	if !reflect.DeepEqual(existing.Spec.TLS, desired.Spec.TLS) {
		existing.Spec.TLS = desired.Spec.TLS
		update = true
	}

	return update, nil
}
