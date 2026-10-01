import { createSharedComposable } from '@vueuse/core'
import type { AssignableRole, OrgInvite, OrgMember } from '~/utils/members'
import { ApiError } from '~/utils/api'

function membersState() {
  return {
    members: useState<OrgMember[]>('org-members', () => []),
    invites: useState<OrgInvite[]>('org-invites', () => []),
    ready: useState('org-members-ready', () => false),
    pending: useState('org-members-pending', () => false),
    loadedOrgId: useState<number | null>('org-members-org', () => null)
  }
}

/** Safe to call from session clear — no useRoute. */
export function clearMembers() {
  const { members, invites, ready, loadedOrgId } = membersState()
  members.value = []
  invites.value = []
  ready.value = false
  loadedOrgId.value = null
}

const _useMembers = () => {
  const { api } = useApi()
  const { selected } = useOrganizations()
  const { user } = useSession()
  const { members, invites, ready, pending, loadedOrgId } = membersState()

  const currentMember = computed(() =>
    members.value.find(m => m.email === user.value?.email) ?? null
  )

  const ownerCount = computed(() =>
    members.value.filter(m => m.role === 'owner').length
  )

  function orgPath(suffix: string) {
    const org = selected.value
    if (!org) {
      throw new ApiError('No organization selected', 400)
    }
    return `/api/v1/organizations/${org.id}${suffix}`
  }

  async function fetchAll(force = false) {
    const org = selected.value
    if (!org) {
      clearMembers()
      return
    }

    if (!force && ready.value && loadedOrgId.value === org.id) {
      return
    }

    pending.value = true
    try {
      const [membersRes, invitesRes] = await Promise.all([
        api<OrgMember[]>(orgPath('/members')),
        api<OrgInvite[]>(orgPath('/invites'))
      ])
      members.value = membersRes.data
      invites.value = invitesRes.data
      loadedOrgId.value = org.id
      ready.value = true
    } finally {
      pending.value = false
    }
  }

  async function invite(email: string, role: AssignableRole) {
    await api(orgPath('/invites'), {
      method: 'POST',
      body: { email: email.trim().toLowerCase(), role }
    })
    await fetchAll(true)
  }

  async function setRole(id: number, role: AssignableRole) {
    const member = members.value.find(m => m.id === id)
    if (!member) {
      throw new ApiError('Member not found', 404)
    }
    if (member.role === 'owner' && ownerCount.value <= 1) {
      throw new ApiError('Keep at least one owner', 422)
    }

    const res = await api<OrgMember>(orgPath(`/members/${id}`), {
      method: 'PATCH',
      body: { role }
    })
    const idx = members.value.findIndex(m => m.id === id)
    if (idx !== -1) {
      members.value[idx] = res.data
    }
    return res.data
  }

  async function revokeInvite(id: number) {
    await api(orgPath(`/invites/${id}`), { method: 'DELETE' })
    invites.value = invites.value.filter(i => i.id !== id)
  }

  async function remove(id: number) {
    const member = members.value.find(m => m.id === id)
    if (!member) {
      throw new ApiError('Member not found', 404)
    }
    if (member.role === 'owner' && ownerCount.value <= 1) {
      throw new ApiError('Cannot remove the only owner', 422)
    }
    if (member.email === user.value?.email) {
      throw new ApiError('You cannot remove yourself here', 422)
    }

    await api(orgPath(`/members/${id}`), { method: 'DELETE' })
    members.value = members.value.filter(m => m.id !== id)
    return member
  }

  async function acceptInvite(organizationId: number) {
    await api(`/api/v1/organizations/${organizationId}/invites/accept`, {
      method: 'POST'
    })
  }

  watch(() => selected.value?.id, (id, prev) => {
    if (id === prev) {
      return
    }
    if (!id) {
      clearMembers()
      return
    }
    if (id !== loadedOrgId.value) {
      ready.value = false
      fetchAll()
    }
  })

  return {
    members,
    invites,
    pendingInvites: invites,
    ready,
    pending,
    currentMember,
    ownerCount,
    fetchAll,
    invite,
    setRole,
    revokeInvite,
    remove,
    acceptInvite,
    clearMembers
  }
}

export const useMembers = createSharedComposable(_useMembers)
