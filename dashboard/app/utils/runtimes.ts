export type ProvisionStatus = 'none' | 'pending' | 'running' | 'ready' | 'failed'
export type RuntimeName = 'php' | 'node'
export type RuntimeVersionStatus = 'installed' | 'installing' | 'failed'

export interface RuntimeVersion {
  version: string
  status: RuntimeVersionStatus
}

export interface RuntimeFamily {
  default: string
  versions: RuntimeVersion[]
}

export interface ServerRuntimes {
  php: RuntimeFamily
  node: RuntimeFamily
  addons: string[]
}

export interface RuntimeCatalog {
  php: string[]
  node: string[]
}

export interface ProvisionAddons {
  redis: boolean
  mysql: boolean
  mariadb: boolean
  postgres: boolean
}

export interface ProvisionInput {
  phpVersions: string[]
  nodeVersions: string[]
  addons: ProvisionAddons
}

export interface AgentJob {
  id: number
  kind: string
  payload: Record<string, unknown>
  status: 'pending' | 'running' | 'succeeded' | 'failed'
  message: string
  exitCode: number
  claimedAt?: string | null
  finishedAt?: string | null
  createdAt: string
}

export interface AgentJobApi {
  id: number
  kind: string
  payload: Record<string, unknown>
  status: AgentJob['status']
  message: string
  exit_code: number
  claimed_at?: string | null
  finished_at?: string | null
  created_at: string
}

export function emptyRuntimes(): ServerRuntimes {
  return {
    php: { default: '', versions: [] },
    node: { default: '', versions: [] },
    addons: []
  }
}

export function mapRuntimes(data?: Partial<ServerRuntimes> | null): ServerRuntimes {
  const base = emptyRuntimes()
  if (!data) {
    return base
  }
  return {
    php: {
      default: data.php?.default || '',
      versions: data.php?.versions ?? []
    },
    node: {
      default: data.node?.default || '',
      versions: data.node?.versions ?? []
    },
    addons: data.addons ?? []
  }
}

export function mapJob(data: AgentJobApi): AgentJob {
  return {
    id: data.id,
    kind: data.kind,
    payload: data.payload ?? {},
    status: data.status,
    message: data.message || '',
    exitCode: data.exit_code ?? 0,
    claimedAt: data.claimed_at,
    finishedAt: data.finished_at,
    createdAt: data.created_at
  }
}

function asStringList(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return []
  }
  return value.map(v => String(v)).filter(Boolean)
}

/** Restore provision form from a job payload (php_versions / node_versions / addons). */
export function selectionFromJobPayload(payload: Record<string, unknown> | null | undefined): ProvisionInput | null {
  if (!payload) {
    return null
  }
  const phpVersions = asStringList(payload.php_versions)
  const nodeVersions = asStringList(payload.node_versions)
  if (!phpVersions.length || !nodeVersions.length) {
    return null
  }
  const raw = (payload.addons && typeof payload.addons === 'object')
    ? payload.addons as Record<string, unknown>
    : {}
  return {
    phpVersions,
    nodeVersions,
    addons: {
      redis: !!raw.redis,
      mysql: !!raw.mysql,
      mariadb: !!raw.mariadb,
      postgres: !!raw.postgres
    }
  }
}

/** Restore provision form from what's actually installed on the server. */
export function selectionFromRuntimes(runtimes: ServerRuntimes | null | undefined): ProvisionInput | null {
  if (!runtimes) {
    return null
  }
  const phpVersions = runtimes.php.versions
    .filter(v => v.status === 'installed')
    .map(v => v.version)
  const nodeVersions = runtimes.node.versions
    .filter(v => v.status === 'installed')
    .map(v => v.version)
  if (!phpVersions.length || !nodeVersions.length) {
    return null
  }
  const set = new Set(runtimes.addons)
  return {
    phpVersions,
    nodeVersions,
    addons: {
      redis: set.has('redis'),
      mysql: set.has('mysql'),
      mariadb: set.has('mariadb'),
      postgres: set.has('postgres')
    }
  }
}

export function provisionBadgeColor(status: ProvisionStatus) {
  if (status === 'ready') {
    return 'success' as const
  }
  if (status === 'pending' || status === 'running') {
    return 'warning' as const
  }
  if (status === 'failed') {
    return 'error' as const
  }
  return 'neutral' as const
}

export function runtimeStatusColor(status: RuntimeVersionStatus) {
  if (status === 'installed') {
    return 'success' as const
  }
  if (status === 'failed') {
    return 'error' as const
  }
  return 'warning' as const
}

export function addonLabel(key: string) {
  const known = ADDON_OPTIONS.find(o => o.key === key)
  if (known) {
    return known.label
  }
  if (key === 'mysql') {
    return 'MySQL'
  }
  return key
}

export const DEFAULT_PHP_VERSIONS = ['8.2', '8.3', '8.4', '8.5']
export const DEFAULT_NODE_VERSIONS = ['20', '22', '24']

export const ADDON_OPTIONS = [
  { key: 'redis' as const, label: 'Redis' },
  { key: 'mariadb' as const, label: 'MariaDB' },
  { key: 'postgres' as const, label: 'PostgreSQL' }
]

export const RUNTIME_INSTALL_KIND: Record<RuntimeName, string> = {
  php: 'runtime.php.install',
  node: 'runtime.node.install'
}

export const RUNTIME_DEFAULT_KIND: Record<RuntimeName, string> = {
  php: 'runtime.php.default',
  node: 'runtime.node.default'
}
