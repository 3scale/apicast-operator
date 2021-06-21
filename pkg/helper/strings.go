package helper

import (
	"regexp"
	"strings"
)

var (
	// InvalidDNS1123Regexp not alphanumeric
	InvalidDNS1123Regexp = regexp.MustCompile(`[^0-9A-Za-z-]`)
)

func DNS1123Name(in string) string {
	tmp := strings.ToLower(in)
	return InvalidDNS1123Regexp.ReplaceAllString(tmp, "")
}
