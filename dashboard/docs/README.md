# Dashboard documentation

Frontend-only docs for the Nuxt app in this folder.

**Product features, backend architecture, agent, and monorepo phases:** [../../docs/](../../docs/) (start with [DEVELOPMENT.md](../../docs/DEVELOPMENT.md) / [AGENT.md](../../docs/AGENT.md) for how servers connect).

## Reading order

1. **[SCOPE.md](SCOPE.md)** — What this app owns; template demo vs product UI.
2. **[AUTH.md](AUTH.md)** — Session, invites, orgs, servers abilities, Filament-style `useGate`.
3. **[DESIGN.md](DESIGN.md)** — Information architecture, user flows, Nuxt UI patterns.
4. **[API.md](API.md)** — Nitro proxy, composables (`useServers`, …).
5. **[ROADMAP.md](ROADMAP.md)** — Frontend timeline and migration checklist.

Agent-oriented rules: [../AGENTS.md](../AGENTS.md).

## Conventions

| Label | Meaning |
|-------|---------|
| **Template demo** | Upstream Nuxt dashboard sample; not Outpost product logic |
| **Planned** | Specified UI work; not implemented |
| **Future** | Deferred |

Paths are relative to `dashboard/` (e.g. `app/layouts/default.vue`).

## Contributing

- Auth / permission changes → update [AUTH.md](AUTH.md) and [AGENTS.md](../AGENTS.md).
- UI and routing changes → update DESIGN and ROADMAP here.
- New product capabilities → update [../../docs/FEATURES.md](../../docs/FEATURES.md) first, then DESIGN mapping.
