# Features

Complete Outpost capability catalog. Unless marked **Template demo** (dashboard only), features are **Planned**.

UI mapping: [dashboard/docs/DESIGN.md](../dashboard/docs/DESIGN.md).

## Implementation status (summary)

| Domain | Status | Notes |
|--------|--------|-------|
| Dashboard shell (sidebar, search, notifications UI) | Template demo | Nuxt UI layout; notifications not wired to product events |
| Home / analytics widgets | Template demo | Placeholder for org overview |
| Inbox | Template demo | To be removed or repurposed |
| Customers list/table | Template demo | Pattern for Servers / Sites lists |
| Settings (general, members, notifications, security) | Template demo | Pattern for Organization settings |
| Teams menu | Template demo | Becomes organization switcher |
| Authentication & account | Planned | |
| Organizations | Planned | |
| Subscription & billing | Planned | |
| Organization members (product RBAC) | Planned | |
| Servers (Terraform + manual) | Planned | |
| Sites & stacks | Planned | |
| Deployments & zero-downtime | Planned | |
| Runtime & version management | Planned | |
| Background services & cron | Done | Supervisor + server-wide cron; site-scoped later |
| Logs & observability | Planned | Metrics/alerts: **Future** |
| Nginx & web server | Planned | |
| Database manager | Planned | |
| Domains & SSL | Planned | |
| Site network & access | Planned | |
| Activity & notifications (product) | Planned | |

---

## 1. Authentication & account

**Status:** Planned

### User goals

- Create an account and sign in securely.
- Recover access and verify email.
- Manage personal profile separate from organization settings.

### Key actions

- Register (email/password; OAuth optional **Future**).
- Login, logout, session refresh.
- Forgot password, reset password, verify email.
- Update display name, email, password, optional 2FA (**Future**).

### Notes for backend / agent

- Sessions and tokens issued only by the control plane API; dashboard stores session per browser (httpOnly cookies or bearer + refresh — **TBD** in [ARCHITECTURE.md](ARCHITECTURE.md)).
- Auth routes use a dedicated layout (no dashboard sidebar). See [dashboard/docs/DESIGN.md](../dashboard/docs/DESIGN.md).

---

## 2. Organizations

**Status:** Planned

### User goals

- Work inside a named workspace (organization) with isolated servers, sites, and billing.
- Switch between organizations they belong to.

### Key actions

- On **register**: create default organization named from the user’s display name (editable).
- Create additional organizations (**Planned**; limits may depend on plan).
- Rename org, upload avatar, set slug (**Planned**).
- Transfer ownership, delete org (with safeguards) (**Planned**).

### Notes for backend / agent

- Every server, site, deploy, and billable resource carries `organization_id`.
- UI org switcher maps to `dashboard/app/components/TeamsMenu.vue` (today: mock teams).

---

## 3. Subscription & billing

**Status:** Planned

### User goals

- Understand plan limits and pay for usage.
- Upgrade or downgrade without losing access unexpectedly.

### Key actions

- View current plan, renewal date, usage (servers, sites, seats).
- Add/update payment method, view invoices, download PDF (**Planned**).
- Upgrade, downgrade, cancel at period end, start trial (**Planned**).
- Billing alerts (payment failed, limit reached).

### Role matrix (billing)

| Role | View plan | Change plan | Payment methods | Invoices |
|------|-----------|-------------|-----------------|----------|
| Owner | Yes | Yes | Yes | Yes |
| Admin | Yes | Yes | Yes | Yes |
| Member | Yes (read-only) | No | No | No |
| Viewer | Yes (read-only) | No | No | No |

### Notes for backend / agent

- Billing provider integration (Stripe or similar) lives in API only; dashboard renders state from API.
- Hard limits enforced server-side before create server/site/member.

---

## 4. Organization members

**Status:** Planned (UI pattern: Template demo on `/settings/members`)

### User goals

- Collaborate with teammates with appropriate permissions.

### Key actions

- Invite by email, resend/revoke invite.
- Assign role: **owner**, **admin**, **member**, **viewer**.
- Remove member, suspend access (**Planned**).
- Accept invite flow for new and existing users.

### Role matrix (operations)

| Capability | Owner | Admin | Member | Viewer |
|------------|-------|-------|--------|--------|
| Manage billing | Yes | Yes | No | No |
| Invite/remove members | Yes | Yes | No | No |
| Create/delete servers | Yes | Yes | Yes | No |
| Create/delete sites | Yes | Yes | Yes | No |
| Deploy | Yes | Yes | Yes | No |
| View logs & metrics | Yes | Yes | Yes | Yes |
| Edit firewall / Nginx snippets | Yes | Yes | No | No |
| Database create/drop | Yes | Yes | Yes | No |

