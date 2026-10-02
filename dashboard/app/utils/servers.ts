import { mapRuntimes, type ProvisionStatus, type ServerRuntimes } from '~/utils/runtimes'

export type AgentStatus = 'pending' | 'online' | 'offline' | 'degraded'
export type SshAuthMethod = 'key' | 'password'

export interface Server {
  id: number
  hostname: string
  ip: string
  sshPort: number
  sshUser: string
  authMethod: SshAuthMethod
  hasKeyUploaded: boolean
  region?: string
  labels: string[]
  notes?: string
  agent: AgentStatus
  agentVersion?: string
  lastSeenAt?: string | null
  provisionStatus: ProvisionStatus
  provisionedAt?: string | null
  runtimes: ServerRuntimes
  sites: number
  createdAt: string
  agentToken?: string
  bootstrap?: {
    ok: boolean
    message: string
  }
}

export interface CreateServerInput {
  hostname: string
  ip: string
  sshPort: number
  sshUser: string
  authMethod: SshAuthMethod
  privateKey?: string
  passphrase?: string
  password?: string
  region?: string
  labels: string[]
  notes?: string
}

export interface ServerApi {
  id: number
  hostname: string
  ip: string
  ssh_port: number
  ssh_user: string
  auth_method: SshAuthMethod
  has_key: boolean
  region?: string
  labels: string[]
  notes?: string
  agent: AgentStatus
  agent_version?: string
  last_seen_at?: string | null
  provision_status?: ProvisionStatus
  provisioned_at?: string | null
  runtimes?: ServerRuntimes
  sites: number
  created_at: string
  agent_token?: string
  bootstrap?: {
    ok: boolean
    message: string
  }
}

export function mapServer(data: ServerApi): Server {
  return {
    id: data.id,
    hostname: data.hostname,
    ip: data.ip,
    sshPort: data.ssh_port,
    sshUser: data.ssh_user,
    authMethod: data.auth_method,
    hasKeyUploaded: data.has_key,
    region: data.region || undefined,
    labels: data.labels ?? [],
    notes: data.notes || undefined,
    agent: data.agent,
    agentVersion: data.agent_version || undefined,
    lastSeenAt: data.last_seen_at,
    provisionStatus: data.provision_status || 'none',
    provisionedAt: data.provisioned_at,
    runtimes: mapRuntimes(data.runtimes),
    sites: data.sites ?? 0,
    createdAt: data.created_at,
    agentToken: data.agent_token,
    bootstrap: data.bootstrap
  }
}

export function agentBadgeColor(agent: AgentStatus) {
  if (agent === 'online') {
    return 'success' as const
  }
  if (agent === 'degraded' || agent === 'pending') {
    return 'warning' as const
  }
  return 'error' as const
}
