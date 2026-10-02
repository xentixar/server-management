# Roadmap

Phased delivery for the **Outpost monorepo**. Dates are **TBD**; order reflects dependency.

Legend:

- **Dashboard** — [dashboard/](../dashboard/)
- **API** — Go control plane
- **Agent** — Go on servers
- **TF** — Terraform modules

Dashboard-specific tasks: [dashboard/docs/ROADMAP.md](../dashboard/docs/ROADMAP.md).

---

## Phase 0 — Foundation

**Goal:** Product narrative, platform docs, and UI information architecture.

| Dashboard | API | Agent | TF |
|-----------|-----|-------|-----|
| Frontend docs + template mapping | — | — | — |
| Root `docs/` | — | — | — |

**Status:** Done.

---

## Phase 1 — Identity & organization

**Goal:** Real users, orgs, and members.

| Dashboard | API | Agent | TF |
|-----------|-----|-------|-----|
| Auth UI, org switcher, organization settings | Auth, sessions, Postgres | — | — |
| Members UI | Org CRUD, invites, RBAC | — | — |

**Status:** Done. Exit criteria met (register, default org, invite, switch; API enforces roles).

---

## Phase 2 — Manual servers, agent, basic site & deploy

**Goal:** End-to-end path on a single VPS / home-lab VM without cloud provisioning.

| Dashboard | API | Agent | TF |
|-----------|-----|-------|-----|
| Servers list, add form, test connection | Server CRUD, SSH test, bootstrap | Binary install, heartbeat, job poll | — |
| Site/deploy UI (**remaining**) | Deploy queue (**remaining**) | Nginx / git deploy (**remaining**) | — |
| Log viewer (**remaining**) | Log stream (**remaining**) | | |

**Status:** In progress.

**Done:** Manual server CRUD; SSH test connection; agent SCP + systemd + env; heartbeat flips status `online`; `ping` jobs; dashboard `useServers` wired.

**Remaining:** Sites, deploys, log streaming, richer provision UI (real job stream vs stub).

**Exit criteria:** Manual server with agent; one site deploys successfully; logs visible in UI.

---

## Phase 3 — Cloud provisioning & billing

**Goal:** Create servers in DO/AWS and monetize orgs.

| Dashboard | API | Agent | TF |
|-----------|-----|-------|-----|
| Terraform wizards, billing UI | Provision jobs, Stripe webhooks | Post-provision bootstrap | DO + AWS modules |

**Exit criteria:** User provisions DO or AWS server from UI; plan limits block over-quota creates.

---

## Phase 4 — Production operations

**Goal:** Forge/Ploi-class operations depth.

| Dashboard | API | Agent | TF |
|-----------|-----|-------|-----|
| Runtimes, cron, SSL, DB, network UIs | Matching CRUD + jobs | Nginx, ACME, firewall, services | — |

**Exit criteria:** Laravel/Next/Nuxt-style site with SSL, cron, zero-downtime deploy, and DB without SSH.

---

## Future (post Phase 4)

| Item | Notes |
|------|--------|
| Metrics & alerting | CPU, disk, HTTP error rate |
| OAuth login | GitHub/Google |
| Additional providers | Hetzner, Linode, etc. |
| Remote builders | Build off-server |
| Backup storage to S3 | Org-wide backup policies |
| Audit log UI | Org security tab |
| Public API & CLI | Org-scoped tokens |

---

## Risk & dependencies

| Risk | Mitigation |
|------|------------|
| Agent protocol churn | Version field + compatibility matrix in API |
| Terraform state leaks | Separate state per org/env; no local prod state on laptops |
| SSH credentials in browser | Never persist client-side; upload to API once |
| Long deploy logs | Stream + object storage for archives |
| Multi-server firewall | Server-scoped rules first; cloud SG later |
| Home-lab host keys | Accept-any today; known_hosts before production |

See [ARCHITECTURE.md](ARCHITECTURE.md) and [DEVELOPMENT.md](DEVELOPMENT.md).
