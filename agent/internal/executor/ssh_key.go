package executor

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"agent/internal/protocol"
)

func handleSshKeySync(_ context.Context, job protocol.Job) (string, map[string]any, error) {
	username := payloadString(job.Payload, "user")
	if username == "" {
		username = managedUnixUser
	}
	if username != managedUnixUser {
		return "", nil, fmt.Errorf("unsupported ssh key user %q", username)
	}

	keys, err := sshKeysFromPayload(job.Payload)
	if err != nil {
		return "", nil, err
	}

	u, err := user.Lookup(username)
	if err != nil {
		return "", nil, fmt.Errorf("managed user %q missing; provision the server first", username)
	}
	uidNum, err1 := strconv.Atoi(u.Uid)
	gidNum, err2 := strconv.Atoi(u.Gid)
	if err1 != nil || err2 != nil {
		return "", nil, fmt.Errorf("invalid uid/gid for %s", username)
	}

	home := u.HomeDir
	if home == "" {
		home = "/home/" + username
	}
	sshDir := filepath.Join(home, ".ssh")
	authKeysPath := filepath.Join(sshDir, "authorized_keys")

	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return "", nil, err
	}
	if err := os.Chown(sshDir, uidNum, gidNum); err != nil {
		return "", nil, err
	}
	if err := os.Chmod(sshDir, 0o700); err != nil {
		return "", nil, err
	}

	var b strings.Builder
	b.WriteString("# Managed by server-agent - do not edit\n")
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(authKeysPath, []byte(b.String()), 0o600); err != nil {
		return "", nil, err
	}
	if err := os.Chown(authKeysPath, uidNum, gidNum); err != nil {
		return "", nil, err
	}
	if err := os.Chmod(authKeysPath, 0o600); err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("synced %d ssh keys for %s", len(keys), username), nil, nil
}

func sshKeysFromPayload(payload map[string]any) ([]string, error) {
	if payload == nil {
		return nil, nil
	}
	raw, ok := payload["keys"]
	if !ok || raw == nil {
		return nil, nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("keys must be a list")
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		key := payloadString(m, "public_key")
		if key == "" {
			return nil, fmt.Errorf("public_key is required")
		}
		if strings.ContainsAny(key, "\n\r") {
			return nil, fmt.Errorf("public_key must be a single line")
		}
		parts := strings.Fields(key)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid public_key")
		}
		switch parts[0] {
		case "ssh-ed25519", "ssh-rsa", "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521", "sk-ssh-ed25519@openssh.com", "sk-ecdsa-sha2-nistp256@openssh.com":
		default:
			return nil, fmt.Errorf("unsupported key type %q", parts[0])
		}
		out = append(out, key)
	}
	return out, nil
}
