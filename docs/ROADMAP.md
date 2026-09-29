# Roadmap

Phased delivery for the **Outpost monorepo**. Dates are **TBD**; order reflects dependency (API and agent before full UI depth).

Legend:

- **Dashboard** — [dashboard/](../dashboard/)
- **API** — Go control plane (planned)
- **Agent** — Go + bash on servers (planned)
- **TF** — Terraform modules (planned)

Dashboard-specific tasks and migration checklist: [dashboard/docs/ROADMAP.md](../dashboard/docs/ROADMAP.md).

---

## Phase 0 — Foundation (current)

**Goal:** Product narrative, platform docs, and UI information architecture.

| Dashboard | API | Agent | TF |
|-----------|-----|-------|-----|
| Frontend docs + template mapping | — | — | — |
| Optional: stub routes, rename nav | — | — | — |
| Root `docs/` (features, architecture) | — | — | — |

**Exit criteria:** Docs complete; team agrees on IA ([dashboard/docs/DESIGN.md](../dashboard/docs/DESIGN.md)) and feature scope ([FEATURES.md](FEATURES.md)).

---

## Phase 1 — Identity & organization

**Goal:** Real users, orgs, and members.

| Dashboard | API | Agent | TF |
|-----------|-----|-------|-----|
| Auth UI, org switcher, organization settings | Auth, sessions, Postgres | — | — |
| Members UI | Org CRUD, invites, RBAC | — | — |
| Types + Zod for forms | OpenAPI or shared types | — | — |

**Exit criteria:** User can register, get default org, invite member, switch org; API enforces roles.

---

## Phase 2 — Manual servers, agent, basic site & deploy

**Goal:** End-to-end path on a single VPS without cloud provisioning.

| Dashboard | API | Agent | TF |
|-----------|-----|-------|-----|
| Servers list, manual add, deploy UI | Server CRUD, deploy queue | Install script, heartbeat, Nginx | — |
| Log viewer | Log stream/storage | Git pull/build | — |

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
| SSH credentials in browser | Never; upload encrypted to API once |
| Long deploy logs | Stream + object storage for archives |
| Multi-server firewall | Server-scoped rules first; cloud SG later |

See [ARCHITECTURE.md](ARCHITECTURE.md) for trust boundaries.
