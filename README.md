# Outpost

Organization-centric platform to connect servers, install an on-host agent, and (soon) deploy sites from one dashboard.

## Repository layout

| Path | Role | Status |
|------|------|--------|
| [`dashboard/`](dashboard/) | Nuxt UI | Auth, orgs, members, servers wired to API |
| [`api/`](api/) | Go control plane | Auth, orgs, RBAC, server CRUD, SSH bootstrap, agent endpoints |
| [`agent/`](agent/) | On-server agent | Heartbeat + job poll loop |
| `terraform/` | DigitalOcean / AWS modules | Planned |
| [`docs/`](docs/) | Platform docs | |

## How it works (today)

```text
Browser → Dashboard (Nuxt) → Go API → Postgres
                              │
                              ├─ SSH (one-time) → install agent on VPS / Proxmox VM
                              │
                              └─ HTTPS ← agent heartbeats + pulls jobs
```

1. User signs in and works inside an **organization**.
2. **Add server** with SSH credentials (key or password).
3. API **tests SSH**, saves the server, uploads a Linux agent binary, writes env (`API_URL` + `SERVER_TOKEN`), and starts `server-agent.service`.
4. Agent **heartbeats** → server status becomes `online`. Jobs are claimed via poll (`ping` today; deploy/nginx later).

Details: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md), [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md), [docs/AGENT.md](docs/AGENT.md).

## Quick start

### 1. API

```bash
cd api
cp .env.example .env
# Set DB_* and PUBLIC_API_URL to a URL your VMs can reach (LAN IP, not localhost)
# Example: PUBLIC_API_URL=http://192.168.1.50:8080

# Build agent binaries the API will SCP onto hosts
mkdir -p bin
(cd ../agent && GOOS=linux GOARCH=amd64 go build -o ../api/bin/agent-linux-amd64 .)
(cd ../agent && GOOS=linux GOARCH=arm64 go build -o ../api/bin/agent-linux-arm64 .)

go run .
```

See [api/README.md](api/README.md).

### 2. Dashboard

```bash
cd dashboard
pnpm install
# optional: NUXT_API_BASE=http://localhost:8080
pnpm dev
```

Open http://localhost:3000 — register, then **Servers → Add server**.

### 3. Home-lab / Proxmox VM

On the guest: OpenSSH, systemd, and a user that can write `/usr/local/bin` + units (usually `root`). From the UI use **Test connection**, then **Add server**.

Full checklist: [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).

## Documentation

| Document | Description |
|----------|-------------|
| [docs/README.md](docs/README.md) | Doc index |
| [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) | Local + Proxmox setup |
| [docs/AGENT.md](docs/AGENT.md) | Agent protocol, install paths, env |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Components, trust, flows |
| [docs/FEATURES.md](docs/FEATURES.md) | Product catalog |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Delivery phases |
| [dashboard/docs/](dashboard/docs/) | UI, auth gate, API client |

## Status

**Phase 1 complete** (identity & orgs). **Phase 2 in progress**: manual servers + agent heartbeat/bootstrap work; sites/deploys still planned.