Owners cannot be removed without ownership transfer. At least one owner per org.

### Notes for backend / agent

- RBAC enforced on every API call; UI hides actions for UX only.

---

## 5. Servers

**Status:** In progress (manual path working; Terraform **Planned**)

### User goals

- Add machines Outpost can manage—either by provisioning in the cloud or connecting existing VPS / home-lab VM.

### Key actions

#### Provisioned (Terraform)

- Provider: **DigitalOcean**, **AWS** (extensible to others **Future**).
- Wizard: region, instance size, image, SSH keys, tags, name.
- Track job: planning → applying → ready / failed; logs and rollback guidance.
- Store provider credentials at org level (encrypted); rotate (**Planned**).

#### Manual configuration

- Hostname/IP, SSH port, username. **Done** (API + dashboard).
- Auth: SSH private key upload or password (discouraged; document risks). **Done**.
- Test connection over SSH. **Done**.
- Bootstrap/install **agent** (binary + systemd + env) and confirm heartbeat. **Done**.
- Labels, notes. **Done** (maintenance mode **Planned**).
- Re-run bootstrap without recreating the server. **Done** (`POST …/bootstrap`).

#### Server detail

- Status, agent version, last seen. **Partial** (status + version/last seen in API; UI overview still thin).
- CPU/memory/disk (**Future** metrics depth).
- Installed runtimes summary, linked sites count. **Planned** (sites stub only).
- Reinstall agent, decommission server (with confirmation). Reinstall via bootstrap; delete **Done**.

### Notes for backend / agent

- Terraform runs in control plane workers; state stored securely; not executed from the browser (**Planned**).
- Agent install: API SSHs and uploads `agent-linux-{amd64|arm64}`, then agent uses HTTPS + server token. mTLS **TBD**.
- Manual and provisioned servers converge to the same agent-managed model once connected.
- Dev / Proxmox: [DEVELOPMENT.md](DEVELOPMENT.md), [AGENT.md](AGENT.md).

---

## 6. Sites

**Status:** Planned

### User goals

- Host applications on a server with the right stack and configuration.

### Key actions

- Create site on a selected server.
- Choose **site type / stack** (extensible enum):

  | Type | Examples / notes |
  |------|------------------|
  | `static` | HTML/assets, optional build step |
  | `php` | Generic PHP, document root, PHP-FPM pool |
  | `laravel` | PHP + Composer, `public/` root, queue worker optional |
  | `node` | Generic Node app |
  | `nextjs` | Node, build output, SSR or static export mode |
  | `nuxt` | Node, Nitro output |
  | `custom` | User-defined build/start commands |

- Configure: domain (initial), web root, PHP version, Node version, environment variables.
- Git repository URL, branch, deploy key / webhook (**Planned**).
- Link to database (create new or existing).
- Delete site (remove vhost, optional file retention policy).

### Notes for backend / agent

- Site definition is declarative; agent renders Nginx, process manager units, and deploy directories.
- Build runs on server (or **Future**: remote builders).

---

## 7. Deployments

**Status:** Planned

### User goals

- Ship code safely with visible history and logs.

### Key actions

- Trigger deploy manually or via Git push webhook.
- View live build log stream, artifact path, duration, commit/ref.
- Release history: who deployed, when, success/failure.
- Rollback to previous release (**Planned**).
- Configure deploy strategy:

  - **Standard**: brief downtime acceptable.
  - **Zero-downtime**: symlink or blue/green directory swap, health check, Nginx reload.

### Zero-downtime (agent behavior)

- Prepare new release directory → build → health check on internal URL/port → atomically switch `current` symlink → reload Nginx → retire old release after TTL.
- UI shows strategy, steps, and failure point.

### Notes for backend / agent

- Deploy jobs queued; dashboard polls or subscribes via WebSocket/SSE (**TBD**).
- No arbitrary shell from UI; allowed commands defined per stack template.

---

## 8. Runtime & version management

**Status:** Partial — stack provision + inventory + thin Runtimes UI **Done**; per-site bind and rich management UI **Planned**.

### User goals

- Install and switch language/runtime versions without SSH expertise.

### Key actions

