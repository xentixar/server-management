import { ApiError } from '~/utils/api'
import {
  mapJob,
  mapRuntimes,
  type AgentJob,
  type AgentJobApi,
  type ProvisionInput,
  type RuntimeCatalog,
  type RuntimeName,
  type ServerRuntimes
} from '~/utils/runtimes'

export function useServerRuntimes(serverId: MaybeRefOrGetter<string | number>) {
  const { api } = useApi()
  const { selected } = useOrganizations()
  const { fetchAll, getById } = useServers()

  function basePath() {
    const org = selected.value
    if (!org) {
      throw new ApiError('No organization selected', 400)
    }
    const id = toValue(serverId)
    return `/api/v1/organizations/${org.id}/servers/${id}`
  }

  async function provision(input: ProvisionInput) {
    const res = await api<AgentJobApi>(`${basePath()}/provision`, {
      method: 'POST',
      body: {
        php_versions: input.phpVersions,
        node_versions: input.nodeVersions,
        addons: input.addons
      }
    })
    return mapJob(res.data)
  }

  async function getRuntimes() {
    const res = await api<ServerRuntimes>(`${basePath()}/runtimes`)
    return mapRuntimes(res.data)
  }

  async function getCatalog() {
    const res = await api<RuntimeCatalog>(`${basePath()}/runtimes/catalog`)
    return res.data
  }

  async function installRuntime(runtime: RuntimeName, version: string, setDefault = false) {
    const res = await api<AgentJobApi>(`${basePath()}/runtimes`, {
      method: 'POST',
      body: {
        runtime,
        version,
        set_default: setDefault
      }
    })
    return mapJob(res.data)
  }

  async function setDefault(runtime: RuntimeName, version: string) {
    const res = await api<ServerRuntimes>(`${basePath()}/runtimes/default`, {
      method: 'PATCH',
      body: { runtime, version }
    })
    return mapRuntimes(res.data)
  }

  async function getJob(jobId: number) {
    const res = await api<AgentJobApi>(`${basePath()}/jobs/${jobId}`)
    return mapJob(res.data)
  }

  async function getLatestJob(kind?: string) {
    const query = kind ? `?kind=${encodeURIComponent(kind)}` : ''
    try {
      const res = await api<AgentJobApi>(`${basePath()}/jobs/latest${query}`)
      return mapJob(res.data)
    } catch (error) {
      if (error instanceof ApiError && error.status === 404) {
        return null
      }
      throw error
    }
  }

  async function pollJob(
    jobId: number,
    opts?: {
      intervalMs?: number
      maxAttempts?: number
      onTick?: (job: AgentJob) => void
    }
  ) {
    const intervalMs = opts?.intervalMs ?? 3000
    const maxAttempts = opts?.maxAttempts ?? 240 // ~12m for apt + fnm
    let last: AgentJob | undefined
    for (let i = 0; i < maxAttempts; i++) {
      last = await getJob(jobId)
      opts?.onTick?.(last)
      if (last.status === 'succeeded' || last.status === 'failed') {
        await fetchAll(true)
        return last
      }
      await new Promise(resolve => setTimeout(resolve, intervalMs))
    }
    await fetchAll(true)
    return last
  }

  function server() {
    return getById(toValue(serverId))
  }

  return {
    server,
    provision,
    getRuntimes,
    getCatalog,
    installRuntime,
    setDefault,
    getJob,
    getLatestJob,
    pollJob,
    fetchAll
  }
}
