<script setup lang="ts">
import type { AssignableRole, OrgInvite, OrgMember } from '~/utils/members'
import { formatDate } from '~/utils/billing'
import { ApiError } from '~/utils/api'
import { Ability } from '~/utils/auth/abilities'

const {
  members,
  pendingInvites,
  pending,
  currentMember,
  fetchAll,
  invite,
  revokeInvite,
  remove
} = useMembers()
const { selected } = useOrganizations()
const { can } = useGate()

await fetchAll()

const toast = useToast()
const q = ref('')
const inviteOpen = ref(false)
const removeOpen = ref(false)
const memberToRemove = ref<OrgMember | null>(null)

const visibleInvites = computed(() =>
  can(Ability.InvitesView) ? pendingInvites.value : []
)

function inviteLink(email: string) {
  const org = selected.value
  if (!org) {
    return ''
  }
  const url = new URL(`/invites/${org.id}`, window.location.origin)
  url.searchParams.set('email', email)
  return url.toString()
}

const filteredActive = computed(() => {
  const query = q.value.trim().toLowerCase()
  if (!query) {
    return members.value
  }
  return members.value.filter((member) => {
    return member.name.toLowerCase().includes(query)
      || member.email.toLowerCase().includes(query)
  })
})

async function onInvite(email: string, role: AssignableRole) {
  try {
    await invite(email, role)
    inviteOpen.value = false
    const link = inviteLink(email)
    if (import.meta.client && link && navigator.clipboard) {
      await navigator.clipboard.writeText(link).catch(() => null)
    }
    toast.add({
      title: 'Invite sent',
      description: link
        ? `Link copied for ${email}. Share it so they can accept.`
        : `${email} was invited as ${role}.`,
      color: 'success'
    })
  } catch (error) {
    const err = error instanceof ApiError ? error : new ApiError('Invite failed', 500)
    toast.add({ title: err.message, color: 'error' })
  }
}

function onAskRemove(member: OrgMember) {
  memberToRemove.value = member
  removeOpen.value = true
}

async function onConfirmRemove() {
  if (!memberToRemove.value) {
    return
  }
  try {
    const removed = await remove(memberToRemove.value.id)
    toast.add({
      title: 'Member removed',
      description: removed.email,
      color: 'neutral'
    })
    memberToRemove.value = null
  } catch (error) {
    const err = error instanceof ApiError ? error : new ApiError('Remove failed', 500)
    toast.add({ title: err.message, color: 'error' })
  }
}

async function onRevoke(inviteRow: OrgInvite) {
  try {
    await revokeInvite(inviteRow.id)
    toast.add({
      title: 'Invite revoked',
      description: inviteRow.email,
      color: 'neutral'
    })
  } catch (error) {
    const err = error instanceof ApiError ? error : new ApiError('Revoke failed', 500)
    toast.add({ title: err.message, color: 'error' })
  }
}

async function onCopyLink(email: string) {
  const link = inviteLink(email)
  if (!link) {
    return
  }
  try {
    await navigator.clipboard.writeText(link)
    toast.add({ title: 'Invite link copied', color: 'success' })
  } catch {
    toast.add({ title: 'Could not copy link', description: link, color: 'error' })
  }
}
</script>

<template>
  <div class="flex flex-col gap-4 sm:gap-6">
    <div>
      <UPageCard
        title="Members"
        description="People with access to this organization."
        variant="naked"
        orientation="horizontal"
        class="mb-4"
      >
        <UButton
          v-if="can(Ability.MembersInvite)"
          label="Invite people"
          color="neutral"
          icon="i-lucide-user-plus"
          class="w-fit lg:ms-auto"
          @click="inviteOpen = true"
        />
      </UPageCard>

      <UPageCard
        variant="subtle"
        :ui="{ container: 'p-0 sm:p-0 gap-y-0', wrapper: 'items-stretch', header: 'p-4 mb-0 border-b border-default' }"
      >
        <template #header>
          <UInput
            v-model="q"
            icon="i-lucide-search"
            placeholder="Search members"
            class="w-full"
          />
        </template>

        <div
          v-if="pending && !members.length"
          class="text-muted px-4 py-6 text-sm sm:px-6"
        >
          Loading members…
        </div>
        <OrganizationMembersList
          v-else-if="filteredActive.length"
          :members="filteredActive"
          :current-user-id="currentMember?.id"
          @remove="onAskRemove"
        />
        <p
          v-else
          class="text-muted px-4 py-6 text-sm sm:px-6"
        >
          No members match your search.
        </p>
      </UPageCard>
    </div>

    <div v-if="visibleInvites.length">
      <UPageCard
        title="Pending invites"
        :description="`${visibleInvites.length} waiting to accept.`"
        variant="naked"
        class="mb-4"
      />

      <UPageCard
        variant="subtle"
        :ui="{ container: 'p-0 sm:p-0' }"
      >
        <ul
          role="list"
          class="divide-y divide-default"
        >
          <li
            v-for="inviteRow in visibleInvites"
            :key="inviteRow.id"
            class="flex flex-wrap items-center justify-between gap-3 px-4 py-3 sm:px-6"
          >
            <div class="min-w-0 text-sm">
              <p class="text-highlighted font-medium">
                {{ inviteRow.email }}
              </p>
              <p class="text-muted">
                {{ inviteRow.role }} · invited
                {{ formatDate(inviteRow.created_at) }}
              </p>
            </div>
            <div class="flex items-center gap-2">
              <UButton
                v-if="can(Ability.MembersInvite)"
                label="Copy link"
                color="neutral"
                variant="ghost"
                size="sm"
                @click="onCopyLink(inviteRow.email)"
              />
              <UButton
                v-if="can(Ability.InvitesRevoke)"
                label="Revoke"
                color="error"
                variant="ghost"
                size="sm"
                @click="onRevoke(inviteRow)"
              />
            </div>
          </li>
        </ul>
      </UPageCard>
    </div>

    <OrganizationInviteMemberModal
      v-if="can(Ability.MembersInvite)"
      v-model:open="inviteOpen"
      @invite="onInvite"
    />

    <OrganizationRemoveMemberModal
      v-if="can(Ability.MembersRemove)"
      v-model:open="removeOpen"
      :member="memberToRemove"
      @confirm="onConfirmRemove"
    />
  </div>
</template>
