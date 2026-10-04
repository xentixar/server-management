package helpers

import (
	"regexp"
	"strings"
	"time"
)

var hostHostnameRe = regexp.MustCompile(`(?i)^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*$`)

func IsValidHostHostname(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 253 {
		return false
	}
	return hostHostnameRe.MatchString(name)
}

func NormalizeTimezone(tz string) string {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		return "UTC"
	}
	return tz
}

func IsValidTimezone(tz string) bool {
	tz = NormalizeTimezone(tz)
	_, err := time.LoadLocation(tz)
	return err == nil
}
