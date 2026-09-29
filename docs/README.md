# Outpost — platform documentation

System-wide product and engineering docs for the **server-management** monorepo. The Nuxt app has its own docs in [dashboard/docs/](../dashboard/docs/).

## Reading order

1. **[FEATURES.md](FEATURES.md)** — Full product catalog and status (auth through operations).
2. **[ARCHITECTURE.md](ARCHITECTURE.md)** — Control plane, agent, Terraform, Postgres, Redis, trust boundaries.
3. **[ROADMAP.md](ROADMAP.md)** — Phased delivery across dashboard, API, agent, and Terraform.

Then read [dashboard/docs/DESIGN.md](../dashboard/docs/DESIGN.md) for UI flows and template mapping.

## Conventions

| Label | Meaning |
|-------|---------|
| **Planned** | Specified; not implemented |
| **Future** | Deferred (see ROADMAP) |

Dashboard-specific **Template demo** labels are defined in [dashboard/docs/SCOPE.md](../dashboard/docs/SCOPE.md).

## Contributing

- Platform behavior and backend notes → update files in this folder.
- Routes, layouts, and UI patterns → [dashboard/docs/](../dashboard/docs/).
- When a feature ships, update FEATURES status and the relevant ROADMAP phase in both places if the dashboard deliverable changed.
