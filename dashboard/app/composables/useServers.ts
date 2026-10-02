import { ApiError } from '~/utils/api'
import {
  mapServer,
  type CreateServerInput,
  type Server,
  type ServerApi
} from '~/utils/servers'

function serversState() {
  return {
    servers: useState<Server[]>('org-servers', () => []),
    ready: useState('org-servers-ready', () => false),
    pending: useState('org-servers-pending', () => false),
    loadedOrgId: useState<number | null>('org-servers-org', () => null)
  }
}

/** Safe to call from session clear — no useRoute. */
export function clearServers() {
  const { servers, ready, loadedOrgId } = serversState()
  servers.value = []
  ready.value = false
  loadedOrgId.value = null
}

export function useServers() {
  const { api } = useApi()
  const { selected } = useOrganizations()
  const { currentPlan, setUsage } = useBilling()
  const { servers, ready, pending, loadedOrgId } = serversState()

  const serverLimit = computed(() =>
    currentPlan.value.limits.find(l => l.key === 'servers')?.max ?? 1
  )

  const atCapacity = computed(() =>
    servers.value.length >= serverLimit.value
  )

  function orgPath(suffix = '') {
    const org = selected.value
    if (!org) {
      throw new ApiError('No organization selected', 400)
    }
    return `/api/v1/organizations/${org.id}/servers${suffix}`
  }

  function syncUsage(list: Server[]) {
    setUsage({
      servers: list.length,
      sites: list.reduce((sum, s) => sum + s.sites, 0)
    })
  }

  async function fetchAll(force = false) {
    const org = selected.value
    if (!org) {
      clearServers()
      return
    }

    if (!force && ready.value && loadedOrgId.value === org.id) {
      return
    }

    pending.value = true
    try {
      const res = await api<ServerApi[]>(orgPath())
      servers.value = res.data.map(mapServer)
      loadedOrgId.value = org.id
      ready.value = true
      syncUsage(servers.value)
    } finally {
      pending.value = false
    }
  }

  function sshBody(input: CreateServerInput) {
    const body: Record<string, unknown> = {
      ip: input.ip.trim(),
      ssh_port: input.sshPort,
      ssh_user: input.sshUser.trim() || 'root',
      auth_method: input.authMethod
    }
    if (input.authMethod === 'key') {
      body.private_key = input.privateKey || ''
      body.passphrase = input.passphrase || ''
    } else {
      body.password = input.password || ''
    }
    return body
  }

  async function testConnection(input: CreateServerInput) {
    const res = await api<{ ok: boolean, message: string }>(orgPath('/test-connection'), {
      method: 'POST',
      body: sshBody(input)
    })
    return res.data
  }

  async function create(input: CreateServerInput) {
    if (servers.value.length >= serverLimit.value) {
      throw new ApiError(
        `Plan limit reached (${serverLimit.value} servers on ${currentPlan.value.name}).`,
        422
      )
    }

    const body: Record<string, unknown> = {
      ...sshBody(input),
      hostname: input.hostname.trim(),
      region: input.region?.trim() || '',
      labels: input.labels,
      notes: input.notes?.trim() || ''
    }

    const res = await api<ServerApi>(orgPath(), {
      method: 'POST',
      body
    })
    const server = mapServer(res.data)
    servers.value = [server, ...servers.value]
    syncUsage(servers.value)
    return server
  }

  async function bootstrap(id: number) {
    const res = await api<{ ok: boolean, message: string }>(orgPath(`/${id}/bootstrap`), {
      method: 'POST'
    })
    return res.data
  }

  async function remove(id: number) {
    const existing = servers.value.find(s => s.id === id)
    if (!existing) {
      throw new ApiError('Server not found', 404)
    }

    const res = await api<{ agent_removed?: boolean, cleanup?: string }>(orgPath(`/${id}`), { method: 'DELETE' })
    servers.value = servers.value.filter(s => s.id !== id)
    syncUsage(servers.value)
    return { server: existing, message: res.message, cleanup: res.data }
  }

  function getById(id: string | number) {
    const n = typeof id === 'number' ? id : Number(id)
    if (!Number.isFinite(n)) {
      return undefined
    }
    return servers.value.find(s => s.id === n)
  }

  function setAgent(id: string | number, agent: Server['agent']) {
    const server = getById(id)
    if (!server) {
      return { ok: false as const, error: 'Server not found.' }
    }
    server.agent = agent
    return { ok: true as const, server }
  }

  function setSitesCount(id: string | number, sites: number) {
    const server = getById(id)
    if (!server) {
      return { ok: false as const, error: 'Server not found.' }
    }
    server.sites = sites
    return { ok: true as const, server }
  }

  // Org switch remounts /o/:orgSlug/* pages which call fetchAll(); no shared watcher
  // (createSharedComposable + useRoute is SSR-fragile in Nuxt).
  if (selected.value?.id && selected.value.id !== loadedOrgId.value) {
    ready.value = false
  }

  return {
    servers,
    ready,
    pending,
    serverLimit,
    atCapacity,
    fetchAll,
    testConnection,
    create,
    bootstrap,
    remove,
    getById,
    setAgent,
    setSitesCount,
    clearServers
  }
}
