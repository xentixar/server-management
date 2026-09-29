export type MemberRole = 'owner' | 'admin' | 'member' | 'viewer'
export type MemberStatus = 'active' | 'invited' | 'suspended'

export interface OrgMember {
  id: string
  name: string
  email: string
  role: MemberRole
  status: MemberStatus
  avatar?: {
    src?: string
    alt?: string
  }
  invitedAt?: string
}

export const MEMBER_ROLES: { value: MemberRole, label: string, description: string }[] = [{
  value: 'owner',
  label: 'Owner',
  description: 'Full control including billing and danger zone'
}, {
  value: 'admin',
  label: 'Admin',
  description: 'Manage members, billing, and infrastructure'
}, {
  value: 'member',
  label: 'Member',
  description: 'Create servers, sites, and deploys'
}, {
  value: 'viewer',
  label: 'Viewer',
  description: 'Read-only access to logs and metrics'
}]

export const ROLE_OPTIONS = MEMBER_ROLES.map(r => ({
  label: r.label,
  value: r.value
}))

export function roleLabel(role: MemberRole) {
  return MEMBER_ROLES.find(r => r.value === role)?.label ?? role
}
