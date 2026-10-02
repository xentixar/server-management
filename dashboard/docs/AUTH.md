# Auth & permissions (dashboard)

How the Nuxt app handles sessions, organizations, and Filament-style ability checks.

API source of truth: `../api/app/constants/` and `../api/app/policies/`.

## Session

| Piece | Path | Role |
|-------|------|------|
| Cookie | `outpost_token` | Bearer token (via `useApi`) |
| Session | `useSession()` | `login`, `register`, `logout`, `ensureSession`, `user` |
| Middleware | `app/middleware/auth.global.ts` | Guest routes, invite routes, else require auth + `ensureOrganizations()` |
| Proxy | Nitro `/api/v1/*` | Same-origin → Go API (`NUXT_API_BASE` / runtimeConfig `apiBase`) |

### Rules

- Call the Go API only through `useApi().api()` (never raw `$fetch` to the backend from pages).
- Await `ensureSession()` / `ensureOrganizations()` in middleware — do not re-fetch auth in every page.
- Guest routes: `/login`, `/register`, `/forgot-password`, `/reset-password`, `/verify-email`.
- Invite accept: `/invites/:orgId` (optional `?email=`). Guests may open it; after login/register, redirect back and accept.
- Preserve `?redirect=` between login ↔ register. Keep `redirect` as a **path only** (no nested query). Pass `email` as a sibling query key.

### Invite flow

1. Owner invites → copy `/invites/{orgId}?email=…`
2. Guest → Create account or Sign in (existing auth pages) with `redirect=/invites/{orgId}`
3. Authenticated → page accepts invite → land on that org

## Organizations

| Piece | Path |
|-------|------|
| State | `useOrganizations()` — `useState` list + selected by route slug |
| Ensure | `ensureOrganizations()` / `clearOrganizations()` (safe outside components) |
| Shape | `{ id, name, slug, role?, permissions?: string[] }` |

`permissions` comes from the API for the **current user’s membership**. UI gates use that array only — never invent client-side role matrices.

## Abilities (Filament-style gate)

### Files

| File | Purpose |
|------|---------|
| `app/utils/auth/abilities.ts` | `Ability` string constants (must match API keys) |
| `app/composables/useGate.ts` | `can` / `cannot` against `selected.permissions` |

### Usage in UI

```vue
<script setup lang="ts">
import { Ability } from '~/utils/auth/abilities'

const { can } = useGate()
</script>

<template>
  <UButton
    v-if="can(Ability.MembersInvite)"
    label="Invite people"
    @click="inviteOpen = true"
  />
</template>
```

### Rules

- Prefer `can(Ability.X)` over checking `role === 'admin'`.
- Hide actions the user cannot perform; the API still enforces policies (never trust the UI alone).
- Do not rebuild permission lists in the dashboard. If a key is missing, fix the API role map and mirror the string in `Ability`.

### Adding a new ability

1. API: add constant to `api/app/constants/permissions.go` (`AllAbilities`) and assign on roles in `roles.go`.
2. API: add/extend a method in `api/app/policies/*` and call it from the controller.
3. Dashboard: add the same string to `Ability` in `abilities.ts`.
4. Dashboard: gate UI with `can(Ability.YourKey)`.

## Members

| Piece | Path |
|-------|------|
| Composable | `useMembers()` — list/invite/role/remove/revoke/accept |
| Page load | `await fetchAll()` in the members page (SSR) |
| Clear | `clearMembers()` on logout |

## Servers

| Piece | Path |
|-------|------|
| Composable | `useServers()` — fetch/create/delete/`testConnection`/`bootstrap` |
| Abilities | `ServersView`, `ServersCreate`, `ServersUpdate`, `ServersDelete` |
| Clear | `clearServers()` on logout |

Do not keep module-level stub stores for auth-sensitive data; use `useState` so SSR and client match.

## Related pages

| Route | Notes |
|-------|--------|
| `/login`, `/register` | Auth layout; honor `redirect` + `email` |
| `/invites/:orgId` | Accept flow; links to login/register |
| `/o/:orgSlug/organization` | Gate update/delete |
| `/o/:orgSlug/organization/members` | Gate invite/role/remove |
| `/o/:orgSlug/organization/billing` | Gate manage vs view |
| `/o/:orgSlug/servers` | List / add / detail; gate with server abilities |
