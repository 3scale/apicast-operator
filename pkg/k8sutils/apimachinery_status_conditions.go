package k8sutils

import (
	"encoding/json"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CopyConditions copies the set of conditions
func CopyConditions(conditions []metav1.Condition) []metav1.Condition {
	newConditions := append([]metav1.Condition{}, conditions...)
	return newConditions
}

// ConditionMarshal marshals the set of conditions as a JSON array, sorted by condition type.
func ConditionMarshal(conditions []metav1.Condition) ([]byte, error) {
	var condCopy []metav1.Condition
	condCopy = append(condCopy, conditions...)
	sort.Slice(condCopy, func(a, b int) bool {
		return condCopy[a].Type < condCopy[b].Type
	})
	return json.Marshal(condCopy)
}
