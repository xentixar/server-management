import type { MemberRole, MemberStatus, OrgMember } from '~/utils/members'

function seedMembers(): OrgMember[] {
  return [{
    id: 'm_you',
    name: 'Outpost User',
    email: 'you@outpost.dev',
    role: 'owner',
    status: 'active',
    avatar: { alt: 'Outpost User' }
  }, {
    id: 'm_alex',
    name: 'Alex Smith',
    email: 'alex@example.com',
    role: 'admin',
    status: 'active',
    avatar: { alt: 'Alex Smith' }
  }, {
    id: 'm_sam',
    name: 'Sam Rivera',
    email: 'sam@example.com',
    role: 'member',
    status: 'active',
    avatar: { alt: 'Sam Rivera' }
  }, {
    id: 'm_invite_1',
    name: 'jamie@example.com',
    email: 'jamie@example.com',
    role: 'viewer',
    status: 'invited',
    avatar: { alt: 'jamie@example.com' },
    invitedAt: new Date(Date.now() - 1000 * 60 * 60 * 26).toISOString()
  }]
}

// ponytail: local stub until members API exists; keyed by org id
const byOrg = reactive<Record<string, OrgMember[]>>({})

function ensure(orgId: string | number): OrgMember[] {
  const key = String(orgId)
  if (!byOrg[key]) {
    byOrg[key] = seedMembers()
  }
  return byOrg[key]!
}

export function useMembers() {
  const { selected } = useOrganizations()

  const members = computed(() => ensure(selected.value?.id ?? '_'))

  const activeMembers = computed(() =>
    members.value.filter(m => m.status === 'active' || m.status === 'suspended')
  )

  const pendingInvites = computed(() =>
    members.value.filter(m => m.status === 'invited')
  )

  const ownerCount = computed(() =>
    members.value.filter(m => m.role === 'owner' && m.status !== 'invited').length
  )

  function invite(email: string, role: MemberRole) {
    const list = ensure(selected.value?.id ?? '_')
    const normalized = email.trim().toLowerCase()
    if (list.some(m => m.email.toLowerCase() === normalized)) {
      return { ok: false as const, error: 'That email is already a member or invited.' }
    }
    if (role === 'owner') {
      return { ok: false as const, error: 'Invite as admin or below, then transfer ownership.' }
    }

    const member: OrgMember = {
      id: `m_${Date.now()}`,
      name: normalized,
      email: normalized,
      role,
      status: 'invited',
      avatar: { alt: normalized },
      invitedAt: new Date().toISOString()
    }
    list.unshift(member)
    return { ok: true as const, member }
  }

  function setRole(id: string, role: MemberRole) {
    const list = ensure(selected.value?.id ?? '_')
    const member = list.find(m => m.id === id)
    if (!member) {
      return { ok: false as const, error: 'Member not found.' }
    }

    if (member.role === 'owner' && role !== 'owner' && ownerCount.value <= 1) {
      return { ok: false as const, error: 'Keep at least one owner.' }
    }

    member.role = role
    return { ok: true as const, member }
  }

  function setStatus(id: string, status: MemberStatus) {
    const list = ensure(selected.value?.id ?? '_')
    const member = list.find(m => m.id === id)
    if (!member) {
      return { ok: false as const, error: 'Member not found.' }
    }
    if (member.role === 'owner' && status === 'suspended' && ownerCount.value <= 1) {
      return { ok: false as const, error: 'Cannot suspend the only owner.' }
    }
    member.status = status
    return { ok: true as const, member }
  }

  function resendInvite(id: string) {
    const list = ensure(selected.value?.id ?? '_')
    const member = list.find(m => m.id === id && m.status === 'invited')
    if (!member) {
      return { ok: false as const, error: 'Invite not found.' }
    }
    member.invitedAt = new Date().toISOString()
    return { ok: true as const, member }
  }

  function revokeInvite(id: string) {
    const list = ensure(selected.value?.id ?? '_')
    const index = list.findIndex(m => m.id === id && m.status === 'invited')
    if (index === -1) {
      return { ok: false as const, error: 'Invite not found.' }
    }
    const [removed] = list.splice(index, 1)
    return { ok: true as const, member: removed! }
  }

  function remove(id: string) {
    const list = ensure(selected.value?.id ?? '_')
    const member = list.find(m => m.id === id)
    if (!member) {
      return { ok: false as const, error: 'Member not found.' }
    }
    if (member.role === 'owner' && ownerCount.value <= 1) {
      return { ok: false as const, error: 'Cannot remove the only owner.' }
    }
    if (member.id === 'm_you') {
      return { ok: false as const, error: 'You cannot remove yourself here.' }
    }

    const index = list.findIndex(m => m.id === id)
    list.splice(index, 1)
    return { ok: true as const, member }
  }

  return {
    members,
    activeMembers,
    pendingInvites,
    ownerCount,
    invite,
    setRole,
    setStatus,
    resendInvite,
    revokeInvite,
    remove
  }
}
