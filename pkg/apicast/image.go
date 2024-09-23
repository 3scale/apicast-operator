package apicast

import "github.com/3scale/apicast-operator/pkg/helper"

const defaultImageVersion = "quay.io/3scale/apicast:v3.15.0"

func GetDefaultImageVersion() string {
	return helper.GetEnvVar("RELATED_IMAGE_APICAST", defaultImageVersion)
}
