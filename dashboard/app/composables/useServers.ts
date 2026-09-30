import type { CreateServerInput, Server } from '~/utils/servers'

// ponytail: fixed clock so SSR stub HTML matches client; swap for API later
const ANCHOR = Date.UTC(2026, 8, 30, 12, 0, 0)

function daysAgo(days: number) {
  return new Date(ANCHOR - days * 86400000).toISOString()
}

function sampleServers(): Server[] {
  return [{
    id: 'srv_1',
    hostname: 'edge-01.acme.dev',
    ip: '104.248.10.11',
    sshPort: 22,
    sshUser: 'root',
    authMethod: 'key',
    hasKeyUploaded: true,
    region: 'nyc1',
    labels: ['edge'],
    agent: 'online',
    sites: 4,
    createdAt: daysAgo(12)
  }, {
    id: 'srv_2',
    hostname: 'app-02.acme.dev',
    ip: '164.92.44.20',
    sshPort: 22,
    sshUser: 'deploy',
    authMethod: 'key',
    hasKeyUploaded: true,
    region: 'sfo3',
    labels: ['app'],
    agent: 'degraded',
    sites: 3,
    createdAt: daysAgo(8)
  }, {
    id: 'srv_3',
    hostname: 'worker-03.acme.dev',
    ip: '167.99.88.30',
    sshPort: 22,
    sshUser: 'root',
    authMethod: 'key',
    hasKeyUploaded: true,
    region: 'ams3',
    labels: ['worker'],
    agent: 'offline',
    sites: 1,
    createdAt: daysAgo(3)
  }]
}

function ensure(orgId: string | number): Server[] {
  const key = String(orgId)
  const store = useState<Record<string, Server[]>>('servers-by-org', () => ({}))
  if (!store.value[key]) {
    store.value = { ...store.value, [key]: sampleServers() }
  }
  return store.value[key]!
}

export function useServers() {
  const { selected } = useOrganizations()
  const { currentPlan, setUsage } = useBilling()

  const servers = computed(() => ensure(selected.value?.id ?? '_'))

  const serverLimit = computed(() =>
    currentPlan.value.limits.find(l => l.key === 'servers')?.max ?? 1
  )

  const atCapacity = computed(() =>
    servers.value.length >= serverLimit.value
  )

  function syncUsage(list: Server[]) {
    setUsage({
      servers: list.length,
      sites: list.reduce((sum, s) => sum + s.sites, 0)
    })
  }

  function create(input: CreateServerInput) {
    const list = ensure(selected.value?.id ?? '_')
    if (list.length >= serverLimit.value) {
      return {
        ok: false as const,
        error: `Plan limit reached (${serverLimit.value} servers on ${currentPlan.value.name}).`
      }
    }

    const hostname = input.hostname.trim()
    if (list.some(s => s.hostname.toLowerCase() === hostname.toLowerCase())) {
      return { ok: false as const, error: 'A server with that hostname already exists.' }
    }

    const server: Server = {
      id: `srv_${Date.now()}`,
      hostname,
      ip: input.ip.trim(),
      sshPort: input.sshPort,
      sshUser: input.sshUser.trim() || 'root',
      authMethod: input.authMethod,
      hasKeyUploaded: input.authMethod === 'key' ? input.hasKeyUploaded : false,
      region: input.region?.trim() || undefined,
      labels: input.labels,
      notes: input.notes?.trim() || undefined,
      agent: 'pending',
      sites: 0,
      createdAt: new Date().toISOString()
    }

    list.unshift(server)
    syncUsage(list)
    return { ok: true as const, server }
  }

  function remove(id: string) {
    const list = ensure(selected.value?.id ?? '_')
    const index = list.findIndex(s => s.id === id)
    if (index === -1) {
      return { ok: false as const, error: 'Server not found.' }
    }
    const [removed] = list.splice(index, 1)
    syncUsage(list)
    return { ok: true as const, server: removed! }
  }

  function getById(id: string) {
    return ensure(selected.value?.id ?? '_').find(s => s.id === id)
  }

  function setAgent(id: string, agent: Server['agent']) {
    const server = getById(id)
    if (!server) {
      return { ok: false as const, error: 'Server not found.' }
    }
    server.agent = agent
    return { ok: true as const, server }
  }

  function setSitesCount(id: string, sites: number) {
    const server = getById(id)
    if (!server) {
      return { ok: false as const, error: 'Server not found.' }
    }
    server.sites = sites
    return { ok: true as const, server }
  }

  watch(() => selected.value?.id ?? '_', (id) => {
    syncUsage(ensure(id))
  }, { immediate: true })

  return {
    servers,
    serverLimit,
    atCapacity,
    create,
    remove,
    getById,
    setAgent,
    setSitesCount
  }
}
