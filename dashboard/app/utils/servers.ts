export type AgentStatus = 'pending' | 'online' | 'offline' | 'degraded'
export type SshAuthMethod = 'key' | 'password'

export interface Server {
  id: string
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
  sites: number
  createdAt: string
}

export interface CreateServerInput {
  hostname: string
  ip: string
  sshPort: number
  sshUser: string
  authMethod: SshAuthMethod
  hasKeyUploaded: boolean
  region?: string
  labels: string[]
  notes?: string
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
