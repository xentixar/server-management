package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"agent/internal/protocol"
)

var serviceNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

const supervisorConfDir = "/etc/supervisor/conf.d"
const supervisorLogDir = "/var/log/supervisor"

func programName(name string) string {
	return managedProgramName(name)
}

func programConfPath(name string) string {
	return managedProgramConfPath(name)
}

func ensureSupervisor(ctx context.Context) error {
	if !hasCommand("apt-get") {
		return fmt.Errorf("ubuntu/debian apt-get required")
	}
	if !hasCommand("supervisorctl") {
		env := "DEBIAN_FRONTEND=noninteractive"
		if _, err := runShell(ctx, env+" apt-get update -y"); err != nil {
			return err
		}
		if _, err := runShell(ctx, env+" apt-get install -y supervisor"); err != nil {
			return err
		}
	}
	_ = os.MkdirAll(supervisorConfDir, 0o755)
	_ = os.MkdirAll(supervisorLogDir, 0o755)
	if hasCommand("systemctl") {
		_, _ = run(ctx, "systemctl", "enable", "--now", "supervisor")
	}
	return nil
}

func supervisorRereadUpdate(ctx context.Context) error {
	if _, err := run(ctx, "supervisorctl", "reread"); err != nil {
		return err
	}
	_, err := run(ctx, "supervisorctl", "update")
	return err
}

func writeProgramConf(name, command, directory, user string, processes int, autostart, autorestart, enabled bool) error {
	if processes < 1 {
		processes = 1
	}
	prog := programName(name)
	auto := autostart && enabled
	var b strings.Builder
	fmt.Fprintf(&b, "[program:%s]\n", prog)
	fmt.Fprintf(&b, "command=%s\n", command)
	if directory != "" {
		fmt.Fprintf(&b, "directory=%s\n", directory)
	}
	fmt.Fprintf(&b, "user=%s\n", user)
	fmt.Fprintf(&b, "numprocs=%d\n", processes)
	fmt.Fprintf(&b, "autostart=%v\n", auto)
	fmt.Fprintf(&b, "autorestart=%v\n", autorestart && enabled)
	fmt.Fprintf(&b, "startsecs=1\n")
	fmt.Fprintf(&b, "stopwaitsecs=10\n")
	fmt.Fprintf(&b, "stdout_logfile=%s/%s.out.log\n", supervisorLogDir, prog)
	fmt.Fprintf(&b, "stderr_logfile=%s/%s.err.log\n", supervisorLogDir, prog)
	fmt.Fprintf(&b, "stdout_logfile_maxbytes=2MB\n")
	fmt.Fprintf(&b, "stderr_logfile_maxbytes=2MB\n")
	return os.WriteFile(programConfPath(name), []byte(b.String()), 0o644)
}

func parseSupervisorStatus(out string) string {
	u := strings.ToUpper(out)
	switch {
	case strings.Contains(u, "RUNNING"):
		return "running"
	case strings.Contains(u, "STARTING"), strings.Contains(u, "STOPPING"):
		return "unknown"
	case strings.Contains(u, "FATAL"), strings.Contains(u, "BACKOFF"), strings.Contains(u, "EXITED"):
		if strings.Contains(u, "EXITED") && !strings.Contains(u, "FATAL") {
			return "stopped"
		}
		if strings.Contains(u, "FATAL") || strings.Contains(u, "BACKOFF") {
			return "failed"
		}
		return "stopped"
	case strings.Contains(u, "STOPPED"):
		return "stopped"
	default:
		return "unknown"
	}
}

func supervisorStatus(ctx context.Context, name string) (string, string) {
	out, err := run(ctx, "supervisorctl", "status", programName(name))
	if err != nil {
		// supervisorctl returns non-zero for STOPPED/FATAL sometimes
		if out == "" {
			return "unknown", err.Error()
		}
	}
	return parseSupervisorStatus(out), out
}

