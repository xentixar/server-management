# Frontend scope

What the **dashboard** folder is responsible for in the Outpost monorepo.

## In scope

- Authenticated app shell (sidebar, org switcher, command palette, notifications slideover).
- Auth pages (login, register, password reset) when implemented — separate layout, no ops nav.
- Organization, server, and site **screens**: forms, tables, wizards, log viewers, status displays.
- Client-side validation and UX (Zod, loading/error/empty states, role-based **visibility** of actions).
- HTTP client to control plane API (direct or via Nitro proxy).
- Real-time log/status UI (SSE/WebSocket consumption only — streams originate from API).

## Out of scope

These belong in sibling directories documented at [../../docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md):

| Concern | Owner |
|---------|--------|
| User sessions, RBAC enforcement | `api/` |
| Postgres schema, migrations | `api/` |
| Job queue, workers | `api/` + Redis |
| Terraform apply, cloud credentials | `api/` + `terraform/` |
| SSH bootstrap, Nginx edits, deploy scripts | `agent/` + `scripts/` |
| Let’s Encrypt, firewall on host | `agent/` |

The dashboard must not embed secrets, run Terraform, or open SSH from the browser.

## Current state (Phase 0)

| Asset | Status |
|-------|--------|
| `app/layouts/default.vue` | Template demo nav |
| `app/components/TeamsMenu.vue` | Mock orgs |
| `app/pages/index.vue` | Demo analytics home |
| `app/pages/customers.vue` | Demo table (pattern for servers) |
| `app/pages/inbox.vue` | Demo mail |
| `app/pages/settings/*` | Demo org-settings pattern |
| `server/api/*.ts` | Mock JSON for template |
| Auth routes / `layouts/auth.vue` | Not started |
| `/servers`, `/sites`, `/organization` | Not started |

Product behavior is defined in [../../docs/FEATURES.md](../../docs/FEATURES.md). UI mapping is in [DESIGN.md](DESIGN.md).

## Dependencies on other components

Until the API exists, the app uses Nitro mocks under `server/api/`. Phase 1+ replaces mocks with API calls per [API.md](API.md).

Timeline: [ROADMAP.md](ROADMAP.md). Monorepo-wide phases: [../../docs/ROADMAP.md](../../docs/ROADMAP.md).
