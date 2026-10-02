package executor

import (
	"context"
	"fmt"
	"strings"

	"agent/internal/protocol"
)

func handleRuntimePHPInstall(ctx context.Context, job protocol.Job) (string, map[string]any, error) {
	if !hasCommand("apt-get") {
		return "", nil, fmt.Errorf("ubuntu/debian apt-get required")
	}
	version, _ := job.Payload["version"].(string)
	pkgs, err := phpPackages(version)
	if err != nil {
		return "", nil, err
	}
	if err := ensureSuryPHP(ctx); err != nil {
		return "", nil, err
	}
	env := "DEBIAN_FRONTEND=noninteractive"
	if _, err := runShell(ctx, env+" apt-get install -y "+strings.Join(pkgs, " ")); err != nil {
		return "", nil, err
	}
	_ = enableService(ctx, "php"+version+"-fpm")
	return "installed php " + version, nil, nil
}

func handleRuntimeNodeInstall(ctx context.Context, job protocol.Job) (string, map[string]any, error) {
	version, _ := job.Payload["version"].(string)
	if err := ensureFnm(ctx); err != nil {
		return "", nil, err
	}
	if err := installNodeVersion(ctx, version); err != nil {
		return "", nil, err
	}
	return "installed node " + version, nil, nil
}

func handleRuntimeRemove(_ context.Context, job protocol.Job) (string, map[string]any, error) {
	// ponytail: remove not wired for v1 UI; fail closed until DELETE is implemented
	return "", nil, fmt.Errorf("unsupported job kind %q", job.Kind)
}
