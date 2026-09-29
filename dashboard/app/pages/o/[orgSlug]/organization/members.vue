<script setup lang="ts">
import type { MemberRole, OrgMember } from '~/utils/members'
import { formatDate } from '~/utils/billing'

const {
  activeMembers,
  pendingInvites,
  invite,
  resendInvite,
  revokeInvite,
  remove
} = useMembers()

const toast = useToast()
const q = ref('')
const inviteOpen = ref(false)
const removeOpen = ref(false)
const memberToRemove = ref<OrgMember | null>(null)

const currentUserId = 'm_you'

const filteredActive = computed(() => {
  const query = q.value.trim().toLowerCase()
  if (!query) {
    return activeMembers.value
  }
  return activeMembers.value.filter((member) => {
    return member.name.toLowerCase().includes(query)
      || member.email.toLowerCase().includes(query)
  })
})

function onInvite(email: string, role: MemberRole) {
  const result = invite(email, role)
  if (!result.ok) {
    toast.add({ title: result.error, color: 'error' })
    return
  }
  inviteOpen.value = false
  toast.add({
    title: 'Invite sent',
    description: `${email} was invited as ${role}. Local stub until the API is wired.`,
    color: 'success'
  })
}

function onAskRemove(member: OrgMember) {
  memberToRemove.value = member
  removeOpen.value = true
}

function onConfirmRemove() {
  if (!memberToRemove.value) {
    return
  }
  const result = remove(memberToRemove.value.id)
  if (!result.ok) {
    toast.add({ title: result.error, color: 'error' })
    return
  }
  toast.add({
    title: 'Member removed',
    description: result.member.email,
    color: 'neutral'
  })
  memberToRemove.value = null
}

function onResend(member: OrgMember) {
  const result = resendInvite(member.id)
  if (!result.ok) {
    toast.add({ title: result.error, color: 'error' })
    return
  }
  toast.add({
    title: 'Invite resent',
    description: member.email,
    color: 'success'
  })
}

function onRevoke(member: OrgMember) {
  const result = revokeInvite(member.id)
  if (!result.ok) {
    toast.add({ title: result.error, color: 'error' })
    return
  }
  toast.add({
    title: 'Invite revoked',
    description: member.email,
    color: 'neutral'
  })
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

        <OrganizationMembersList
          v-if="filteredActive.length"
          :members="filteredActive"
          :current-user-id="currentUserId"
          @remove="onAskRemove"
          @resend="onResend"
          @revoke="onRevoke"
        />
        <p
          v-else
          class="text-muted px-4 py-6 text-sm sm:px-6"
        >
          No members match your search.
        </p>
      </UPageCard>
    </div>

    <div v-if="pendingInvites.length">
      <UPageCard
        title="Pending invites"
        :description="`${pendingInvites.length} waiting to accept.`"
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
            v-for="inviteRow in pendingInvites"
            :key="inviteRow.id"
            class="flex flex-wrap items-center justify-between gap-3 px-4 py-3 sm:px-6"
          >
            <div class="min-w-0 text-sm">
              <p class="text-highlighted font-medium">
                {{ inviteRow.email }}
              </p>
              <p class="text-muted">
                {{ inviteRow.role }} · invited
                {{ inviteRow.invitedAt ? formatDate(inviteRow.invitedAt) : '' }}
              </p>
            </div>
            <div class="flex items-center gap-2">
              <UButton
                label="Resend"
                color="neutral"
                variant="ghost"
                size="sm"
                @click="onResend(inviteRow)"
              />
              <UButton
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
      v-model:open="inviteOpen"
      @invite="onInvite"
    />

    <OrganizationRemoveMemberModal
      v-model:open="removeOpen"
      :member="memberToRemove"
      @confirm="onConfirmRemove"
    />
  </div>
</template>
