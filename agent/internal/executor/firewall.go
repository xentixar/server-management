package executor

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"agent/internal/protocol"
)

const (
	firewallChain     = "SA-FW"
	firewallRulesPath = "/etc/server-agent/firewall.v4"
	firewallUnitPath  = "/etc/systemd/system/server-agent-firewall.service"
	firewallUnitName  = "server-agent-firewall.service"
)

var firewallUIDRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type firewallRule struct {
	uid       string
	name      string
	action    string
	direction string
	protocol  string
	portStart int
	portEnd   int
	source    string
	position  int
}

func handleFirewallSync(ctx context.Context, job protocol.Job) (string, map[string]any, error) {
	sshPort := payloadInt(job.Payload, "ssh_port", 22)
	if sshPort < 1 || sshPort > 65535 {
		return "", nil, fmt.Errorf("invalid ssh_port")
	}
	rules, err := firewallRulesFromPayload(job.Payload)
	if err != nil {
		return "", nil, err
	}
	for _, r := range rules {
		if err := validateFirewallRule(r); err != nil {
			return "", nil, err
		}
	}
	iptables, restore, err := ensureIptables(ctx)
	if err != nil {
		return "", nil, err
	}

	script := buildIptablesRestore(sshPort, rules)
	if err := os.MkdirAll(filepath.Dir(firewallRulesPath), 0o755); err != nil {
		return "", nil, err
	}
	if err := os.WriteFile(firewallRulesPath, []byte(script), 0o600); err != nil {
		return "", nil, err
	}
	if err := ensureFirewallUnit(ctx, iptables, restore); err != nil {
		return "", nil, err
	}
	// Live apply owns the chain; the restore file is for boot via systemd oneshot.
	if err := applyFirewallLive(ctx, iptables, sshPort, rules); err != nil {
		return "", nil, err
	}
	// Drop keep-alive / stale flows so allow→deny takes effect without waiting for idle timeout.
	flushDeniedConntrack(ctx, rules)

	return fmt.Sprintf("synced %d firewall rules (ssh %d/tcp always allowed)", len(rules), sshPort), nil, nil
}

func ensureIptables(ctx context.Context) (iptables, restore string, err error) {
	iptables = findBinary("iptables")
	restore = findBinary("iptables-restore")
	if iptables != "" && restore != "" {
		return iptables, restore, nil
	}
	// Minimal images often ship nftables only; install the iptables CLI (nft backend on modern Ubuntu).
	if _, err := runEnv(ctx, []string{"DEBIAN_FRONTEND=noninteractive"}, "apt-get", "update", "-qq"); err != nil {
		return "", "", fmt.Errorf("iptables not found and apt-get update failed: %w", err)
	}
	if _, err := runEnv(ctx, []string{"DEBIAN_FRONTEND=noninteractive"}, "apt-get", "install", "-y", "-qq", "iptables"); err != nil {
		return "", "", fmt.Errorf("iptables not found and install failed: %w", err)
	}
	iptables = findBinary("iptables")
	restore = findBinary("iptables-restore")
	if iptables == "" || restore == "" {
		return "", "", fmt.Errorf("iptables not found after install")
	}
	return iptables, restore, nil
}

func ensureFirewallUnit(ctx context.Context, iptables, restore string) error {
	unit := `[Unit]
Description=Restore server-agent firewall rules
DefaultDependencies=no
After=network-pre.target
Before=network-online.target

[Service]
Type=oneshot
ExecStart=/bin/bash -c '` + restore + ` --noflush ` + firewallRulesPath + `; while ` + iptables + ` -D INPUT -j ` + firewallChain + ` 2>/dev/null; do :; done; ` + iptables + ` -I INPUT 1 -j ` + firewallChain + `'

[Install]
WantedBy=multi-user.target
`
	if err := os.WriteFile(firewallUnitPath, []byte(unit), 0o644); err != nil {
		return err
	}
	if hasCommand("systemctl") {
		_, _ = run(ctx, "systemctl", "daemon-reload")
		_, _ = run(ctx, "systemctl", "enable", firewallUnitName)
	}
	return nil
}

func ensureFirewallJump(ctx context.Context, iptables string) error {
	for {
		if _, err := run(ctx, iptables, "-D", "INPUT", "-j", firewallChain); err != nil {
			break
		}
	}
	_, err := run(ctx, iptables, "-I", "INPUT", "1", "-j", firewallChain)
	return err
}

func applyFirewallLive(ctx context.Context, iptables string, sshPort int, rules []firewallRule) error {
	_, _ = run(ctx, iptables, "-N", firewallChain)
	if _, err := run(ctx, iptables, "-F", firewallChain); err != nil {
		if _, err2 := run(ctx, iptables, "-N", firewallChain); err2 != nil {
			return err
		}
		if _, err2 := run(ctx, iptables, "-F", firewallChain); err2 != nil {
			return err2
		}
	}
	if err := ensureFirewallJump(ctx, iptables); err != nil {
		return err
	}
	for _, args := range firewallIptablesArgs(sshPort, rules) {
		if _, err := run(ctx, iptables, args...); err != nil {
			return err
		}
	}
	return nil
}

