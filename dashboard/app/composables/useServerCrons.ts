import { ApiError } from '~/utils/api'
import { mapJob, type AgentJob, type AgentJobApi } from '~/utils/runtimes'
import {
  mapCron,
  mapCronJobResponse,
  type CreateCronInput,
  type CronJobResponse,
  type ServerCron,
  type ServerCronApi,
  type UpdateCronInput
} from '~/utils/processes'

export function useServerCrons(serverId: MaybeRefOrGetter<string | number>) {
  const { api } = useApi()
  const { selected } = useOrganizations()
  const crons = ref<ServerCron[]>([])

  function basePath() {
    const org = selected.value
    if (!org) {
      throw new ApiError('No organization selected', 400)
    }
    return `/api/v1/organizations/${org.id}/servers/${toValue(serverId)}`
  }

  async function fetchAll() {
    const res = await api<ServerCronApi[]>(`${basePath()}/crons`)
    crons.value = (res.data ?? []).map(mapCron)
    return crons.value
  }

  async function create(input: CreateCronInput): Promise<CronJobResponse> {
    const res = await api<{ job: AgentJobApi, cron: ServerCronApi }>(`${basePath()}/crons`, {
      method: 'POST',
      body: {
        schedule: input.schedule,
        command: input.command,
        user: input.user || undefined,
        timezone: input.timezone || 'UTC',
        enabled: input.enabled ?? true
      }
    })
    return mapCronJobResponse(res.data)
  }

  async function update(id: string, input: UpdateCronInput): Promise<CronJobResponse> {
    const body: Record<string, unknown> = {}
    if (input.schedule !== undefined) body.schedule = input.schedule
    if (input.command !== undefined) body.command = input.command
    if (input.user !== undefined) body.user = input.user
    if (input.timezone !== undefined) body.timezone = input.timezone
    if (input.enabled !== undefined) body.enabled = input.enabled

    const res = await api<{ job: AgentJobApi, cron: ServerCronApi }>(
      `${basePath()}/crons/${encodeURIComponent(id)}`,
      { method: 'PATCH', body }
    )
    return mapCronJobResponse(res.data)
  }

  async function remove(id: string): Promise<CronJobResponse> {
    const res = await api<{ job: AgentJobApi }>(
      `${basePath()}/crons/${encodeURIComponent(id)}`,
      { method: 'DELETE' }
    )
    return mapCronJobResponse(res.data)
  }

  async function fetchLogs(id: string): Promise<CronJobResponse> {
    const res = await api<{ job: AgentJobApi, cron: ServerCronApi }>(
      `${basePath()}/crons/${encodeURIComponent(id)}/logs`,
      { method: 'POST' }
    )
    return mapCronJobResponse(res.data)
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
    crons,
    fetchAll,
    create,
    update,
    remove,
    fetchLogs,
    getJob,
    pollJob
  }
}
