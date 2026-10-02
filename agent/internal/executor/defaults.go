package executor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// applyPHPDefault points /usr/bin/php (and friends) at the chosen Sury build via update-alternatives.
func applyPHPDefault(ctx context.Context, version string) error {
	if _, ok := phpVersions[version]; !ok {
		return fmt.Errorf("unsupported php version %q", version)
	}
	prio := strings.ReplaceAll(version, ".", "")
	if _, err := strconv.Atoi(prio); err != nil {
		prio = "1"
	}
	script := fmt.Sprintf(`set -e
BIN="/usr/bin/php%s"
test -x "$BIN"
update-alternatives --install /usr/bin/php php "$BIN" %s >/dev/null 2>&1 || true
update-alternatives --set php "$BIN"
for tool in phpize php-config phar phar.phar; do
  tbin="/usr/bin/${tool}%s"
  if [ -x "$tbin" ]; then
    update-alternatives --install "/usr/bin/$tool" "$tool" "$tbin" %s >/dev/null 2>&1 || true
    update-alternatives --set "$tool" "$tbin" >/dev/null 2>&1 || true
  fi
done
`, version, prio, version, prio)
	_, err := runShell(ctx, script)
	return err
}

func fnmEnv() []string {
	home := "/root"
	return []string{
		"HOME=" + home,
		"FNM_DIR=" + home + "/.local/share/fnm",
		"PATH=/usr/local/bin:/usr/local/share/fnm:/usr/bin:/bin",
	}
}

// applyNodeDefault sets fnm's default and exposes node/npm/npx on PATH via /usr/local/bin.
func applyNodeDefault(ctx context.Context, version string) error {
	if err := validateNodeVersion(version); err != nil {
		return err
	}
	// Resolve bin via aliases/default (no nested bash -lc — that was failing with empty stderr).
	script := `set -euo pipefail
export PATH="/usr/local/bin:/usr/local/share/fnm:$PATH"
export FNM_DIR="${FNM_DIR:-$HOME/.local/share/fnm}"
if ! command -v fnm >/dev/null 2>&1; then
  echo "fnm not on PATH" >&2
  exit 1
fi
eval "$(fnm env --shell bash)"
fnm install "$SM_NODE_VER"
fnm default "$SM_NODE_VER"

BIN=""
if [ -x "$FNM_DIR/aliases/default/bin/node" ]; then
  BIN="$FNM_DIR/aliases/default/bin"
else
  # match v24*, v24.x.x, or exact
  for cand in "$FNM_DIR"/node-versions/v"$SM_NODE_VER"*/installation/bin "$FNM_DIR"/node-versions/"$SM_NODE_VER"*/installation/bin; do
    if [ -x "$cand/node" ]; then
      BIN="$cand"
    fi
  done
fi

if [ -z "$BIN" ] || [ ! -x "$BIN/node" ]; then
  echo "node binary not found for $SM_NODE_VER (FNM_DIR=$FNM_DIR)" >&2
  fnm list >&2 || true
  ls -la "$FNM_DIR" >&2 || true
  ls -la "$FNM_DIR/node-versions" >&2 || true
  exit 1
fi

ln -sfn "$BIN/node" /usr/local/bin/node
ln -sfn "$BIN/npm" /usr/local/bin/npm
if [ -e "$BIN/npx" ]; then ln -sfn "$BIN/npx" /usr/local/bin/npx; fi

printf '%s\n' \
  'export PATH="/usr/local/share/fnm:$PATH"' \
  'export FNM_DIR="${FNM_DIR:-$HOME/.local/share/fnm}"' \
  'command -v fnm >/dev/null 2>&1 && eval "$(fnm env --shell bash)"' \
  > /etc/profile.d/fnm.sh

/usr/local/bin/node -v >&2
`
	_, err := runShellEnv(ctx, append(fnmEnv(), "SM_NODE_VER="+version), script)
	return err
}

func highestVersion(versions []string) string {
	best := ""
	for _, v := range versions {
		if best == "" || compareVersion(v, best) > 0 {
			best = v
		}
	}
	return best
}

// compareVersion returns 1 if a>b, -1 if a<b, 0 if equal (numeric dotted segments).
func compareVersion(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		var ai, bi int
		if i < len(as) {
			ai, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			bi, _ = strconv.Atoi(bs[i])
		}
		if ai > bi {
			return 1
		}
		if ai < bi {
			return -1
		}
	}
	return 0
}
