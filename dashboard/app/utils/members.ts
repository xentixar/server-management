export type MemberRole = 'owner' | 'admin' | 'viewer'
export type AssignableRole = Exclude<MemberRole, 'owner'>

export interface OrgMember {
  id: number
  name: string
  email: string
  role: MemberRole
}

export interface OrgInvite {
  id: number
  email: string
  role: AssignableRole
  created_at: string
}

export const MEMBER_ROLES: { value: MemberRole, label: string, description: string }[] = [{
  value: 'owner',
  label: 'Owner',
  description: 'Full control including billing and danger zone'
}, {
  value: 'admin',
  label: 'Admin',
  description: 'Manage members and organization settings'
}, {
  value: 'viewer',
  label: 'Viewer',
  description: 'Read-only access'
}]

export const ASSIGNABLE_ROLES = MEMBER_ROLES.filter(
  (r): r is { value: AssignableRole, label: string, description: string } => r.value !== 'owner'
)

export const ROLE_OPTIONS = MEMBER_ROLES.map(r => ({
  label: r.label,
  value: r.value
}))

export const ASSIGNABLE_ROLE_OPTIONS = ASSIGNABLE_ROLES.map(r => ({
  label: r.label,
  value: r.value
}))

export function roleLabel(role: MemberRole) {
  return MEMBER_ROLES.find(r => r.value === role)?.label ?? role
}
