# Development

How to run the monorepo locally and test against a real machine (e.g. Proxmox VM).

## Prerequisites

- Go 1.22+ (repo uses Go 1.27.x toolchain)
- Node 20+ / pnpm (dashboard)
- Postgres
- A Linux host reachable over SSH from the machine running the API (Proxmox guest, VPS, etc.)

## API

```bash
cd api
cp .env.example .env
```

| Variable | Required | Notes |
|----------|----------|--------|
| `APP_PORT` | yes | e.g. `8080` |
| `APP_ENV` | yes | e.g. `local` |
| `DB_*` | yes | Postgres connection |
| `PUBLIC_API_URL` | yes for bootstrap | Base URL **remote agents** use to call the API. Must be reachable from the VM (LAN IP / hostname). Not `localhost` unless agent runs on the same host. Also used as the OAuth callback host for git providers. |
| `PUBLIC_DASHBOARD_URL` | yes for git OAuth | Where the API redirects after a provider OAuth callback (e.g. `http://192.168.1.50:3000`). |
| `APP_ENCRYPTION_KEY` | yes for git OAuth | 32-byte key, base64-encoded. Generate with `openssl rand -base64 32`. Encrypts org git connection tokens at rest. |
| `GITHUB_OAUTH_CLIENT_ID` / `GITHUB_OAUTH_CLIENT_SECRET` | for GitHub connect | OAuth App callback: `{PUBLIC_API_URL}/api/v1/git/oauth/callback/github` |
| `GITLAB_OAUTH_CLIENT_ID` / `GITLAB_OAUTH_CLIENT_SECRET` | for GitLab.com connect | Callback: `{PUBLIC_API_URL}/api/v1/git/oauth/callback/gitlab` |
| `BITBUCKET_OAUTH_CLIENT_ID` / `BITBUCKET_OAUTH_CLIENT_SECRET` | for Bitbucket Cloud connect | Callback: `{PUBLIC_API_URL}/api/v1/git/oauth/callback/bitbucket` |
| `AGENT_BIN_DIR` | no | Default `bin`. Contains `agent-linux-amd64` and/or `agent-linux-arm64`. |

Build agent binaries (from repo root or `api/`):

```bash
mkdir -p api/bin
(cd agent && GOOS=linux GOARCH=amd64 go build -o ../api/bin/agent-linux-amd64 .)
(cd agent && GOOS=linux GOARCH=arm64 go build -o ../api/bin/agent-linux-arm64 .)
```

Run:

```bash
cd api && go run .
# listens on 0.0.0.0:APP_PORT
```

Health: `GET /api/v1/health`.

## Dashboard

```bash
cd dashboard
pnpm install
pnpm dev   # http://localhost:3000
```

Browser calls same-origin `/api/v1/*`; Nitro proxies to `NUXT_API_BASE` / `runtimeConfig.apiBase` (default `http://localhost:8080`).

## Agent (manual run)

Normally the API installs the agent over SSH. To run locally against the API:

```bash
cd agent
cp .env.example .env
# API_URL=http://localhost:8080
# SERVER_TOKEN=<agent_token from server create response / DB>
go run .
```

See [AGENT.md](AGENT.md).

## End-to-end: add a Proxmox (or any) VM

1. **Guest prep**
   - SSH listening (`sshd`)
   - `systemd` available
   - User with permission to write `/usr/local/bin`, `/etc/server-agent`, `/etc/systemd/system` (typically `root`)
   - Network path from API host → guest:22, and from guest → `PUBLIC_API_URL`

2. **API `.env`**
   - Set `PUBLIC_API_URL=http://<your-LAN-IP>:8080` (IP of the machine running `go run .`)
   - Rebuild `api/bin/agent-linux-*` if you changed agent code

3. **UI**
   - Register / login → open org → **Servers → Add server**
   - Fill hostname, IP, SSH user/port, key or password
   - **Test connection** — API SSHs in and runs `uname` (no install)
   - **Add server** — SSH again, uploads binary, writes env, `systemctl enable --now server-agent`

4. **Verify**
   - On guest: `systemctl status server-agent` and `journalctl -u server-agent -f`
   - On guest: `cat /etc/server-agent/agent.env` (token + API URL)
   - In UI: server **agent** badge should move to `online` after heartbeat
   - Retry without recreating: `POST /api/v1/organizations/:orgId/servers/:id/bootstrap`

### Common failures

| Symptom | Likely cause |
|---------|----------------|
| Test connection fails | Wrong IP/user/key; firewall; SSH not on port 22 |
| Binary missing | Build `agent-linux-amd64` / `arm64` into `api/bin` |
| Service inactive | Guest lacks systemd, or user cannot enable units |
| Agent never online | `PUBLIC_API_URL` not reachable from guest; wrong token; API not listening on LAN |
| Create times out | SSH/upload slow; API `WriteTimeout` is 120s — check network |

## API surface (servers / agent)

| Method | Path | Who |
|--------|------|-----|
| `POST` | `/organizations/:org/servers/test-connection` | User JWT |
| `POST` | `/organizations/:org/servers` | User JWT (SSH test + create + bootstrap) |
| `POST` | `/organizations/:org/servers/:id/bootstrap` | User JWT (reinstall agent) |
| CRUD | `/organizations/:org/servers`… | User JWT + server abilities |
| `POST` | `/agent/heartbeat` | Agent token |
| `GET` | `/agent/jobs/next` | Agent token |
| `POST` | `/agent/jobs/:id/result` | Agent token |

Abilities: `organization.servers.view|create|update|delete`, plus `organization.servers.{runtimes,services,crons}.{view,manage}` (see `api/app/constants/`).

## Outpost HTTP workspace

The **Server Management** Outpost workspace has Auth, Organizations, Organization Members, Server, and Agent collections for manual API calls (`BASE_URL`, `BEARER_TOKEN`, `ORGANIZATION_ID`, `SERVER_ID`, `AGENT_TOKEN`).
