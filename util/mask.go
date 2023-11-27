package util

import (
	"os"
	"strconv"
	"strings"
)

// MaskEmail masks every character of the email
// leaving the first character of the username and host and top-level domain visible
// Eg. foo@bar.bar.com.sg ->f*@b*.sg
// Eg. f@b.com -> f*@b*.com
func MaskEmail(email string) string {
	matches := strings.Split(email, "@")
	if len(matches) != 2 || len(matches[0]) == 0 || len(matches[1]) == 0 {
		return email
	}
	username := matches[0]
	provider := matches[1]

	lastDotIndex := strings.LastIndex(provider, ".")
	if lastDotIndex <= 0 {
		return email // Not a valid email
	}
	topLevelDomain := provider[lastDotIndex:]

	return username[:1] + "*" + "@" + provider[:1] + "*" + topLevelDomain
}

// MaskMobile masks the front of the mobile number,
// leaving the back digits visible
// Eg. 9123456789 -> *******789
func MaskMobile(mobile string) string {
	maskLength, _ := strconv.Atoi(os.Getenv("MOBILE_MASK_VISIBLE_LENGTH"))
	if len(mobile) < maskLength {
		return mobile
	}
	return strings.Repeat("*", len(mobile)-maskLength) + mobile[len(mobile)-maskLength:]
}
