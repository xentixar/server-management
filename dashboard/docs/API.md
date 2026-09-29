# API integration (dashboard)

How the Nuxt app will connect to the **Go control plane** once `api/` exists. Server-side behavior is documented in [../../docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md).

## Configuration (planned)

```ts
// nuxt.config.ts
runtimeConfig: {
  public: {
    apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080'
  }
}
```

## Access pattern

Replace mock Nitro routes in `server/api/*.ts` with either:

- **Proxy** — Nitro `$fetch` to API with cookie forwarding (preferred for httpOnly sessions), or
- **Direct** — browser `$fetch` to API with CORS + credentials.

Do not duplicate business logic in Nitro; use it only as a BFF/proxy if needed.

## Composables (planned)

| Composable | Purpose |
|------------|---------|
| `useSession` | Current user, login/logout |
| `useOrg` | Active org from switcher, list memberships |
| `useServer` | Server detail, agent status |
| `useSite` | Site detail, deploy trigger |
| `useJob` | Poll or stream long-running job logs |

## Types

- Align `app/types/` with Go API structs or generated OpenAPI client.
- Zod schemas for forms mirror API request bodies.

## Real-time

- Deploy/provision logs: SSE or WebSocket from API, keyed by job ID.
- Notifications: poll or push into `NotificationsSlideover`.

## Current behavior

Template demo data is served from `server/api/customers.ts`, `members.ts`, etc. Remove or proxy when the API lands ([ROADMAP.md](ROADMAP.md)).
