package executor

import (
	"context"
	"fmt"
	"strings"
)

const (
	managedUnixUser       = "sm"
	managedResourcePrefix = "sa"
)

func managedProgramName(name string) string {
	return managedResourcePrefix + "-" + name
}

func managedProgramConfPath(name string) string {
	return supervisorConfDir + "/" + managedProgramName(name) + ".conf"
}

func managedCronFilePath() string {
	return "/etc/cron.d/" + managedResourcePrefix
}

func ensureManagedUser(ctx context.Context) (map[string]string, error) {
	pass, err := randomPassword(24)
	if err != nil {
		return nil, err
	}
	env := "DEBIAN_FRONTEND=noninteractive"
	if _, err := runShell(ctx, env+" apt-get install -y sudo"); err != nil {
		return nil, fmt.Errorf("install sudo: %w", err)
	}

	script := `set -euo pipefail
USER_NAME="$SM_USER"
PASS="$SM_PASS"
if ! id -u "$USER_NAME" >/dev/null 2>&1; then
  useradd -m -s /bin/bash "$USER_NAME"
fi
echo "${USER_NAME}:${PASS}" | chpasswd
usermod -aG sudo "$USER_NAME"
`
	if _, err := runShellEnv(ctx, []string{
		"SM_USER=" + managedUnixUser,
		"SM_PASS=" + pass,
	}, script); err != nil {
		return nil, err
	}

	return map[string]string{
		"user":     managedUnixUser,
		"password": pass,
	}, nil
}

func defaultProcessUser(payloadUser string) string {
	u := strings.TrimSpace(payloadUser)
	if u == "" {
		return managedUnixUser
	}
	return u
}
