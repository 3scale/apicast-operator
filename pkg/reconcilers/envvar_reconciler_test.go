package reconcilers

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestReconcileEnvVars(t *testing.T) {
	cases := []struct {
		name           string
		existing       []v1.EnvVar
		desired        []v1.EnvVar
		expectedResult bool
		expectedEnvs   []v1.EnvVar
	}{
		{
			name:     "Nil existing",
			existing: nil,
			desired: []v1.EnvVar{
				{Name: "new_var", Value: "value"},
			},
			expectedResult: true,
			expectedEnvs: []v1.EnvVar{
				{Name: "new_var", Value: "value"},
			},
		},
		{
			name:           "No changes when empty",
			existing:       []v1.EnvVar{},
			desired:        []v1.EnvVar{},
			expectedResult: false,
			expectedEnvs:   []v1.EnvVar{},
		},
		{
			name:           "No changes when identical",
			existing:       []v1.EnvVar{{Name: "foo", Value: "bar"}},
			desired:        []v1.EnvVar{{Name: "foo", Value: "bar"}},
			expectedResult: false,
			expectedEnvs:   []v1.EnvVar{{Name: "foo", Value: "bar"}},
		},
		{
			name:           "Add new EnvVar",
			existing:       []v1.EnvVar{},
			desired:        []v1.EnvVar{{Name: "foo", Value: "bar"}},
			expectedResult: true,
			expectedEnvs:   []v1.EnvVar{{Name: "foo", Value: "bar"}},
		},
		{
			name:           "Update existing EnvVar",
			existing:       []v1.EnvVar{{Name: "foo", Value: "old"}},
			desired:        []v1.EnvVar{{Name: "foo", Value: "new"}},
			expectedResult: true,
			expectedEnvs:   []v1.EnvVar{{Name: "foo", Value: "new"}},
		},
		{
			name: "Preserve custom EnvVars",
			existing: []v1.EnvVar{
				{Name: "foo", Value: "bar"},
				{Name: "custom", Value: "custom_value"},
			},
			desired: []v1.EnvVar{
				{Name: "foo", Value: "bar"},
			},
			expectedResult: true,
			expectedEnvs: []v1.EnvVar{
				{Name: "foo", Value: "bar"},
				{Name: "custom", Value: "custom_value"},
			},
		},
		{
			name: "Add and preserve",
			existing: []v1.EnvVar{
				{Name: "existing", Value: "old_value"},
				{Name: "custom", Value: "custom_value"},
			},
			desired: []v1.EnvVar{
				{Name: "existing", Value: "new_value"},
				{Name: "new_var", Value: "new_value"},
			},
			expectedResult: true,
			expectedEnvs: []v1.EnvVar{
				{Name: "existing", Value: "new_value"},
				{Name: "new_var", Value: "new_value"},
				{Name: "custom", Value: "custom_value"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			update := ReconcileEnvVars(&tc.existing, tc.desired)

			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}

			// Check if the resulting env vars match expected
			if !reflect.DeepEqual(tc.existing, tc.expectedEnvs) {
				subT.Fatalf("env vars mismatch:\nexpected: %+v\ngot:      %+v", tc.expectedEnvs, tc.existing)
			}
		})
	}
}
