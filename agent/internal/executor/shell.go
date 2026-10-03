package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func run(ctx context.Context, name string, args ...string) (string, error) {
	return runEnv(ctx, nil, name, args...)
}

func runEnv(ctx context.Context, extraEnv []string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	env := os.Environ()
	// systemd agent units often omit HOME; Composer/fnm need it.
	if os.Getenv("HOME") == "" {
		home := "/root"
		if h, err := os.UserHomeDir(); err == nil && h != "" {
			home = h
		}
		env = append(env, "HOME="+home, "COMPOSER_HOME="+home+"/.composer")
	}
	if len(extraEnv) > 0 {
		env = append(env, extraEnv...)
	}
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := strings.TrimSpace(stdout.String())
	if err != nil {
		// apt and friends often put the useful failure on stdout; keep both.
		msg := strings.TrimSpace(stderr.String())
		switch {
		case out != "" && msg != "":
			msg = out + "\n" + msg
		case out != "":
			msg = out
		case msg == "":
			msg = err.Error()
		}
		msg = trimTail(msg, 4000)
		label := name
		if name != "bash" {
			label = name + " " + strings.Join(args, " ")
		}
		return out, fmt.Errorf("%s: %s", label, msg)
	}
	return out, nil
}

func trimTail(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "…\n" + s[len(s)-max:]
}

func runShell(ctx context.Context, script string) (string, error) {
	return runShellEnv(ctx, nil, script)
}

func runShellEnv(ctx context.Context, extraEnv []string, script string) (string, error) {
	// bash -c (not -lc): login profiles fight with noninteractive agent env (fnm/HOME).
	return runEnv(ctx, extraEnv, "bash", "-c", script)
}

func hasCommand(name string) bool {
	return findBinary(name) != ""
}

func findBinary(name string) string {
	if name == "" {
		return ""
	}
	if filepath.IsAbs(name) {
		if st, err := os.Stat(name); err == nil && !st.IsDir() {
			return name
		}
		return ""
	}
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	for _, dir := range []string{"/usr/sbin", "/sbin", "/usr/local/sbin", "/usr/bin", "/bin", "/usr/local/bin"} {
		p := filepath.Join(dir, name)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}
