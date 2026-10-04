export type GitProvider = 'github' | 'gitlab' | 'bitbucket'
export type GitConnectionStatus = 'active' | 'error'

export interface GitConnectionApi {
  id: string
  provider: GitProvider
  external_account_id: string
  account_login: string
  account_name: string
  avatar_url: string
  scopes: string
  status: GitConnectionStatus
  last_error: string
  token_expires_at: string | null
  created_at: string
  updated_at: string
}

export interface GitConnection {
  id: string
  provider: GitProvider
  externalAccountId: string
  accountLogin: string
  accountName: string
  avatarUrl: string
  scopes: string
  status: GitConnectionStatus
  lastError: string
  tokenExpiresAt: string | null
  createdAt: string
  updatedAt: string
}

export interface GitRepoApi {
  id: string
  name: string
  full_name: string
  private: boolean
  default_branch: string
  clone_url: string
  html_url: string
  provider: GitProvider
}

export interface GitRepo {
  id: string
  name: string
  fullName: string
  private: boolean
  defaultBranch: string
  cloneUrl: string
  htmlUrl: string
  provider: GitProvider
}

export const GIT_PROVIDERS: {
  value: GitProvider
  label: string
  icon: string
  description: string
}[] = [{
  value: 'github',
  label: 'GitHub',
  icon: 'i-simple-icons-github',
  description: 'Connect a GitHub account to list repositories.'
}, {
  value: 'gitlab',
  label: 'GitLab',
  icon: 'i-simple-icons-gitlab',
  description: 'Connect a GitLab.com account to list repositories.'
}, {
  value: 'bitbucket',
  label: 'Bitbucket',
  icon: 'i-simple-icons-bitbucket',
  description: 'Connect a Bitbucket Cloud account to list repositories.'
}]

export function mapGitConnection(row: GitConnectionApi): GitConnection {
  return {
    id: row.id,
    provider: row.provider,
    externalAccountId: row.external_account_id,
    accountLogin: row.account_login,
    accountName: row.account_name,
    avatarUrl: row.avatar_url,
    scopes: row.scopes,
    status: row.status,
    lastError: row.last_error,
    tokenExpiresAt: row.token_expires_at,
    createdAt: row.created_at,
    updatedAt: row.updated_at
  }
}

export function mapGitRepo(row: GitRepoApi): GitRepo {
  return {
    id: row.id,
    name: row.name,
    fullName: row.full_name,
    private: row.private,
    defaultBranch: row.default_branch,
    cloneUrl: row.clone_url,
    htmlUrl: row.html_url,
    provider: row.provider
  }
}

export function providerLabel(provider: GitProvider) {
  return GIT_PROVIDERS.find(p => p.value === provider)?.label ?? provider
}

export function connectionStatusColor(status: GitConnectionStatus) {
  if (status === 'active') {
    return 'success' as const
  }
  return 'error' as const
}
