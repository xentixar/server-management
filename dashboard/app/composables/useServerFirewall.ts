import { ApiError } from '~/utils/api'
import { mapJob, type AgentJob, type AgentJobApi } from '~/utils/runtimes'
import {
  mapFirewallJobResponse,
  mapFirewallRule,
  type CreateFirewallRuleInput,
  type FirewallJobResponse,
  type ServerFirewallRule,
  type ServerFirewallRuleApi,
  type UpdateFirewallRuleInput
} from '~/utils/firewall'

export function useServerFirewall(serverId: MaybeRefOrGetter<string | number>) {
  const { api } = useApi()
  const { selected } = useOrganizations()
  const rules = ref<ServerFirewallRule[]>([])

  function basePath() {
    const org = selected.value
    if (!org) {
      throw new ApiError('No organization selected', 400)
    }
    return `/api/v1/organizations/${org.id}/servers/${toValue(serverId)}`
  }

  async function fetchAll() {
    const res = await api<ServerFirewallRuleApi[]>(`${basePath()}/firewall/rules`)
    rules.value = (res.data ?? []).map(mapFirewallRule)
    return rules.value
  }

  async function create(input: CreateFirewallRuleInput): Promise<FirewallJobResponse> {
    const res = await api<{ job: AgentJobApi, rule: ServerFirewallRuleApi }>(
      `${basePath()}/firewall/rules`,
      {
        method: 'POST',
        body: {
          name: input.name || undefined,
          action: input.action,
          direction: input.direction || 'inbound',
          protocol: input.protocol,
          port_start: input.portStart,
          port_end: input.portEnd,
          source: input.source || undefined,
          position: input.position,
          enabled: input.enabled ?? true
        }
      }
    )
    return mapFirewallJobResponse(res.data)
  }

  async function update(id: string, input: UpdateFirewallRuleInput): Promise<FirewallJobResponse> {
    const body: Record<string, unknown> = {}
    if (input.name !== undefined) body.name = input.name
    if (input.action !== undefined) body.action = input.action
    if (input.direction !== undefined) body.direction = input.direction
    if (input.protocol !== undefined) body.protocol = input.protocol
    if (input.portStart !== undefined) body.port_start = input.portStart
    if (input.portEnd !== undefined) body.port_end = input.portEnd
    if (input.source !== undefined) body.source = input.source
    if (input.position !== undefined) body.position = input.position
    if (input.enabled !== undefined) body.enabled = input.enabled

    const res = await api<{ job: AgentJobApi, rule: ServerFirewallRuleApi }>(
      `${basePath()}/firewall/rules/${encodeURIComponent(id)}`,
      { method: 'PATCH', body }
    )
    return mapFirewallJobResponse(res.data)
  }

  async function remove(id: string): Promise<FirewallJobResponse> {
    const res = await api<{ job: AgentJobApi }>(
      `${basePath()}/firewall/rules/${encodeURIComponent(id)}`,
      { method: 'DELETE' }
    )
    return mapFirewallJobResponse(res.data)
  }

  async function sync(): Promise<FirewallJobResponse> {
    const res = await api<{ job: AgentJobApi }>(`${basePath()}/firewall/sync`, {
      method: 'POST'
    })
    return mapFirewallJobResponse(res.data)
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
    rules,
    fetchAll,
    create,
    update,
    remove,
    sync,
    getJob,
    pollJob
  }
}
