package apicast

const currentImageVersion = "quay.io/3scale/apicast:nightly"

func GetCurrentImageVersion() string {
	return currentImageVersion
}
