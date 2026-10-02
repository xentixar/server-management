# Outpost — platform documentation

System-wide product and engineering docs for the **server-management** monorepo. The Nuxt app has its own docs in [dashboard/docs/](../dashboard/docs/).

## Reading order

1. **[FEATURES.md](FEATURES.md)** — Product catalog and status.
2. **[ARCHITECTURE.md](ARCHITECTURE.md)** — Control plane, agent, trust boundaries, flows.
3. **[DEVELOPMENT.md](DEVELOPMENT.md)** — Run locally; test on Proxmox / SSH hosts.
4. **[AGENT.md](AGENT.md)** — Agent protocol, install layout, bootstrap.
5. **[ROADMAP.md](ROADMAP.md)** — Phased delivery.

Then [dashboard/docs/DESIGN.md](../dashboard/docs/DESIGN.md) for UI flows.

## Component READMEs

| Path | Doc |
|------|-----|
| [../README.md](../README.md) | Monorepo overview + quick start |
| [../api/README.md](../api/README.md) | Control plane |
| [../agent/README.md](../agent/README.md) | On-server agent |
| [../dashboard/README.md](../dashboard/README.md) | Nuxt app |

## Conventions

| Label | Meaning |
|-------|---------|
| **Done / In progress** | Reflects current repo (see ROADMAP) |
| **Planned** | Specified; not implemented |
| **Future** | Deferred |

Dashboard **Template demo** labels: [dashboard/docs/SCOPE.md](../dashboard/docs/SCOPE.md).

## Contributing

- Platform behavior / backend → this folder + `api/` / `agent/` READMEs.
- Routes, layouts, UI patterns → [dashboard/docs/](../dashboard/docs/).
- When a feature ships, update FEATURES status and ROADMAP (root + dashboard if UI changed).
