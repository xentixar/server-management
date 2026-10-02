package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if os.Getenv("HOME") == "" {
		home := "/root"
		if h, err := os.UserHomeDir(); err == nil && h != "" {
			home = h
		}
		cmd.Env = append(os.Environ(), "HOME="+home, "COMPOSER_HOME="+home+"/.composer")
	}
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
	return run(ctx, "bash", "-lc", script)
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
