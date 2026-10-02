package executor

import "fmt"

// Allowlisted runtime versions and their apt/fnm targets.
var phpVersions = map[string]struct{}{
	"8.1": {},
	"8.2": {},
	"8.3": {},
	"8.4": {},
	"8.5": {},
}

var nodeVersions = map[string]struct{}{
	"18": {},
	"20": {},
	"22": {},
	"24": {},
}

var basePackages = []string{
	"nginx",
	"certbot",
	"python3-certbot-nginx",
	"git",
	"unzip",
	"curl",
	"ca-certificates",
	"gnupg",
	"lsb-release",
}

var phpExtensions = []string{
	"cli", "fpm", "mbstring", "xml", "curl", "zip", "mysql", "pgsql", "redis",
}

var addonPackages = map[string][]string{
	"redis":    {"redis-server"},
	"mysql":    {"mysql-server"},
	"mariadb":  {"mariadb-server"},
	"postgres": {"postgresql"},
}

// systemd unit names (may differ from apt package names)
var addonServices = map[string]string{
	"redis":    "redis-server",
	"mysql":    "mysql",
	"mariadb":  "mariadb",
	"postgres": "postgresql",
}

func phpPackages(version string) ([]string, error) {
	if _, ok := phpVersions[version]; !ok {
		return nil, fmt.Errorf("unsupported php version %q", version)
	}
	out := make([]string, 0, len(phpExtensions))
	for _, ext := range phpExtensions {
		out = append(out, "php"+version+"-"+ext)
	}
	return out, nil
}

func validateNodeVersion(version string) error {
	if _, ok := nodeVersions[version]; !ok {
		return fmt.Errorf("unsupported node version %q", version)
	}
	return nil
}

func payloadStringSlice(payload map[string]any, key string) []string {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func payloadBoolMap(payload map[string]any, key string) map[string]bool {
	out := map[string]bool{}
	raw, ok := payload[key]
	if !ok || raw == nil {
		return out
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return out
	}
	for k, v := range m {
		if b, ok := v.(bool); ok {
			out[k] = b
		}
	}
	return out
}
