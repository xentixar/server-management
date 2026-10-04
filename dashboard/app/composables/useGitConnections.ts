import { ApiError } from '~/utils/api'
import {
  mapGitConnection,
  mapGitRepo,
  type GitConnection,
  type GitConnectionApi,
  type GitProvider,
  type GitRepo,
  type GitRepoApi
} from '~/utils/gitConnections'

function gitState() {
  return {
    connections: useState<GitConnection[]>('org-git-connections', () => []),
    ready: useState('org-git-connections-ready', () => false),
    pending: useState('org-git-connections-pending', () => false),
    loadedOrgId: useState<number | null>('org-git-connections-org', () => null)
  }
}

export function clearGitConnections() {
  const { connections, ready, loadedOrgId } = gitState()
  connections.value = []
  ready.value = false
  loadedOrgId.value = null
}

export function useGitConnections() {
  const { api } = useApi()
  const { selected } = useOrganizations()
  const { connections, ready, pending, loadedOrgId } = gitState()

  function orgPath(suffix = '') {
    const org = selected.value
    if (!org) {
      throw new ApiError('No organization selected', 400)
    }
    return `/api/v1/organizations/${org.id}/git-connections${suffix}`
  }

  async function fetchAll(force = false) {
    const org = selected.value
    if (!org) {
      clearGitConnections()
      return
    }

    if (!force && ready.value && loadedOrgId.value === org.id) {
      return
    }

    pending.value = true
    try {
      const res = await api<GitConnectionApi[]>(orgPath())
      connections.value = res.data.map(mapGitConnection)
      loadedOrgId.value = org.id
      ready.value = true
    } finally {
      pending.value = false
    }
  }

  async function authorize(provider: GitProvider): Promise<string> {
    const res = await api<{ authorize_url: string }>(orgPath(`/${provider}/authorize`), {
      method: 'POST'
    })
    const url = res.data.authorize_url
    if (!url) {
      throw new ApiError('Authorize URL missing', 500)
    }
    return url
  }

  async function disconnect(connectionId: string) {
    await api(orgPath(`/${connectionId}`), { method: 'DELETE' })
    connections.value = connections.value.filter(c => c.id !== connectionId)
  }

  async function listRepos(connectionId: string, query = '', page = 1): Promise<GitRepo[]> {
    const params = new URLSearchParams()
    if (query.trim()) {
      params.set('q', query.trim())
    }
    if (page > 1) {
      params.set('page', String(page))
    }
    const qs = params.toString()
    const res = await api<{
      repos: GitRepoApi[]
      page: number
      connection: GitConnectionApi
    }>(orgPath(`/${connectionId}/repos${qs ? `?${qs}` : ''}`))

    const mapped = mapGitConnection(res.data.connection)
    const idx = connections.value.findIndex(c => c.id === mapped.id)
    if (idx >= 0) {
      connections.value[idx] = mapped
    }

    return res.data.repos.map(mapGitRepo)
  }

  return {
    connections,
    ready,
    pending,
    fetchAll,
    authorize,
    disconnect,
    listRepos
  }
}
