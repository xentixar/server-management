package helpers

import (
	"errors"
	"strings"

	"golang.org/x/crypto/ssh"
)

func NormalizePublicKey(raw string) (normalized string, fingerprint string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", errors.New("public key is required")
	}
	if strings.ContainsAny(raw, "\n\r") {
		lines := strings.Split(raw, "\n")
		var kept []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			kept = append(kept, line)
		}
		if len(kept) != 1 {
			return "", "", errors.New("public key must be a single line")
		}
		raw = kept[0]
	}

	pub, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(raw))
	if err != nil {
		return "", "", errors.New("invalid public key")
	}
	normalized = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub)))
	if normalized == "" {
		return "", "", errors.New("invalid public key")
	}
	comment = strings.TrimSpace(comment)
	if comment != "" {
		normalized = normalized + " " + comment
	}
	return normalized, ssh.FingerprintSHA256(pub), nil
}
