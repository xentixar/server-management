import { mapJob, type AgentJob, type AgentJobApi } from '~/utils/runtimes'

export type SshKeyStatus = 'pending' | 'synced' | 'failed'

export interface ServerSshKey {
  id: string
  name: string
  publicKey: string
  fingerprint: string
  status: SshKeyStatus
  message: string
  createdAt: string
  updatedAt: string
}

export interface ServerSshKeyApi {
  id: string
  name: string
  public_key: string
  fingerprint: string
  status: SshKeyStatus
  message?: string
  created_at: string
  updated_at: string
}

export interface SshKeyJobResponse {
  job: AgentJob
  key?: ServerSshKey
}

export interface CreateSshKeyInput {
  name: string
  publicKey: string
}

export interface UpdateSshKeyInput {
  name?: string
  publicKey?: string
}

export function mapSshKey(data: ServerSshKeyApi): ServerSshKey {
  return {
    id: String(data.id ?? ''),
    name: data.name || '',
    publicKey: data.public_key || '',
    fingerprint: data.fingerprint || '',
    status: data.status || 'pending',
    message: data.message || '',
    createdAt: data.created_at,
    updatedAt: data.updated_at
  }
}

export function mapSshKeyJobResponse(data: {
  job: AgentJobApi
  key?: ServerSshKeyApi
}): SshKeyJobResponse {
  return {
    job: mapJob(data.job),
    key: data.key ? mapSshKey(data.key) : undefined
  }
}

export function sshKeyStatusColor(status: SshKeyStatus): 'success' | 'warning' | 'error' | 'neutral' | 'info' {
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

export function truncatePublicKey(key: string, max = 48): string {
  const v = key.trim()
  if (v.length <= max) {
    return v
  }
  return `${v.slice(0, max)}...`
}

export function looksLikePublicKey(value: string): boolean {
  const lines = value
    .split(/\r?\n/)
    .map(l => l.trim())
    .filter(l => l && !l.startsWith('#'))
  if (lines.length !== 1) {
    return false
  }
  const parts = lines[0]!.split(/\s+/)
  if (parts.length < 2) {
    return false
  }
  return /^(ssh-(ed25519|rsa)|ecdsa-sha2-nistp(256|384|521)|sk-ssh-ed25519@openssh\.com|sk-ecdsa-sha2-nistp256@openssh\.com)$/.test(parts[0]!)
}
