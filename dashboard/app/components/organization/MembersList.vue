<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import type { AssignableRole, MemberRole, OrgMember } from '~/utils/members'
import { ASSIGNABLE_ROLE_OPTIONS, ROLE_OPTIONS, roleLabel } from '~/utils/members'
import { ApiError } from '~/utils/api'
import { Ability } from '~/utils/auth/abilities'

const props = defineProps<{
  members: OrgMember[]
  currentUserId?: number | null
}>()

const emit = defineEmits<{
  remove: [member: OrgMember]
}>()

const { setRole } = useMembers()
const { can } = useGate()
const toast = useToast()

function roleItems(member: OrgMember) {
  if (member.role === 'owner') {
    return ROLE_OPTIONS
  }
  return ASSIGNABLE_ROLE_OPTIONS
}

function canEditRole(member: OrgMember) {
  return can(Ability.MembersUpdate)
    && member.role !== 'owner'
    && member.id !== props.currentUserId
}

async function onRoleChange(member: OrgMember, role: MemberRole) {
  if (role === member.role || role === 'owner' || !canEditRole(member)) {
    return
  }
  try {
    await setRole(member.id, role as AssignableRole)
    toast.add({
      title: 'Role updated',
      description: `${member.name} is now ${roleLabel(role)}.`,
      color: 'success'
    })
  } catch (error) {
    const err = error instanceof ApiError ? error : new ApiError('Update failed', 500)
    toast.add({ title: err.message, color: 'error' })
  }
}

function menuItems(member: OrgMember): DropdownMenuItem[][] {
  if (!can(Ability.MembersRemove) || member.id === props.currentUserId || member.role === 'owner') {
    return []
  }
  return [[{
    label: 'Remove member',
    icon: 'i-lucide-user-minus',
    color: 'error',
    onSelect: () => emit('remove', member)
  }]]
}
</script>

<template>
  <ul
    role="list"
    class="divide-y divide-default"
  >
    <li
      v-for="member in members"
      :key="member.id"
      class="flex items-center justify-between gap-3 px-4 py-3 sm:px-6"
    >
      <div class="flex min-w-0 items-center gap-3">
        <UAvatar
          :alt="member.name"
          size="md"
        />

        <div class="min-w-0 text-sm">
          <p class="text-highlighted truncate font-medium">
            {{ member.name }}
            <span
              v-if="member.id === currentUserId"
              class="text-muted font-normal"
            >(you)</span>
          </p>
          <p class="text-muted truncate">
            {{ member.email }}
          </p>
        </div>
      </div>

      <div class="flex shrink-0 items-center gap-2 sm:gap-3">
        <USelect
          v-if="can(Ability.MembersUpdate)"
          :model-value="member.role"
          :items="roleItems(member)"
          value-key="value"
          color="neutral"
          class="w-32"
          :disabled="!canEditRole(member)"
          @update:model-value="(value) => onRoleChange(member, value as MemberRole)"
        />
        <UBadge
          v-else
          :label="roleLabel(member.role)"
          color="neutral"
          variant="outline"
        />

        <UDropdownMenu
          v-if="menuItems(member).length"
          :items="menuItems(member)"
          :content="{ align: 'end' }"
        >
          <UButton
            icon="i-lucide-ellipsis-vertical"
            color="neutral"
            variant="ghost"
            aria-label="Member actions"
          />
        </UDropdownMenu>
      </div>
    </li>
  </ul>
</template>
