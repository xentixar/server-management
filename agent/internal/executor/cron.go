package executor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"agent/internal/protocol"
)

func handleCronSync(ctx context.Context, job protocol.Job) (string, map[string]any, error) {
	entries, err := cronEntriesFromPayload(job.Payload)
	if err != nil {
		return "", nil, err
	}

	var b strings.Builder
	b.WriteString("# Managed by server-agent - do not edit\n")
	b.WriteString("SHELL=/bin/bash\n")
	b.WriteString("PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\n")
	b.WriteString("\n")

	for _, e := range entries {
		if !isValidCronSchedule(e.schedule) {
			return "", nil, fmt.Errorf("invalid schedule for %s", e.uid)
		}
		if !isValidProcessCommand(e.command) {
			return "", nil, fmt.Errorf("invalid command for %s", e.uid)
		}
		if _, err := time.LoadLocation(e.timezone); err != nil {
			return "", nil, fmt.Errorf("invalid timezone for %s", e.uid)
		}
		user := defaultProcessUser(e.user)
		if strings.ContainsAny(user, " \t\n\r") {
			return "", nil, fmt.Errorf("invalid user for %s", e.uid)
		}
		fmt.Fprintf(&b, "# %s\n", e.uid)
		fmt.Fprintf(&b, "CRON_TZ=%s\n", e.timezone)
		fmt.Fprintf(&b, "%s %s %s\n\n", e.schedule, user, e.command)
	}

	if err := os.WriteFile(managedCronFilePath(), []byte(b.String()), 0o644); err != nil {
		return "", nil, err
	}
	if hasCommand("systemctl") {
		_, _ = run(ctx, "systemctl", "reload", "cron")
		_, _ = run(ctx, "systemctl", "reload", "crond")
	}
	return fmt.Sprintf("synced %d cron entries", len(entries)), nil, nil
}

type cronEntry struct {
	uid      string
	schedule string
	command  string
	user     string
	timezone string
}

func cronEntriesFromPayload(payload map[string]any) ([]cronEntry, error) {
	if payload == nil {
		return nil, nil
	}
	raw, ok := payload["entries"]
	if !ok || raw == nil {
		return nil, nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("entries must be a list")
	}
	out := make([]cronEntry, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		enabled := payloadBool(m, "enabled", true)
		if !enabled {
			continue
		}
		tz := payloadString(m, "timezone")
		if tz == "" {
			tz = "UTC"
		}
		out = append(out, cronEntry{
			uid:      payloadString(m, "uid"),
			schedule: payloadString(m, "schedule"),
			command:  payloadString(m, "command"),
			user:     payloadString(m, "user"),
			timezone: tz,
		})
	}
	return out, nil
}
