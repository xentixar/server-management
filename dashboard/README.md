# Outpost dashboard

Nuxt frontend for **Outpost**. This folder is the UI app, frontend docs, and dashboard roadmap.

Platform docs (features, backend architecture, monorepo roadmap): [../docs/](../docs/) and [../README.md](../README.md).

## Scope (summary)

| In scope here | Out of scope (see root [docs/](../docs/)) |
|---------------|---------------------------------------------|
| Pages, layouts, components | Go API, Postgres, Redis |
| Client validation (Zod) | Agent, bash bootstrap |
| API client / Nitro proxy | Terraform modules |
| Auth, orgs, members UI + ability gates | Billing provider webhooks |

Details: [docs/SCOPE.md](docs/SCOPE.md).

## Stack

- [Nuxt 4](https://nuxt.com) + [Vue 3](https://vuejs.org)
- [Nuxt UI 4](https://ui.nuxt.com)
- TypeScript, Zod, TanStack Table

## Auth & permissions (quick)

Session cookie `outpost_token` → `useSession()` / `useOrganizations()`.  
UI checks use a Filament-style gate:

```ts
import { Ability } from '~/utils/auth/abilities'
const { can } = useGate()
// v-if="can(Ability.MembersInvite)"
```

- **Human docs:** [docs/AUTH.md](docs/AUTH.md)
- **AI / agents:** [AGENTS.md](AGENTS.md) and `.cursor/rules/auth.mdc`

Ability keys must match the Go API (`../api/app/constants/`). Do not hardcode role checks in components.

## Documentation

| Document | Description |
|----------|-------------|
| [docs/README.md](docs/README.md) | Frontend doc index |
| [docs/AUTH.md](docs/AUTH.md) | Session, invites, organizations, `useGate` |
| [docs/SCOPE.md](docs/SCOPE.md) | Boundaries and template inventory |
| [docs/DESIGN.md](docs/DESIGN.md) | Navigation, flows, UI conventions |
| [docs/API.md](docs/API.md) | Control-plane access patterns |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Dashboard phases |
| [AGENTS.md](AGENTS.md) | Guidelines for coding agents |

## Development

Requires the Go API (default `http://localhost:8080`, proxied as `/api/v1`). Set `NUXT_API_BASE` if needed.

```bash
pnpm install
pnpm dev      # http://localhost:3000
pnpm build
pnpm preview
pnpm lint
pnpm typecheck
```
