package executor

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"agent/internal/protocol"
)

var hostHostnameRe = regexp.MustCompile(`(?i)^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*$`)

func handleServerSettingsSync(ctx context.Context, job protocol.Job) (string, map[string]any, error) {
	hostname := payloadString(job.Payload, "hostname")
	timezone := payloadString(job.Payload, "timezone")
	if timezone == "" {
		timezone = "UTC"
	}

	if !isValidHostHostname(hostname) {
		return "", nil, fmt.Errorf("invalid hostname")
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return "", nil, fmt.Errorf("invalid timezone")
	}
	if strings.ContainsAny(timezone, " \t\n\r;&|`$") {
		return "", nil, fmt.Errorf("invalid timezone")
	}

	if !hasCommand("hostnamectl") {
		return "", nil, fmt.Errorf("hostnamectl not found")
	}
	if !hasCommand("timedatectl") {
		return "", nil, fmt.Errorf("timedatectl not found")
	}

	if _, err := run(ctx, "hostnamectl", "set-hostname", hostname); err != nil {
		return "", nil, err
	}
	if _, err := run(ctx, "timedatectl", "set-timezone", timezone); err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("applied hostname %s timezone %s", hostname, timezone), nil, nil
}

func isValidHostHostname(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 253 {
		return false
	}
	return hostHostnameRe.MatchString(name)
}
