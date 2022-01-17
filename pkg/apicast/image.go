package apicast

import "github.com/3scale/apicast-operator/pkg/helper"

const defaultImageVersion = "quay.io/3scale/3scale212:apicast-3scale-2.12.0-GA"

func GetDefaultImageVersion() string {
	return helper.GetEnvVar("RELATED_IMAGE_APICAST", defaultImageVersion)
}