func supervisorLogPath(name, stream string) string {
	ext := "out"
	if stream == "stderr" {
		ext = "err"
	}
	return filepath.Join(supervisorLogDir, programName(name)+"."+ext+".log")
}

func supervisorTail(ctx context.Context, name, stream string) string {
	path := supervisorLogPath(name, stream)
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	out, err := run(ctx, "tail", "-n", "200", path)
	if err != nil {
		return trimTail(out, 16384)
	}
	return trimTail(out, 16384)
}

func serviceResultMeta(ctx context.Context, name string) map[string]any {
	status, msg := supervisorStatus(ctx, name)
	return map[string]any{
		"status":      status,
		"stdout_tail": supervisorTail(ctx, name, "stdout"),
		"stderr_tail": supervisorTail(ctx, name, "stderr"),
		"detail":      msg,
	}
}

func handleServiceApply(ctx context.Context, job protocol.Job) (string, map[string]any, error) {
	name := payloadString(job.Payload, "name")
	command := payloadString(job.Payload, "command")
	directory := payloadString(job.Payload, "directory")
	user := defaultProcessUser(payloadString(job.Payload, "user"))
	processes := payloadInt(job.Payload, "processes", 1)
	autoRestart := payloadBool(job.Payload, "auto_restart", true)
	autostart := payloadBool(job.Payload, "autostart", true)
	enabled := payloadBool(job.Payload, "enabled", true)

	if !serviceNameRe.MatchString(name) {
		return "", nil, fmt.Errorf("invalid service name")
	}
	if !isValidProcessCommand(command) {
		return "", nil, fmt.Errorf("invalid command")
	}
	if err := ensureSupervisor(ctx); err != nil {
		return "", nil, err
	}
	if err := writeProgramConf(name, command, directory, user, processes, autostart, autoRestart, enabled); err != nil {
		return "", nil, err
	}
	if err := supervisorRereadUpdate(ctx); err != nil {
		return "", nil, err
	}
	prog := programName(name)
	if enabled && autostart {
		_, _ = run(ctx, "supervisorctl", "start", prog)
	} else {
		_, _ = run(ctx, "supervisorctl", "stop", prog)
	}
	meta := serviceResultMeta(ctx, name)
	return fmt.Sprintf("applied %s (%s)", prog, meta["status"]), meta, nil
}

func handleServiceRemove(ctx context.Context, job protocol.Job) (string, map[string]any, error) {
	name := payloadString(job.Payload, "name")
	if !serviceNameRe.MatchString(name) {
		return "", nil, fmt.Errorf("invalid service name")
	}
	if err := ensureSupervisor(ctx); err != nil {
		return "", nil, err
	}
	prog := programName(name)
	_, _ = run(ctx, "supervisorctl", "stop", prog)
	_ = os.Remove(programConfPath(name))
	if err := supervisorRereadUpdate(ctx); err != nil {
		return "", nil, err
	}
	return "removed " + prog, map[string]any{"status": "stopped"}, nil
}

func handleServiceControl(ctx context.Context, job protocol.Job) (string, map[string]any, error) {
	name := payloadString(job.Payload, "name")
	action := strings.ToLower(payloadString(job.Payload, "action"))
	if !serviceNameRe.MatchString(name) {
		return "", nil, fmt.Errorf("invalid service name")
	}
	switch action {
	case "start", "stop", "restart", "status":
	default:
		return "", nil, fmt.Errorf("invalid action %q", action)
	}
	if err := ensureSupervisor(ctx); err != nil {
		return "", nil, err
	}
	prog := programName(name)
	if action != "status" {
		if _, err := run(ctx, "supervisorctl", action, prog); err != nil {
			meta := serviceResultMeta(ctx, name)
			return "", meta, err
		}
	}
	meta := serviceResultMeta(ctx, name)
	return fmt.Sprintf("%s %s (%s)", action, prog, meta["status"]), meta, nil
}