func firewallIptablesArgs(sshPort int, rules []firewallRule) [][]string {
	// Denies before ESTABLISHED so allow→deny closes keep-alive sessions immediately.
	// lo ACCEPT last so on-box curls still hit user denies.
	out := [][]string{
		{"-A", firewallChain, "-p", "tcp", "--dport", strconv.Itoa(sshPort), "-j", "ACCEPT"},
	}
	for _, r := range rules {
		if r.action != "deny" {
			continue
		}
		args, err := firewallRuleArgs(r)
		if err != nil {
			continue
		}
		out = append(out, append([]string{"-A", firewallChain}, args...))
	}
	out = append(out, []string{"-A", firewallChain, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"})
	for _, r := range rules {
		if r.action != "allow" {
			continue
		}
		args, err := firewallRuleArgs(r)
		if err != nil {
			continue
		}
		out = append(out, append([]string{"-A", firewallChain}, args...))
	}
	out = append(out, []string{"-A", firewallChain, "-i", "lo", "-j", "ACCEPT"})
	return out
}

func flushDeniedConntrack(ctx context.Context, rules []firewallRule) {
	ct := findBinary("conntrack")
	if ct == "" {
		return
	}
	for _, r := range rules {
		if r.action != "deny" {
			continue
		}
		switch r.protocol {
		case "tcp", "udp":
			if r.portStart <= 0 {
				_, _ = run(ctx, ct, "-D", "-p", r.protocol)
				continue
			}
			end := r.portEnd
			if end < r.portStart {
				end = r.portStart
			}
			for p := r.portStart; p <= end; p++ {
				_, _ = run(ctx, ct, "-D", "-p", r.protocol, "--dport", strconv.Itoa(p))
			}
		case "icmp", "any":
			_, _ = run(ctx, ct, "-F")
		}
	}
}

func buildIptablesRestore(sshPort int, rules []firewallRule) string {
	var b strings.Builder
	b.WriteString("# Managed by server-agent - do not edit\n")
	b.WriteString("*filter\n")
	b.WriteString(":" + firewallChain + " - [0:0]\n")
	b.WriteString("-F " + firewallChain + "\n")
	for _, args := range firewallIptablesArgs(sshPort, rules) {
		b.WriteString(strings.Join(args, " "))
		b.WriteByte('\n')
	}
	b.WriteString("COMMIT\n")
	return b.String()
}

func firewallRuleArgs(r firewallRule) ([]string, error) {
	var args []string
	if r.source != "" && r.source != "0.0.0.0/0" {
		args = append(args, "-s", r.source)
	}
	switch r.protocol {
	case "any":
		// source only (or bare match)
	case "icmp":
		args = append(args, "-p", "icmp")
	case "tcp", "udp":
		args = append(args, "-p", r.protocol)
		if r.portStart > 0 {
			if r.portEnd > r.portStart {
				args = append(args, "--dport", fmt.Sprintf("%d:%d", r.portStart, r.portEnd))
			} else {
				args = append(args, "--dport", strconv.Itoa(r.portStart))
			}
		}
	default:
		return nil, fmt.Errorf("invalid protocol")
	}
	if r.action == "deny" {
		// REJECT fails faster than DROP and is easier to spot in curl.
		if r.protocol == "tcp" {
			args = append(args, "-j", "REJECT", "--reject-with", "tcp-reset")
		} else {
			args = append(args, "-j", "REJECT", "--reject-with", "icmp-port-unreachable")
		}
	} else {
		args = append(args, "-j", "ACCEPT")
	}
	return args, nil
}

func validateFirewallRule(r firewallRule) error {
	if !firewallUIDRe.MatchString(r.uid) {
		return fmt.Errorf("invalid firewall rule uid")
	}
	if r.action != "allow" && r.action != "deny" {
		return fmt.Errorf("invalid action for %s", r.uid)
	}
	if r.direction != "" && r.direction != "inbound" {
		return fmt.Errorf("only inbound rules are supported")
	}
	switch r.protocol {
	case "tcp", "udp", "icmp", "any":
	default:
		return fmt.Errorf("invalid protocol for %s", r.uid)
	}
	if _, _, err := net.ParseCIDR(r.source); err != nil {
		return fmt.Errorf("invalid source for %s", r.uid)
	}
	if r.protocol == "icmp" || r.protocol == "any" {
		if r.portStart != 0 || r.portEnd != 0 {
			return fmt.Errorf("ports not allowed for %s", r.uid)
		}
		return nil
	}
	if r.portStart < 0 || r.portStart > 65535 || r.portEnd < 0 || r.portEnd > 65535 {
		return fmt.Errorf("port out of range for %s", r.uid)
	}
	if r.portStart == 0 && r.portEnd == 0 {
		return nil
	}
	if r.portStart == 0 || r.portEnd < r.portStart {
		return fmt.Errorf("invalid port range for %s", r.uid)
	}
	return nil
}

func firewallRulesFromPayload(payload map[string]any) ([]firewallRule, error) {
	if payload == nil {
		return nil, nil
	}
	raw, ok := payload["rules"]
	if !ok || raw == nil {
		return nil, nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("rules must be a list")
	}
	out := make([]firewallRule, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if !payloadBool(m, "enabled", true) {
			continue
		}
		source := payloadString(m, "source")
		if source == "" || source == "any" {
			source = "0.0.0.0/0"
		}
		portStart := payloadInt(m, "port_start", 0)
		portEnd := payloadInt(m, "port_end", portStart)
		out = append(out, firewallRule{
			uid:       payloadString(m, "uid"),
			name:      payloadString(m, "name"),
			action:    strings.ToLower(payloadString(m, "action")),
			direction: strings.ToLower(payloadString(m, "direction")),
			protocol:  strings.ToLower(payloadString(m, "protocol")),
			portStart: portStart,
			portEnd:   portEnd,
			source:    source,
			position:  payloadInt(m, "position", 0),
		})
	}
	return out, nil
}
