# Outpost

**Codename** for an organization-centric platform to provision servers, deploy sites, and operate production infrastructure from a single dashboard.

## Repository layout

| Path | Role |
|------|------|
| [`dashboard/`](dashboard/) | Nuxt frontend (UI only) — [README](dashboard/README.md) |
| `api/` | Go control plane — **planned** |
| `agent/` | On-server agent (Go + bash) — **planned** |
| `terraform/` | DigitalOcean / AWS modules — **planned** |
| `scripts/` | Shared bootstrap scripts — **planned** |

## Product summary

Teams sign in, work inside an **organization**, connect or create **servers** (Terraform on DigitalOcean/AWS or manual SSH), then run **sites** (PHP, Laravel, Next.js, Nuxt, and more) with deployments, runtimes, logs, databases, domains, SSL, and network controls.

Full capability spec: [docs/FEATURES.md](docs/FEATURES.md).

## Documentation

| Document | Description |
|----------|-------------|
| [docs/README.md](docs/README.md) | Index and reading order |
| [docs/FEATURES.md](docs/FEATURES.md) | Product feature catalog (all components) |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | API, agent, Terraform, data stores, security |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Monorepo delivery phases |

Frontend-only docs live under [dashboard/docs/](dashboard/docs/) (design, UI scope, dashboard timeline).

## Status

**Phase 0:** Documentation and information architecture. Backend, agent, and Terraform directories are not present yet. The dashboard ships the [Nuxt UI template](https://dashboard-template.nuxt.dev/) with demo pages until Outpost screens land.

## Quick start (dashboard)

```bash
cd dashboard
pnpm install
pnpm dev
```

See [dashboard/README.md](dashboard/README.md) for build, lint, and frontend details.
