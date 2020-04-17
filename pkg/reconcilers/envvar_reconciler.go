package reconcilers

import (
	"reflect"

	"github.com/3scale/apicast-operator/pkg/k8sutils"
	v1 "k8s.io/api/core/v1"
)

// ReconcileEnvVar reconciles environment var lists
func ReconcileEnvVar(existing *[]v1.EnvVar, desired []v1.EnvVar) bool {
	if *existing == nil {
		*existing = []v1.EnvVar{}
	}

	if len(*existing) != len(desired) {
		*existing = desired
		return true
	}

	for _, desiredEnvVar := range desired {
		if idx := k8sutils.FindEnvVar(*existing, desiredEnvVar.Name); idx < 0 || !reflect.DeepEqual((*existing)[idx], desiredEnvVar) {
			*existing = desired
			return true
		}
	}

	return false
}
