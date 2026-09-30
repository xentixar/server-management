import type { CreateSiteInput, Site } from '~/utils/sites'
import { defaultWebRoot } from '~/utils/sites'

// ponytail: fixed clock so SSR stub HTML matches client; swap for API later
const ANCHOR = Date.UTC(2026, 8, 30, 12, 0, 0)

function daysAgo(days: number) {
  return new Date(ANCHOR - days * 86400000).toISOString()
}

function sampleSites(): Site[] {
  return [{
    id: 'site_1',
    serverId: 'srv_1',
    name: 'Marketing',
    domain: 'acme.dev',
    stack: 'static',
    webRoot: 'public',
    status: 'active',
    createdAt: daysAgo(10)
  }, {
    id: 'site_2',
    serverId: 'srv_1',
    name: 'API',
    domain: 'api.acme.dev',
    stack: 'nuxt',
    webRoot: '.output/public',
    status: 'active',
    gitRepo: 'git@github.com:acme/api.git',
    gitBranch: 'main',
    createdAt: daysAgo(9)
  }, {
    id: 'site_3',
    serverId: 'srv_1',
    name: 'Docs',
    domain: 'docs.acme.dev',
    stack: 'static',
    webRoot: 'dist',
    status: 'active',
    createdAt: daysAgo(7)
  }, {
    id: 'site_4',
    serverId: 'srv_1',
    name: 'Status',
    domain: 'status.acme.dev',
    stack: 'static',
    webRoot: 'public',
    status: 'disabled',
    createdAt: daysAgo(5)
  }, {
    id: 'site_5',
    serverId: 'srv_2',
    name: 'Shop',
    domain: 'shop.acme.dev',
    stack: 'laravel',
    webRoot: 'public',
    status: 'active',
    gitRepo: 'git@github.com:acme/shop.git',
    gitBranch: 'main',
    createdAt: daysAgo(6)
  }, {
    id: 'site_6',
    serverId: 'srv_2',
    name: 'Admin',
    domain: 'admin.acme.dev',
    stack: 'nextjs',
    webRoot: '.next',
    status: 'deploying',
    gitRepo: 'git@github.com:acme/admin.git',
    gitBranch: 'main',
    createdAt: daysAgo(4)
  }, {
    id: 'site_7',
    serverId: 'srv_2',
    name: 'Blog',
    domain: 'blog.acme.dev',
    stack: 'php',
    webRoot: 'public',
    status: 'failed',
    createdAt: daysAgo(2)
  }, {
    id: 'site_8',
    serverId: 'srv_3',
    name: 'Worker UI',
    domain: 'jobs.acme.dev',
    stack: 'node',
    webRoot: 'public',
    status: 'active',
    gitRepo: 'git@github.com:acme/jobs.git',
    gitBranch: 'main',
    createdAt: daysAgo(1)
  }]
}

function ensure(orgId: string | number): Site[] {
  const key = String(orgId)
  const store = useState<Record<string, Site[]>>('sites-by-org', () => ({}))
  if (!store.value[key]) {
    store.value = { ...store.value, [key]: sampleSites() }
  }
  return store.value[key]!
}

export function useSites() {
  const { selected } = useOrganizations()
  const { servers, getById, setSitesCount } = useServers()
  const { setUsage } = useBilling()

  const sites = computed(() => ensure(selected.value?.id ?? '_'))

  function syncCounts() {
    const list = ensure(selected.value?.id ?? '_')
    for (const server of servers.value) {
      setSitesCount(server.id, list.filter(s => s.serverId === server.id).length)
    }
    setUsage({ sites: list.length })
  }

  function create(input: CreateSiteInput) {
    const list = ensure(selected.value?.id ?? '_')
    const domain = input.domain.trim().toLowerCase()
    if (list.some(s => s.domain.toLowerCase() === domain)) {
      return { ok: false as const, error: 'That domain is already used in this organization.' }
    }
    if (!getById(input.serverId)) {
      return { ok: false as const, error: 'Server not found.' }
    }

    const site: Site = {
      id: `site_${Date.now()}`,
      serverId: input.serverId,
      name: input.name.trim(),
      domain,
      stack: input.stack,
      webRoot: input.webRoot.trim() || defaultWebRoot(input.stack),
      status: 'active',
      gitRepo: input.gitRepo?.trim() || undefined,
      gitBranch: input.gitBranch?.trim() || undefined,
      createdAt: new Date().toISOString()
    }

    list.unshift(site)
    syncCounts()
    return { ok: true as const, site }
  }

  function remove(id: string) {
    const list = ensure(selected.value?.id ?? '_')
    const index = list.findIndex(s => s.id === id)
    if (index === -1) {
      return { ok: false as const, error: 'Site not found.' }
    }
    const [removed] = list.splice(index, 1)
    syncCounts()
    return { ok: true as const, site: removed! }
  }

  watch(() => selected.value?.id, () => {
    syncCounts()
  }, { immediate: true })

  return {
    sites,
    create,
    remove
  }
}