- **Node**: install versions via fnm on the host; set server default; per-site override **Planned**.
- **PHP**: multi-version via Ondřej packages + FPM; set server default; per-site pool **Planned**.
- **Redis / MySQL / MariaDB / PostgreSQL**: optional addons at provision time; bind/persistence/DB CRUD **Planned**.
- Thin **Runtimes** tab: list inventory, install more versions, set default (contracts ready for richer UI).

### Notes for backend / agent

- Version installs are long-running agent jobs (`server.provision`, `runtime.*.install`).
- Dashboard reads `runtimes` inventory from the control plane (seeded from job payloads).

---

## 9. Background services & cron

**Status:** Done (server-wide; site-scoped **Planned**)

### User goals

- Run queue workers, custom daemons, and scheduled tasks reliably.

### Key actions

- List services (Supervisor): name, status, command, auto-restart.
- Start, stop, restart; view recent stdout/stderr.
- Cron: CRUD entries with schedule validator, timezone, command, user context.
- Scope: **server-wide** (site-specific e.g. Laravel scheduler — Planned).

### Notes for backend / agent

- Job kinds: `service.apply`, `service.remove`, `service.control`, `cron.sync`.
- Agent installs Supervisor if missing; writes `/etc/supervisor/conf.d/sa-<name>.conf`.
- Agent validates cron expressions before write; materializes `/etc/cron.d/sa`.
- Provision creates managed Unix user `sm` (sudo); password revealed once with addon secrets. Services/cron default to that user.
- Service/cron definitions stored in control plane; agent materializes config files.
- API gates with `organization.servers.view` / `update` (same as runtimes).

---

## 10. Logs & observability

**Status:** Planned (metrics/alerts: **Future**)

### User goals

- Debug production issues from the dashboard.

### Key actions

- **Server logs**: syslog aggregate or agent tail paths (**TBD**).
- **Site logs**: Nginx access/error, PHP-FPM, app log paths per stack.
- Live tail, pause, download chunk, filter by level/text/time range.
- Retention policy per plan (**Planned**).

### Future

- Metric charts (CPU, memory, request rate), alerting rules, PagerDuty/Slack integrations.

---

## 11. Nginx & web server

**Status:** Planned

### User goals

- HTTPS and routing configured automatically from site settings.

### Key actions

- Auto-generate server block: `server_name`, root/upstream, SSL certs, gzip, basic security headers.
- Preview effective config (read-only), run `nginx -t` before apply.
- Advanced: include custom snippets (admin-only), per-location rules.
- Coordinate reload with zero-downtime deploys.

### Notes for backend / agent

- Agent owns file paths under `/etc/nginx/` (or distro equivalent); never expose raw editor to all roles.

---

## 12. Database manager

**Status:** Planned

### User goals

- Create databases and users for applications without manual SQL shell.

### Key actions

- Create database, create user, grant privileges scoped to site.
- Show connection string / `.env` hints (password reveal gated).
- List databases on server; drop with typed confirmation.
- Backup now, restore from backup, scheduled backups (**Planned**).

### Notes for backend / agent

- Privilege separation: agent uses limited admin credentials stored encrypted.
- Backups stored on server disk or S3-compatible storage (**Future**).

---

## 13. Domains & SSL

**Status:** Planned

### User goals

- Serve sites on real domains with valid TLS.

### Key actions

- Primary domain + aliases per site.
- DNS instructions (A/AAAA/CNAME), verification status.
- Issue/renew Let’s Encrypt (HTTP-01, DNS-01 **Planned**).
- Force HTTPS, HSTS optional (**Planned**), cert expiry notifications.

### Notes for backend / agent

- Renewal jobs run on schedule; failures surface in notifications.

---

## 14. Site network & access

**Status:** Planned

### User goals

- Control who can reach the site and how traffic is shaped.

### Key actions

- **Firewall**: allow/deny IP ranges, optional port rules at server or site scope.
- **Redirects**: 301/302 rules, apex ↔ www.
- **Access protection**: HTTP basic auth or password gate for staging.
- Rate limiting (**Future**).

### Notes for backend / agent

- Firewall rules applied via agent (iptables/nftables or cloud firewall API **TBD**).
- Redirects may live in Nginx config generated from site network settings.

---

## 15. Notifications & activity

**Status:** Planned (slideover UI: Template demo)

### User goals

- See what happened across the organization without opening each server.

### Key actions

- Activity feed: deploys, provision jobs, member invites, cert expiry, billing events.
- Notification preferences per channel (email, in-app) — extend template `/settings/notifications`.
- Mark read, clear all; link to resource (site, server, invoice).

### Notes for backend / agent

- Events emitted by API workers; dashboard consumes `/notifications` or WebSocket stream.
