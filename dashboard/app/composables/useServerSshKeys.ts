import { ApiError } from '~/utils/api'
import { mapJob, type AgentJob, type AgentJobApi } from '~/utils/runtimes'
import {
  mapSshKey,
  mapSshKeyJobResponse,
  type CreateSshKeyInput,
  type ServerSshKey,
  type ServerSshKeyApi,
  type SshKeyJobResponse,
  type UpdateSshKeyInput
} from '~/utils/ssh-keys'

export function useServerSshKeys(serverId: MaybeRefOrGetter<string | number>) {
  const { api } = useApi()
  const { selected } = useOrganizations()
  const keys = ref<ServerSshKey[]>([])

  function basePath() {
    const org = selected.value
    if (!org) {
      throw new ApiError('No organization selected', 400)
    }
    return `/api/v1/organizations/${org.id}/servers/${toValue(serverId)}`
  }

  async function fetchAll() {
    const res = await api<ServerSshKeyApi[]>(`${basePath()}/ssh-keys`)
    keys.value = (res.data ?? []).map(mapSshKey)
    return keys.value
  }

  async function create(input: CreateSshKeyInput): Promise<SshKeyJobResponse> {
    const res = await api<{ job: AgentJobApi, key: ServerSshKeyApi }>(
      `${basePath()}/ssh-keys`,
      {
        method: 'POST',
        body: {
          name: input.name.trim(),
          public_key: input.publicKey.trim()
        }
      }
    )
    return mapSshKeyJobResponse(res.data)
  }

  async function update(id: string, input: UpdateSshKeyInput): Promise<SshKeyJobResponse> {
    const body: Record<string, unknown> = {}
    if (input.name !== undefined) body.name = input.name.trim()
    if (input.publicKey !== undefined) body.public_key = input.publicKey.trim()

    const res = await api<{ job: AgentJobApi, key: ServerSshKeyApi }>(
      `${basePath()}/ssh-keys/${encodeURIComponent(id)}`,
      { method: 'PATCH', body }
    )
    return mapSshKeyJobResponse(res.data)
  }

  async function remove(id: string): Promise<SshKeyJobResponse> {
    const res = await api<{ job: AgentJobApi }>(
      `${basePath()}/ssh-keys/${encodeURIComponent(id)}`,
      { method: 'DELETE' }
    )
    return mapSshKeyJobResponse(res.data)
  }

  async function sync(): Promise<SshKeyJobResponse> {
    const res = await api<{ job: AgentJobApi }>(`${basePath()}/ssh-keys/sync`, {
      method: 'POST'
    })
    return mapSshKeyJobResponse(res.data)
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
    keys,
    fetchAll,
    create,
    update,
    remove,
    sync,
    getJob,
    pollJob
  }
}
