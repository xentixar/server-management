import type { CreateServerInput, Server } from '~/utils/servers'

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
    createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 12).toISOString()
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
    createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 8).toISOString()
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
    createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 3).toISOString()
  }]
}

// ponytail: local stub until servers API exists; keyed by org id
const byOrg = reactive<Record<string, Server[]>>({})

function ensure(orgId: string): Server[] {
  if (!byOrg[orgId]) {
    byOrg[orgId] = orgId === '1' ? [] : sampleServers()
  }
  return byOrg[orgId]!
}

export function useServers() {
  const { selected } = useOrganizations()
  const { currentPlan, setUsage } = useBilling()

  const servers = computed(() => ensure(selected.value.id))

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
    const list = ensure(selected.value.id)
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
    const list = ensure(selected.value.id)
    const index = list.findIndex(s => s.id === id)
    if (index === -1) {
      return { ok: false as const, error: 'Server not found.' }
    }
    const [removed] = list.splice(index, 1)
    syncUsage(list)
    return { ok: true as const, server: removed! }
  }

  watch(() => selected.value.id, (id) => {
    syncUsage(ensure(id))
  }, { immediate: true })

  return {
    servers,
    serverLimit,
    atCapacity,
    create,
    remove
  }
}
