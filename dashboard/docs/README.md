# Dashboard documentation

Frontend-only docs for the Nuxt app in this folder.

**Product features, backend architecture, and monorepo phases:** [../../docs/](../../docs/).

## Reading order

1. **[SCOPE.md](SCOPE.md)** — What this app owns today; template demo vs planned UI.
2. **[DESIGN.md](DESIGN.md)** — Information architecture, user flows, Nuxt UI patterns.
3. **[API.md](API.md)** — Planned integration with the Go control plane.
4. **[ROADMAP.md](ROADMAP.md)** — Frontend timeline and template migration checklist.

## Conventions

| Label | Meaning |
|-------|---------|
| **Template demo** | Upstream Nuxt dashboard sample; not Outpost product logic |
| **Planned** | Specified UI work; not implemented |
| **Future** | Deferred |

Paths are relative to `dashboard/` (e.g. `app/layouts/default.vue`).

## Contributing

- UI and routing changes → update DESIGN and ROADMAP here.
- New product capabilities → update [../../docs/FEATURES.md](../../docs/FEATURES.md) first, then DESIGN mapping.
