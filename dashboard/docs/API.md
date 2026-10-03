# API integration (dashboard)

How the Nuxt app talks to the **Go control plane**. Server architecture: [../../docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md). Auth details: [AUTH.md](AUTH.md). Local setup: [../../docs/DEVELOPMENT.md](../../docs/DEVELOPMENT.md).

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
- Expects `{ success, message, data }` envelopes (`~/utils/api`) for user routes.
- Do not duplicate business rules in Nitro — proxy only.
- Long operations (SSH test / create+bootstrap) can take up to ~2 minutes; avoid short client timeouts.

## Composables

| Composable | Purpose |
|------------|---------|
| `useSession` | User, login/register/logout, `ensureSession` |
| `useOrganizations` | Memberships, selected org by slug, CRUD |
| `useMembers` | Members + invites for selected org |
| `useGate` | `can` / `cannot` over `selected.permissions` |
| `useServers` | List/create/delete/testConnection/bootstrap against API |
| `useServerMetrics` | Host metrics series for Monitoring tab |
| `useServerFirewall` | Server firewall rule CRUD + sync |
| `useSites` / `useBilling` | Local stubs until those APIs land |

### Servers

- `await fetchAll()` on org home and servers pages.
- Create body uses snake_case fields (`ssh_port`, `auth_method`, `private_key`, …).
- Create response may include `bootstrap: { ok, message }`.
- Abilities: `organization.servers.view|create|update|delete`, plus `organization.servers.{runtimes,services,crons,firewall}.{view,manage}`.
- Metrics: `GET …/servers/:id/metrics?range=1h|6h|24h|7d` → `{ latest, series }` (gated by `servers.view`). Types in `app/utils/metrics.ts`.
- Firewall: `GET/POST/PATCH/DELETE …/firewall/rules`, `POST …/firewall/sync`. Types in `app/utils/firewall.ts`.

## Types & abilities

- Org from API: `{ id, name, slug, role?, permissions? }`.
- Ability strings: `app/utils/auth/abilities.ts` (must match `api/app/constants/permissions.go`).
- Server mapping: `app/utils/servers.ts` (`mapServer`).

## Real-time (planned)

- Deploy/provision logs: SSE or WebSocket from API, keyed by job ID.
- Notifications: poll or push into `NotificationsSlideover`.
- Server provision panel still uses a local stub timeline until job streams land.
