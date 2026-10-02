# Dashboard roadmap

Frontend timeline for **Phase 0–4**. API, agent, and Terraform work are summarized in [../../docs/ROADMAP.md](../../docs/ROADMAP.md); this file lists **dashboard deliverables only**.

**Current phase:** 2 — servers wired; sites/deploys remaining.

---

## Phase 0 — Foundation

- [x] Frontend docs (`docs/SCOPE`, `DESIGN`, `API`, this file)
- [x] Servers / sites routes under `/o/:orgSlug/...`

**Exit:** DESIGN agreed; SCOPE reflects repo state.

---

## Phase 1 — Identity & organization UI

- [x] Auth layout — login, register
- [x] Org list + switch (API-backed)
- [x] `/organization/*` settings (general, members, …)
- [x] Zod forms + types for auth/org
- [x] Members / invites via API

**Depends on:** API auth + org endpoints — done.

---

## Phase 2 — Servers, sites, deploy UI

- [x] `/servers` list + manual add form
- [x] Test connection + create (API SSH + agent bootstrap)
- [x] Server detail shell + agent status from API
- [ ] Create site wizard (static, php, node minimum) — still stub
- [ ] Deploy action + log viewer (real stream)
- [ ] Replace provision stub with API job events

**Depends on:** API + agent Phase 2 (bootstrap/heartbeat done; sites/deploys pending).

---

## Phase 3 — Provision & billing UI

- [ ] Terraform wizards (DO, AWS)
- [ ] Provision job status UI
- [ ] `/organization/billing` — plan, usage, invoices (beyond stub)

**Depends on:** API + TF Phase 3.

---

## Phase 4 — Operations UI depth

- [ ] Runtimes, cron, SSL, DB, network screens
- [ ] Align with agent capabilities as they ship

See [DESIGN.md](DESIGN.md) for IA.
