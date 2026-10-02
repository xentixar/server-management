// Package provision SSHs into a host to test connectivity and install the agent.
package provision

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"api/app/sshrun"
	"api/config"
)

const (
	remoteBin   = "/usr/local/bin/server-agent"
	remoteEnv   = "/etc/server-agent/agent.env"
	remoteUnit  = "/etc/systemd/system/server-agent.service"
	dialTimeout = 20 * time.Second
)

type Target = sshrun.Target

type Result struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func TestConnection(target Target) Result {
	client, err := sshrun.Dial(target, dialTimeout)
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	defer client.Close()

	out, errOut, err := client.Run("uname -srm")
	if err != nil {
		msg := strings.TrimSpace(errOut)
		if msg == "" {
			msg = err.Error()
		}
		return Result{OK: false, Message: msg}
	}
	return Result{OK: true, Message: "SSH ok: " + strings.TrimSpace(out)}
}

// InstallAgent uploads the matching linux binary, writes env + systemd unit, and starts it.
func InstallAgent(target Target, apiURL, serverToken string) Result {
	if strings.TrimSpace(apiURL) == "" {
		return Result{OK: false, Message: "PUBLIC_API_URL is not configured on the API"}
	}
	if strings.TrimSpace(serverToken) == "" {
		return Result{OK: false, Message: "server token is empty"}
	}

	client, err := sshrun.Dial(target, dialTimeout)
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	defer client.Close()

	archOut, archErr, err := client.Run("uname -m")
	if err != nil {
		return Result{OK: false, Message: "detect arch: " + firstNonEmpty(strings.TrimSpace(archErr), err.Error())}
	}
	goarch, err := mapGOARCH(strings.TrimSpace(archOut))
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}

	binPath := filepath.Join(config.AgentBinDir(), "agent-linux-"+goarch)
	binary, err := os.ReadFile(binPath)
	if err != nil {
		return Result{OK: false, Message: fmt.Sprintf(
			"agent binary missing at %s (build with: GOOS=linux GOARCH=%s go build -o %s ./agent)",
			binPath, goarch, binPath,
		)}
	}

	if _, _, err := client.Run("mkdir -p /etc/server-agent /tmp"); err != nil {
		return Result{OK: false, Message: "mkdir: " + err.Error()}
	}

	// Stop first so replacing the binary does not hit ETXTBSY ("Text file busy").
	_, _, _ = client.Run("systemctl stop server-agent.service 2>/dev/null || true")

	tmpBin := remoteBin + ".new"
	if err := client.UploadBytes(tmpBin, "0755", binary); err != nil {
		return Result{OK: false, Message: "upload binary: " + err.Error()}
	}
	if _, errOut, err := client.Run("mv -f " + tmpBin + " " + remoteBin); err != nil {
		return Result{OK: false, Message: "install binary: " + firstNonEmpty(strings.TrimSpace(errOut), err.Error())}
	}

	envBody := fmt.Sprintf("API_URL=%s\nSERVER_TOKEN=%s\nHEARTBEAT_INTERVAL=15s\nPOLL_INTERVAL=5s\n",
		strings.TrimRight(apiURL, "/"),
		serverToken,
	)
	if err := client.UploadBytes(remoteEnv, "0600", []byte(envBody)); err != nil {
		return Result{OK: false, Message: "upload env: " + err.Error()}
	}

	unit := `[Unit]
Description=Server management agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=/etc/server-agent/agent.env
ExecStart=/usr/local/bin/server-agent
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`
	if err := client.UploadBytes(remoteUnit, "0644", []byte(unit)); err != nil {
		return Result{OK: false, Message: "upload unit: " + err.Error()}
	}

	if _, errOut, err := client.Run("systemctl daemon-reload && systemctl enable server-agent.service && systemctl restart server-agent.service"); err != nil {
		return Result{OK: false, Message: "systemd: " + firstNonEmpty(strings.TrimSpace(errOut), err.Error())}
	}

	// Brief wait — systemd may report activating for a moment.
	time.Sleep(500 * time.Millisecond)
	statusOut, _, _ := client.Run("systemctl is-active server-agent.service || true")
	status := strings.TrimSpace(statusOut)
	if status != "active" && status != "activating" {
		return Result{OK: false, Message: "agent service not active (got " + status + ")"}
	}

	return Result{OK: true, Message: "agent installed and started (" + goarch + ")"}
}

func TargetFromServer(host, user, method, key, password, passphrase string, port int) Target {
	return Target{
		Host:       host,
		Port:       port,
		User:       user,
		Method:     method,
		KeyPEM:     key,
		Password:   password,
		Passphrase: passphrase,
	}
}

func mapGOARCH(unameM string) (string, error) {
	switch unameM {
	case "x86_64", "amd64":
		return "amd64", nil
	case "aarch64", "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported architecture %q (need x86_64 or aarch64)", unameM)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
