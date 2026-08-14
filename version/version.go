package version

import (
	"strings"
)

var (
	Version           = "0.99.99"
	threescaleRelease = "2.99.99"
)

func ThreescaleVersionMajorMinor() string {
	parts := strings.Split(threescaleRelease, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return ""
}

func ThreescaleVersionMajorMinorPatch() string {
	return threescaleRelease
}
