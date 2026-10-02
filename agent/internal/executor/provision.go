package executor

import (
	"context"
	"fmt"
	"strings"

	"agent/internal/protocol"
)

func handleServerProvision(ctx context.Context, job protocol.Job) (string, map[string]any, error) {
	if !hasCommand("apt-get") {
		return "", nil, fmt.Errorf("ubuntu/debian apt-get required")
	}

	phpList := payloadStringSlice(job.Payload, "php_versions")
	nodeList := payloadStringSlice(job.Payload, "node_versions")
	addons := payloadBoolMap(job.Payload, "addons")

	if len(phpList) == 0 {
		return "", nil, fmt.Errorf("php_versions required")
	}
	if len(nodeList) == 0 {
		return "", nil, fmt.Errorf("node_versions required")
	}
	if addons["mysql"] && addons["mariadb"] {
		return "", nil, fmt.Errorf("mysql and mariadb are mutually exclusive")
	}

	env := "DEBIAN_FRONTEND=noninteractive"
	if _, err := runShell(ctx, env+" apt-get update -y"); err != nil {
		return "", nil, err
	}
	if _, err := runShell(ctx, env+" apt-get install -y "+strings.Join(basePackages, " ")); err != nil {
		return "", nil, err
	}

	if err := ensureSuryPHP(ctx); err != nil {
		return "", nil, err
	}

	var phpPkgs []string
	for _, ver := range phpList {
		pkgs, err := phpPackages(ver)
		if err != nil {
			return "", nil, err
		}
		phpPkgs = append(phpPkgs, pkgs...)
	}
	if _, err := runShell(ctx, env+" apt-get install -y "+strings.Join(phpPkgs, " ")); err != nil {
		return "", nil, err
	}

	if err := installComposer(ctx); err != nil {
		return "", nil, err
	}
	if err := ensureFnm(ctx); err != nil {
		return "", nil, err
	}
	for _, ver := range nodeList {
		if err := installNodeVersion(ctx, ver); err != nil {
			return "", nil, err
		}
	}
	if def := highestVersion(nodeList); def != "" {
		if err := applyNodeDefault(ctx, def); err != nil {
			return "", nil, fmt.Errorf("node default: %w", err)
		}
	}
	if def := highestVersion(phpList); def != "" {
		if err := applyPHPDefault(ctx, def); err != nil {
			return "", nil, fmt.Errorf("php default: %w", err)
		}
	}

	secrets := map[string]any{}
	sysCred, err := ensureManagedUser(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("managed user: %w", err)
	}
	secrets["system"] = sysCred

	var enabledAddons []string
	for _, name := range []string{"redis", "mysql", "mariadb", "postgres"} {
		if !addons[name] {
			continue
		}
		pkgs := addonPackages[name]
		if _, err := runShell(ctx, env+" apt-get install -y "+strings.Join(pkgs, " ")); err != nil {
			return "", nil, err
		}
		if unit := addonServices[name]; unit != "" {
			_ = enableService(ctx, unit)
		}
		cred, err := configureAddon(ctx, name)
		if err != nil {
			return "", nil, fmt.Errorf("%s credentials: %w", name, err)
		}
		secrets[name] = cred
		enabledAddons = append(enabledAddons, name)
	}

	_ = enableService(ctx, "nginx")
	for _, ver := range phpList {
		_ = enableService(ctx, "php"+ver+"-fpm")
	}

	return fmt.Sprintf(
		"provisioned php=%s node=%s addons=%s",
		strings.Join(phpList, ","),
		strings.Join(nodeList, ","),
		strings.Join(enabledAddons, ","),
	), secrets, nil
}

// ensureSuryPHP adds packages.sury.org (works on Debian + Ubuntu; no software-properties-common).
func ensureSuryPHP(ctx context.Context) error {
	env := "DEBIAN_FRONTEND=noninteractive"
	script := `set -e
` + env + ` apt-get install -y lsb-release ca-certificates curl
curl -sSLo /tmp/debsuryorg-archive-keyring.deb https://packages.sury.org/debsuryorg-archive-keyring.deb
dpkg -i /tmp/debsuryorg-archive-keyring.deb
echo "deb [signed-by=/usr/share/keyrings/debsuryorg-archive-keyring.gpg] https://packages.sury.org/php/ $(lsb_release -sc) main" > /etc/apt/sources.list.d/php-sury.list
` + env + ` apt-get update -y`
	_, err := runShell(ctx, script)
	return err
}

func installComposer(ctx context.Context) error {
	if hasCommand("composer") {
		return nil
	}
	script := `set -e
curl -sS https://getcomposer.org/installer -o /tmp/composer-setup.php
php /tmp/composer-setup.php --install-dir=/usr/local/bin --filename=composer
rm -f /tmp/composer-setup.php`
	_, err := runShell(ctx, script)
	return err
}

func ensureFnm(ctx context.Context) error {
	if hasCommand("fnm") {
		return nil
	}
	script := `set -e
curl -fsSL https://fnm.vercel.app/install | bash -s -- --install-dir /usr/local/share/fnm --skip-shell
ln -sfn /usr/local/share/fnm/fnm /usr/local/bin/fnm`
	_, err := runShell(ctx, script)
	return err
}

func installNodeVersion(ctx context.Context, version string) error {
	if err := validateNodeVersion(version); err != nil {
		return err
	}
	script := `set -euo pipefail
export PATH="/usr/local/bin:/usr/local/share/fnm:$PATH"
export FNM_DIR="${FNM_DIR:-$HOME/.local/share/fnm}"
eval "$(fnm env --shell bash)"
fnm install "$SM_NODE_VER"`
	_, err := runShellEnv(ctx, append(fnmEnv(), "SM_NODE_VER="+version), script)
	return err
}

func enableService(ctx context.Context, unit string) error {
	if !hasCommand("systemctl") {
		return nil
	}
	_, err := run(ctx, "systemctl", "enable", "--now", unit)
	return err
}
