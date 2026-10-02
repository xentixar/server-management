# AI guidelines — Outpost dashboard

Short rules for agents working in `dashboard/`. Product detail: [docs/AUTH.md](docs/AUTH.md).

## Auth & API

- Use `useApi().api()` for all Go control-plane calls (Bearer cookie via Nitro `/api/v1` proxy).
- Session: `useSession()` — do not invent a second auth store.
- Orgs: `useOrganizations()` + `ensureOrganizations()` / `clearOrganizations()` (middleware-safe).
- Members: `useMembers()` + `await fetchAll()` on the members page for SSR; `clearMembers()` on logout.
- Servers: `useServers()` + `await fetchAll()` on org home / servers pages; `clearServers()` on logout. Create/testConnection hit real SSH bootstrap on the API — do not fake SSH in the browser.
- Invite accept lives at `/invites/:orgId`. Reuse `/login` and `/register` with `redirect` (path only) and optional sibling `email` query — never nest `?email=` inside `redirect`.
- Platform setup (PUBLIC_API_URL, agent binaries, Proxmox): [../docs/DEVELOPMENT.md](../docs/DEVELOPMENT.md).

## Permissions (Filament-style)

- Ability strings live in `app/utils/auth/abilities.ts` and **must match** API keys.
- Gate UI with `useGate()`:

```ts
import { Ability } from '~/utils/auth/abilities'
const { can } = useGate()
// v-if="can(Ability.MembersInvite)"
```

- Do **not** check `role === 'admin'` in components. Do **not** hardcode permission matrices in the frontend.
- Adding an ability: API constants + role map + policy → mirror key in `Ability` → `can(...)` in UI.
- Org payload includes `role` and `permissions[]` from the API; that is the only client source of truth.

## Code style

- Follow existing Nuxt UI patterns; prefer smallest diff; no new deps without need.
- Prefer `useState` over module-level `reactive` for anything rendered on SSR.
- Keep stubs with fixed clocks if they still render on SSR (no `Date.now()` in seed data).
- Update [docs/AUTH.md](docs/AUTH.md) when auth/permission flows change.
