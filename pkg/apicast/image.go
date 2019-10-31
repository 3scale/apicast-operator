package apicast

import "github.com/3scale/apicast-operator/pkg/helper"

const defaultImageVersion = "quay.io/3scale/apicast:nightly"

func GetDefaultImageVersion() string {
	return helper.GetEnvVar("APICAST_IMAGE", defaultImageVersion)
}
