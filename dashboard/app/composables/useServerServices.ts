import { ApiError } from '~/utils/api'
import { mapJob, type AgentJob, type AgentJobApi } from '~/utils/runtimes'
import {
  mapService,
  mapServiceJobResponse,
  type CreateServiceInput,
  type ServerService,
  type ServerServiceApi,
  type ServiceControlAction,
  type ServiceJobResponse,
  type UpdateServiceInput
} from '~/utils/processes'

export function useServerServices(serverId: MaybeRefOrGetter<string | number>) {
  const { api } = useApi()
  const { selected } = useOrganizations()
  const services = ref<ServerService[]>([])

  function basePath() {
    const org = selected.value
    if (!org) {
      throw new ApiError('No organization selected', 400)
    }
    return `/api/v1/organizations/${org.id}/servers/${toValue(serverId)}`
  }

  async function fetchAll() {
    const res = await api<ServerServiceApi[]>(`${basePath()}/services`)
    services.value = (res.data ?? []).map(mapService)
    return services.value
  }

  async function create(input: CreateServiceInput): Promise<ServiceJobResponse> {
    const res = await api<{ job: AgentJobApi, service: ServerServiceApi }>(`${basePath()}/services`, {
      method: 'POST',
      body: {
        name: input.name,
        command: input.command,
        directory: input.directory || '',
        user: input.user || undefined,
        processes: input.processes ?? 1,
        auto_restart: input.autoRestart ?? true,
        autostart: input.autostart ?? true,
        enabled: input.enabled ?? true
      }
    })
    return mapServiceJobResponse(res.data)
  }

  async function update(id: string, input: UpdateServiceInput): Promise<ServiceJobResponse> {
    const body: Record<string, unknown> = {}
    if (input.command !== undefined) body.command = input.command
    if (input.directory !== undefined) body.directory = input.directory
    if (input.user !== undefined) body.user = input.user
    if (input.processes !== undefined) body.processes = input.processes
    if (input.autoRestart !== undefined) body.auto_restart = input.autoRestart
    if (input.autostart !== undefined) body.autostart = input.autostart
    if (input.enabled !== undefined) body.enabled = input.enabled

    const res = await api<{ job: AgentJobApi, service: ServerServiceApi }>(
      `${basePath()}/services/${encodeURIComponent(id)}`,
      { method: 'PATCH', body }
    )
    return mapServiceJobResponse(res.data)
  }

  async function remove(id: string): Promise<ServiceJobResponse> {
    const res = await api<{ job: AgentJobApi, service: ServerServiceApi }>(
      `${basePath()}/services/${encodeURIComponent(id)}`,
      { method: 'DELETE' }
    )
    return mapServiceJobResponse(res.data)
  }

  async function control(id: string, action: ServiceControlAction): Promise<ServiceJobResponse> {
    const res = await api<{ job: AgentJobApi, service: ServerServiceApi }>(
      `${basePath()}/services/${encodeURIComponent(id)}/control`,
      { method: 'POST', body: { action } }
    )
    return mapServiceJobResponse(res.data)
  }

  async function getJob(jobId: string): Promise<AgentJob> {
    const res = await api<AgentJobApi>(`${basePath()}/jobs/${encodeURIComponent(jobId)}`)
    return mapJob(res.data)
  }

  async function pollJob(
    jobId: string,
    opts?: {
      intervalMs?: number
      maxAttempts?: number
      onTick?: (job: AgentJob) => void
    }
  ) {
    const intervalMs = opts?.intervalMs ?? 2000
    const maxAttempts = opts?.maxAttempts ?? 90
    let last: AgentJob | undefined
    for (let i = 0; i < maxAttempts; i++) {
      last = await getJob(jobId)
      opts?.onTick?.(last)
      if (last.status === 'succeeded' || last.status === 'failed') {
        await fetchAll()
        return last
      }
      await new Promise(resolve => setTimeout(resolve, intervalMs))
    }
    await fetchAll()
    return last
  }

  return {
    services,
    fetchAll,
    create,
    update,
    remove,
    control,
    getJob,
    pollJob
  }
}
