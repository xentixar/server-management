# Outpost dashboard

Nuxt frontend for **Outpost** (codename). This folder contains **only** the UI app, its scope, design notes, and frontend timeline.

Platform docs (features, backend architecture, monorepo roadmap): [../docs/](../docs/) and [../README.md](../README.md).

> **Current timeline:** Phase 0 — documentation and IA; template demo pages remain until Outpost screens are built. No production API integration yet.

## Scope (summary)

| In scope here | Out of scope (see root [docs/](../docs/)) |
|---------------|---------------------------------------------|
| Pages, layouts, components | Go API, Postgres, Redis |
| Client validation (Zod) | Agent, bash bootstrap |
| API client / Nitro proxy | Terraform modules |
| Auth & org **UI** | Billing provider webhooks |

Details: [docs/SCOPE.md](docs/SCOPE.md).

## Stack

- [Nuxt 4](https://nuxt.com) + [Vue 3](https://vuejs.org)
- [Nuxt UI 4](https://ui.nuxt.com) — [dashboard template](https://dashboard-template.nuxt.dev/)
- TypeScript, Zod, TanStack Table

## Documentation

| Document | Description |
|----------|-------------|
| [docs/README.md](docs/README.md) | Frontend doc index |
| [docs/SCOPE.md](docs/SCOPE.md) | Boundaries and template demo inventory |
| [docs/DESIGN.md](docs/DESIGN.md) | Navigation, flows, UI conventions |
| [docs/API.md](docs/API.md) | How the app will talk to the control plane |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Dashboard phases and migration checklist |

## Development

```bash
pnpm install
pnpm dev      # http://localhost:3000
pnpm build
pnpm preview
pnpm lint
pnpm typecheck
```

## Template demo

Demo routes (Home, Inbox, Customers) and mock handlers in `server/api/` are placeholders. See [docs/ROADMAP.md](docs/ROADMAP.md) for replacement plan.
