package helper

import (
	"github.com/3scale/apicast-operator/version"
)

type ComponentType string

const (
	ApplicationType    ComponentType = "application"
	InfrastructureType ComponentType = "infrastructure"
)

func MeteringLabels(componentType ComponentType) map[string]string {
	return map[string]string{
		"com.company":   "Red_Hat",
		"rht.prod_name": "Red_Hat_Integration",
		// It should be updated on release branch
		"rht.prod_ver":  "2021.Q4",
		"rht.comp":      "3scale_apicast",
		"rht.comp_ver":  version.ThreescaleRelease,
		"rht.subcomp":   "apicast",
		"rht.subcomp_t": string(componentType),
	}
}
