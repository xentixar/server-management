# Dashboard roadmap

Frontend timeline for **Phase 0–4**. API, agent, and Terraform work are summarized in [../../docs/ROADMAP.md](../../docs/ROADMAP.md); this file lists **dashboard deliverables only**.

**Current phase:** 0 — docs and IA; no auth or API wiring.

---

## Phase 0 — Foundation (now)

- [x] Frontend docs (`docs/SCOPE`, `DESIGN`, `API`, this file)
- [ ] Optional: rename sidebar labels; stub `/servers`, `/sites`
- [ ] Optional: remove or hide template demo routes when stubs exist

**Exit:** DESIGN agreed; SCOPE reflects repo state.

---

## Phase 1 — Identity & organization UI

- [ ] `layouts/auth.vue` — login, register, forgot password
- [ ] Wire `TeamsMenu` to org list + switch (API-backed)
- [ ] `/organization/*` from settings pattern (general, members, notifications, security)
- [ ] Zod forms + types for auth/org
- [ ] Remove or proxy `server/api/members.ts`

**Depends on:** API auth + org endpoints ([../../docs/ROADMAP.md](../../docs/ROADMAP.md) Phase 1).

---

## Phase 2 — Servers, sites, deploy UI

- [ ] `/servers` list + manual add form (from `customers.vue` pattern)
- [ ] Server overview — agent status, maintenance banner
- [ ] Create site wizard (static, php, node minimum)
- [ ] Deploy action + log viewer component
- [ ] Server/site contextual sub-nav per [DESIGN.md](DESIGN.md)

**Depends on:** API + agent Phase 2.

---

## Phase 3 — Provision & billing UI

- [ ] Terraform wizards (DO, AWS)
- [ ] Provision job status UI
- [ ] `/organization/billing` — plan, usage, invoices

**Depends on:** API + TF Phase 3.

---

## Phase 4 — Operations UI

- [ ] Zero-downtime deploy settings and step indicator
- [ ] Runtimes, services, cron editors
- [ ] Log tail components (server + site)
- [ ] Domains/SSL, database manager, network screens
- [ ] Product notifications wired to API events

**Depends on:** API + agent Phase 4.

---

## Template migration checklist

- [ ] `app/pages/index.vue` — org dashboard (remove demo sales/mail CTAs)
- [ ] `app/pages/customers.vue` — move pattern to `app/pages/servers/index.vue`
- [ ] `app/pages/inbox.vue` — delete or repurpose
- [ ] `app/pages/settings.vue` — rename route group to `/organization`
- [ ] `app/components/TeamsMenu.vue` — real orgs from API
- [ ] `app/composables/useDashboard.ts` — shortcuts for new routes
- [ ] `server/api/*` — delete mocks or proxy to Go API
- [ ] `app/types/index.d.ts` — domain types instead of demo User/Customer

---

## Future (dashboard)

- OAuth login buttons (API-driven)
- Metrics charts on server overview
- Audit log table under organization security
- Public API token management UI
