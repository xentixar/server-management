import { mapJob, type AgentJob, type AgentJobApi } from '~/utils/runtimes'

export type ServiceStatus = 'pending' | 'running' | 'stopped' | 'failed' | 'unknown'
export type CronStatus = 'pending' | 'synced' | 'failed'
export type ServiceControlAction = 'start' | 'stop' | 'restart' | 'status'

export interface ServerService {
  id: string
  name: string
  command: string
  directory: string
  user: string
  processes: number
  autoRestart: boolean
  autostart: boolean
  enabled: boolean
  status: ServiceStatus
  stdoutTail: string
  stderrTail: string
  message: string
  createdAt: string
  updatedAt: string
}

export interface ServerServiceApi {
  id: string
  name: string
  command: string
  directory?: string
  user: string
  processes: number
  auto_restart: boolean
  autostart: boolean
  enabled: boolean
  status: ServiceStatus
  stdout_tail?: string
  stderr_tail?: string
  message?: string
  created_at: string
  updated_at: string
}

export interface ServerCron {
  id: string
  schedule: string
  command: string
  user: string
  timezone: string
  enabled: boolean
  status: CronStatus
  message: string
  logTail: string
  createdAt: string
  updatedAt: string
}

export interface ServerCronApi {
  id: string
  schedule: string
  command: string
  user: string
  timezone: string
  enabled: boolean
  status: CronStatus
  message?: string
  log_tail?: string
  created_at: string
  updated_at: string
}

export interface ServiceJobResponse {
  job: AgentJob
  service: ServerService
}

export interface CronJobResponse {
  job: AgentJob
  cron?: ServerCron
}

export interface CreateServiceInput {
  name: string
  command: string
  directory?: string
  user?: string
  processes?: number
  autoRestart?: boolean
  autostart?: boolean
  enabled?: boolean
}

export interface UpdateServiceInput {
  command?: string
  directory?: string
  user?: string
  processes?: number
  autoRestart?: boolean
  autostart?: boolean
  enabled?: boolean
}

export interface CreateCronInput {
  schedule: string
  command: string
  user?: string
  timezone?: string
  enabled?: boolean
}

export interface UpdateCronInput {
  schedule?: string
  command?: string
  user?: string
  timezone?: string
  enabled?: boolean
}

export function mapService(data: ServerServiceApi): ServerService {
  return {
    id: String(data.id ?? ''),
    name: data.name,
    command: data.command,
    directory: data.directory || '',
    user: data.user || 'sm',
    processes: data.processes ?? 1,
    autoRestart: data.auto_restart ?? true,
    autostart: data.autostart ?? true,
    enabled: data.enabled ?? true,
    status: data.status || 'unknown',
    stdoutTail: data.stdout_tail || '',
    stderrTail: data.stderr_tail || '',
    message: data.message || '',
    createdAt: data.created_at,
    updatedAt: data.updated_at
  }
}

export function mapCron(data: ServerCronApi): ServerCron {
  return {
    id: String(data.id ?? ''),
    schedule: data.schedule,
    command: data.command,
    user: data.user || 'sm',
    timezone: data.timezone || 'UTC',
    enabled: data.enabled ?? true,
    status: data.status || 'pending',
    message: data.message || '',
    logTail: data.log_tail || '',
    createdAt: data.created_at,
    updatedAt: data.updated_at
  }
}

export function mapServiceJobResponse(data: { job: AgentJobApi, service: ServerServiceApi }): ServiceJobResponse {
  return {
    job: mapJob(data.job),
    service: mapService(data.service)
  }
}

export function mapCronJobResponse(data: { job: AgentJobApi, cron?: ServerCronApi }): CronJobResponse {
  return {
    job: mapJob(data.job),
    cron: data.cron ? mapCron(data.cron) : undefined
  }
}

export function serviceStatusColor(status: ServiceStatus): 'success' | 'warning' | 'error' | 'neutral' | 'info' {
  switch (status) {
    case 'running':
      return 'success'
    case 'pending':
      return 'info'
    case 'failed':
      return 'error'
    case 'stopped':
      return 'neutral'
    default:
      return 'warning'
  }
}

export function cronStatusColor(status: CronStatus): 'success' | 'warning' | 'error' | 'neutral' | 'info' {
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

const cronFieldRe = /^(\*|([0-9]|[1-5][0-9])|\*\/([0-9]|[1-5][0-9])|([0-9]|[1-5][0-9])-([0-9]|[1-5][0-9]))(,(\*|([0-9]|[1-5][0-9])|\*\/([0-9]|[1-5][0-9])|([0-9]|[1-5][0-9])-([0-9]|[1-5][0-9])))*$/

export function looksLikeCronSchedule(expr: string): boolean {
  const fields = expr.trim().split(/\s+/)
  return fields.length === 5 && fields.every(f => cronFieldRe.test(f) || f.includes('/'))
}

export function isValidServiceName(name: string): boolean {
  return /^[a-z0-9][a-z0-9_-]{0,62}$/.test(name)
}
