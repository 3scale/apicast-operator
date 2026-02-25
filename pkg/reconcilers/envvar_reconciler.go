package reconcilers

import (
	"reflect"

	v1 "k8s.io/api/core/v1"
)

// ReconcileEnvVar reconciles a complete list of environment variables.
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Removed when not in desired and exists in existing
func ReconcileEnvVars(existing *[]v1.EnvVar, desired []v1.EnvVar) bool {
	if existing == nil {
		*existing = make([]v1.EnvVar, 0, len(desired))
	}

	// Build maps for fast lookups
	existingMap := make(map[string]int, len(*existing))
	for i, ev := range *existing {
		existingMap[ev.Name] = i
	}

	desiredMap := make(map[string]v1.EnvVar, len(desired))
	for _, ev := range desired {
		desiredMap[ev.Name] = ev
	}

	update := false
	result := make([]v1.EnvVar, 0, len(desired))

	for _, desiredVar := range desired {
		if existingIdx, found := existingMap[desiredVar.Name]; found {
			// Exists - check if update needed
			if !reflect.DeepEqual((*existing)[existingIdx], desiredVar) {
				result = append(result, desiredVar)
				update = true
			} else {
				result = append(result, (*existing)[existingIdx])
			}
		} else {
			// New var - add it
			result = append(result, desiredVar)
			update = true
		}
	}

	// Check for custom vars (in existing but not in desired)
	for _, existingVar := range *existing {
		if _, found := desiredMap[existingVar.Name]; !found {
			result = append(result, existingVar)
			update = true
		}
	}

	if update {
		*existing = result
	}

	return update
}
