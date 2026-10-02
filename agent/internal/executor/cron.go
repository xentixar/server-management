package executor

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"agent/internal/protocol"
)

var cronUIDRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func managedLogDir() string {
	return "/var/log/" + managedResourcePrefix
}

func cronLogPath(uid string) string {
	return filepath.Join(managedLogDir(), "cron-"+uid+".log")
}

func ensureCronLogFile(uid, username string) error {
	if !cronUIDRe.MatchString(uid) {
		return fmt.Errorf("invalid cron uid")
	}
	if err := os.MkdirAll(managedLogDir(), 0o755); err != nil {
		return err
	}
	path := cronLogPath(uid)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	_ = f.Close()

	u, err := user.Lookup(username)
	if err != nil {
		return nil // file exists; ownership best-effort
	}
	uidNum, err1 := strconv.Atoi(u.Uid)
	gidNum, err2 := strconv.Atoi(u.Gid)
	if err1 == nil && err2 == nil {
		_ = os.Chown(path, uidNum, gidNum)
	}
	return nil
}

func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func cronLoggedCommand(uid, command string) string {
	path := cronLogPath(uid)
	inner := command + " >> " + path + " 2>&1"
	return "/bin/bash -c " + shellSingleQuote(inner)
}

func handleCronSync(ctx context.Context, job protocol.Job) (string, map[string]any, error) {
	entries, err := cronEntriesFromPayload(job.Payload)
	if err != nil {
		return "", nil, err
	}

	if err := os.MkdirAll(managedLogDir(), 0o755); err != nil {
		return "", nil, err
	}

	var b strings.Builder
	b.WriteString("# Managed by server-agent - do not edit\n")
	b.WriteString("SHELL=/bin/bash\n")
	b.WriteString("PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\n")
	b.WriteString("\n")

	for _, e := range entries {
		if !cronUIDRe.MatchString(e.uid) {
			return "", nil, fmt.Errorf("invalid uid for cron entry")
		}
		if !isValidCronSchedule(e.schedule) {
			return "", nil, fmt.Errorf("invalid schedule for %s", e.uid)
		}
		if !isValidProcessCommand(e.command) {
			return "", nil, fmt.Errorf("invalid command for %s", e.uid)
		}
		if _, err := time.LoadLocation(e.timezone); err != nil {
			return "", nil, fmt.Errorf("invalid timezone for %s", e.uid)
		}
		userName := defaultProcessUser(e.user)
		if strings.ContainsAny(userName, " \t\n\r") {
			return "", nil, fmt.Errorf("invalid user for %s", e.uid)
		}
		if err := ensureCronLogFile(e.uid, userName); err != nil {
			return "", nil, err
		}
		fmt.Fprintf(&b, "# %s\n", e.uid)
		fmt.Fprintf(&b, "CRON_TZ=%s\n", e.timezone)
		fmt.Fprintf(&b, "%s %s %s\n\n", e.schedule, userName, cronLoggedCommand(e.uid, e.command))
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

func handleCronLogs(ctx context.Context, job protocol.Job) (string, map[string]any, error) {
	uid := payloadString(job.Payload, "uid")
	if !cronUIDRe.MatchString(uid) {
		return "", nil, fmt.Errorf("invalid cron uid")
	}
	path := cronLogPath(uid)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "no log yet", map[string]any{"log_tail": "(empty)"}, nil
		}
		return "", nil, err
	}
	out, err := run(ctx, "tail", "-n", "100", path)
	if err != nil {
		return "", map[string]any{"log_tail": trimTail(out, 4096)}, err
	}
	tail := trimTail(out, 4096)
	if strings.TrimSpace(tail) == "" {
		tail = "(empty)"
	}
	return "ok", map[string]any{"log_tail": tail}, nil
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
