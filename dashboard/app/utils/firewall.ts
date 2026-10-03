import { mapJob, type AgentJob, type AgentJobApi } from '~/utils/runtimes'

export type FirewallStatus = 'pending' | 'synced' | 'failed'
export type FirewallAction = 'allow' | 'deny'
export type FirewallProtocol = 'tcp' | 'udp' | 'icmp' | 'any'

export interface ServerFirewallRule {
  id: string
  name: string
  action: FirewallAction
  direction: string
  protocol: FirewallProtocol
  portStart: number
  portEnd: number
  source: string
  position: number
  enabled: boolean
  status: FirewallStatus
  message: string
  createdAt: string
  updatedAt: string
}

export interface ServerFirewallRuleApi {
  id: string
  name?: string
  action: FirewallAction
  direction?: string
  protocol: FirewallProtocol
  port_start: number
  port_end: number
  source: string
  position: number
  enabled: boolean
  status: FirewallStatus
  message?: string
  created_at: string
  updated_at: string
}

export interface FirewallJobResponse {
  job: AgentJob
  rule?: ServerFirewallRule
}

export interface CreateFirewallRuleInput {
  name?: string
  action: FirewallAction
  direction?: string
  protocol: FirewallProtocol
  portStart?: number
  portEnd?: number
  source?: string
  position?: number
  enabled?: boolean
}

export interface UpdateFirewallRuleInput {
  name?: string
  action?: FirewallAction
  direction?: string
  protocol?: FirewallProtocol
  portStart?: number
  portEnd?: number
  source?: string
  position?: number
  enabled?: boolean
}

export function mapFirewallRule(data: ServerFirewallRuleApi): ServerFirewallRule {
  return {
    id: String(data.id ?? ''),
    name: data.name || '',
    action: data.action || 'allow',
    direction: data.direction || 'inbound',
    protocol: data.protocol || 'tcp',
    portStart: data.port_start ?? 0,
    portEnd: data.port_end ?? data.port_start ?? 0,
    source: data.source || '0.0.0.0/0',
    position: data.position ?? 0,
    enabled: data.enabled ?? true,
    status: data.status || 'pending',
    message: data.message || '',
    createdAt: data.created_at,
    updatedAt: data.updated_at
  }
}

export function mapFirewallJobResponse(data: {
  job: AgentJobApi
  rule?: ServerFirewallRuleApi
}): FirewallJobResponse {
  return {
    job: mapJob(data.job),
    rule: data.rule ? mapFirewallRule(data.rule) : undefined
  }
}

export function firewallStatusColor(status: FirewallStatus): 'success' | 'warning' | 'error' | 'neutral' | 'info' {
  switch (status) {
    case 'synced':
      return 'success'
    case 'pending':
      return 'info'
    case 'failed':
      return 'error'
    default:
      return 'neutral'
  }
}

export function formatFirewallPorts(rule: Pick<ServerFirewallRule, 'protocol' | 'portStart' | 'portEnd'>): string {
  if (rule.protocol === 'icmp' || rule.protocol === 'any') {
    return '-'
  }
  if (!rule.portStart && !rule.portEnd) {
    return 'any'
  }
  if (!rule.portEnd || rule.portEnd === rule.portStart) {
    return String(rule.portStart)
  }
  return `${rule.portStart}-${rule.portEnd}`
}

export function looksLikeCIDR(value: string): boolean {
  const v = value.trim()
  if (!v || v === 'any') {
    return true
  }
  if (/^\d{1,3}(\.\d{1,3}){3}$/.test(v)) {
    return true
  }
  return /^\d{1,3}(\.\d{1,3}){3}\/\d{1,2}$/.test(v)
}
