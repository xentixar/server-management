# API integration (dashboard)

How the Nuxt app talks to the **Go control plane**. Server architecture: [../../docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md). Auth details: [AUTH.md](AUTH.md).

## Configuration

```ts
// nuxt.config.ts
runtimeConfig: {
  apiBase: 'http://localhost:8080' // override with NUXT_API_BASE
}
```

Browser calls same-origin `/api/v1/*`; Nitro proxies to `apiBase` (avoids CORS, attaches auth in `useApi`).

## Access pattern

Use **`useApi().api<T>(path, options)`** for all authenticated control-plane requests.

- Sends `Authorization: Bearer` from the `outpost_token` cookie.
- Expects `{ success, message, data }` envelopes (`~/utils/api`).
- Do not duplicate business rules in Nitro — proxy only.

## Composables

| Composable | Purpose |
|------------|---------|
| `useSession` | User, login/register/logout, `ensureSession` |
| `useOrganizations` | Memberships, selected org by slug, CRUD |
| `useMembers` | Members + invites for selected org |
| `useGate` | `can` / `cannot` over `selected.permissions` |
| `useServers` / `useSites` / `useBilling` | Local stubs until those APIs land |

## Types & abilities

- Org from API: `{ id, name, slug, role?, permissions? }`.
- Ability strings: `app/utils/auth/abilities.ts` (must match `api/app/constants/permissions.go`).
- Zod schemas for forms mirror API request bodies.

## Real-time (planned)

- Deploy/provision logs: SSE or WebSocket from API, keyed by job ID.
- Notifications: poll or push into `NotificationsSlideover`.
