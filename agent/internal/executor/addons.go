package executor

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
)

const passwordAlphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

const (
	managedDBUser = "sm"
	managedDropIn = "sm.conf"
)

func randomPassword(n int) (string, error) {
	if n < 8 {
		n = 8
	}
	out := make([]byte, n)
	max := big.NewInt(int64(len(passwordAlphabet)))
	for i := range out {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = passwordAlphabet[v.Int64()]
	}
	return string(out), nil
}

// configureAddon sets a generated password for the service. Returns secret fields for the control plane.
func configureAddon(ctx context.Context, name string) (map[string]string, error) {
	pass, err := randomPassword(24)
	if err != nil {
		return nil, err
	}
	switch name {
	case "mysql", "mariadb":
		if err := configureMySQLFamily(ctx, pass); err != nil {
			return nil, err
		}
		return map[string]string{
			"user":     managedDBUser,
			"password": pass,
			"host":     "127.0.0.1",
			"port":     "3306",
		}, nil
	case "postgres":
		if err := configurePostgres(ctx, pass); err != nil {
			return nil, err
		}
		return map[string]string{
			"user":     managedDBUser,
			"password": pass,
			"host":     "127.0.0.1",
			"port":     "5432",
		}, nil
	case "redis":
		if err := configureRedis(ctx, pass); err != nil {
			return nil, err
		}
		return map[string]string{
			"password": pass,
			"host":     "127.0.0.1",
			"port":     "6379",
		}, nil
	default:
		return nil, fmt.Errorf("unknown addon %q", name)
	}
}

func configureMySQLFamily(ctx context.Context, pass string) error {
	// Create managed admin; strip remote root; force local root to socket-only (no password login).
	script := `set -e
mysql --protocol=socket <<SQL
CREATE USER IF NOT EXISTS '${SM_USER}'@'localhost' IDENTIFIED BY '${SM_PASS}';
CREATE USER IF NOT EXISTS '${SM_USER}'@'127.0.0.1' IDENTIFIED BY '${SM_PASS}';
ALTER USER '${SM_USER}'@'localhost' IDENTIFIED BY '${SM_PASS}';
ALTER USER '${SM_USER}'@'127.0.0.1' IDENTIFIED BY '${SM_PASS}';
GRANT ALL PRIVILEGES ON *.* TO '${SM_USER}'@'localhost' WITH GRANT OPTION;
GRANT ALL PRIVILEGES ON *.* TO '${SM_USER}'@'127.0.0.1' WITH GRANT OPTION;
FLUSH PRIVILEGES;
SQL
mysql --protocol=socket -e "DROP USER IF EXISTS 'root'@'%';" || true
mysql --protocol=socket -e "DROP USER IF EXISTS 'root'@'127.0.0.1';" || true
mysql --protocol=socket -e "DROP USER IF EXISTS 'root'@'::1';" || true
# MariaDB unix_socket, else MySQL auth_socket
mysql --protocol=socket -e "ALTER USER 'root'@'localhost' IDENTIFIED VIA unix_socket;" \
  || mysql --protocol=socket -e "ALTER USER 'root'@'localhost' IDENTIFIED WITH auth_socket;" \
  || mysql --protocol=socket -e "ALTER USER 'root'@'localhost' ACCOUNT LOCK;"
mysql --protocol=socket -e "FLUSH PRIVILEGES;" || true`
	_, err := runShellEnv(ctx, []string{
		"SM_USER=" + managedDBUser,
		"SM_PASS=" + pass,
	}, script)
	return err
}

func configurePostgres(ctx context.Context, pass string) error {
	// Managed SUPERUSER for apps; scramble postgres password (never returned).
	// Keep postgres LOGIN so local peer (sudo -u postgres) still works for maintenance.
	rootPass, err := randomPassword(24)
	if err != nil {
		return err
	}
	script := `set -e
sudo -u postgres psql -v ON_ERROR_STOP=1 <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '${SM_USER}') THEN
    CREATE ROLE ${SM_USER} LOGIN SUPERUSER PASSWORD '${SM_PASS}';
  ELSE
    ALTER ROLE ${SM_USER} WITH LOGIN SUPERUSER PASSWORD '${SM_PASS}';
  END IF;
END
\$\$;
ALTER ROLE postgres PASSWORD '${SM_ROOT_PASS}';
SQL`
	_, err = runShellEnv(ctx, []string{
		"SM_USER=" + managedDBUser,
		"SM_PASS=" + pass,
		"SM_ROOT_PASS=" + rootPass,
	}, script)
	return err
}

func configureRedis(ctx context.Context, pass string) error {
	// Prefer drop-in so package upgrades don't wipe requirepass.
	script := `set -e
DROP_IN="` + managedDropIn + `"
mkdir -p /etc/redis/redis.conf.d 2>/dev/null || true
CONF=""
if [ -d /etc/redis/redis.conf.d ]; then
  CONF="/etc/redis/redis.conf.d/${DROP_IN}"
elif [ -f /etc/redis/redis.conf ]; then
  CONF=/etc/redis/redis.conf
elif [ -f /etc/redis.conf ]; then
  CONF=/etc/redis.conf
else
  echo "redis config not found" >&2
  exit 1
fi
if [ "$CONF" = "/etc/redis/redis.conf" ] || [ "$CONF" = "/etc/redis.conf" ]; then
  sed -i -E '/^requirepass /d' "$CONF"
  echo "requirepass ${SM_PASS}" >> "$CONF"
else
  printf 'requirepass %s\n' "${SM_PASS}" > "$CONF"
  if [ -f /etc/redis/redis.conf ] && ! grep -qF "$DROP_IN" /etc/redis/redis.conf; then
    echo "include /etc/redis/redis.conf.d/${DROP_IN}" >> /etc/redis/redis.conf
  fi
fi
systemctl restart redis-server 2>/dev/null || systemctl restart redis`
	_, err := runShellEnv(ctx, []string{"SM_PASS=" + pass}, script)
	return err
}
