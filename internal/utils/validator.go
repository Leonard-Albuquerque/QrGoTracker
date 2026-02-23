package utils

import (
	"net/url"
	"regexp"
	"strings"
)

var shortCodePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,40}$`)

func ValidURL(u string) bool {
	parsed, err := url.Parse(strings.TrimSpace(u))
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	if parsed.Host == "" {
		return false
	}
	return true
}

func ValidShortCode(code string) bool {
	return shortCodePattern.MatchString(strings.TrimSpace(code))
}
