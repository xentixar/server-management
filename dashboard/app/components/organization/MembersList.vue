<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import type { MemberRole, OrgMember } from '~/utils/members'
import { ROLE_OPTIONS, roleLabel } from '~/utils/members'

const props = defineProps<{
  members: OrgMember[]
  currentUserId?: string
}>()

const emit = defineEmits<{
  remove: [member: OrgMember]
  resend: [member: OrgMember]
  revoke: [member: OrgMember]
}>()

const { setRole, setStatus } = useMembers()
const toast = useToast()

function onRoleChange(member: OrgMember, role: MemberRole) {
  if (role === member.role) {
    return
  }
  const result = setRole(member.id, role)
  if (!result.ok) {
    toast.add({ title: result.error, color: 'error' })
    return
  }
  toast.add({
    title: 'Role updated',
    description: `${member.name} is now ${roleLabel(role)}.`,
    color: 'success'
  })
}

function menuItems(member: OrgMember): DropdownMenuItem[][] {
  const isSelf = member.id === props.currentUserId
  const rows: DropdownMenuItem[][] = []

  if (member.status === 'invited') {
    rows.push([{
      label: 'Resend invite',
      icon: 'i-lucide-mail',
      onSelect: () => emit('resend', member)
    }, {
      label: 'Revoke invite',
      icon: 'i-lucide-x',
      color: 'error',
      onSelect: () => emit('revoke', member)
    }])
    return rows
  }

  if (member.status === 'active') {
    rows.push([{
      label: 'Suspend access',
      icon: 'i-lucide-pause',
      onSelect: () => {
        const result = setStatus(member.id, 'suspended')
        if (!result.ok) {
          toast.add({ title: result.error, color: 'error' })
          return
        }
        toast.add({ title: 'Member suspended', color: 'neutral' })
      }
    }])
  }

  if (member.status === 'suspended') {
    rows.push([{
      label: 'Restore access',
      icon: 'i-lucide-play',
      onSelect: () => {
        const result = setStatus(member.id, 'active')
        if (!result.ok) {
          toast.add({ title: result.error, color: 'error' })
          return
        }
        toast.add({ title: 'Access restored', color: 'success' })
      }
    }])
  }

  if (!isSelf) {
    rows.push([{
      label: 'Remove member',
      icon: 'i-lucide-user-minus',
      color: 'error',
      onSelect: () => emit('remove', member)
    }])
  }

  return rows
}

function statusColor(status: OrgMember['status']) {
  if (status === 'active') {
    return 'success' as const
  }
  if (status === 'invited') {
    return 'warning' as const
  }
  return 'neutral' as const
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
          v-bind="member.avatar"
          :alt="member.name"
          size="md"
        />

        <div class="min-w-0 text-sm">
          <div class="flex flex-wrap items-center gap-2">
            <p class="text-highlighted truncate font-medium">
              {{ member.name }}
              <span
                v-if="member.id === currentUserId"
                class="text-muted font-normal"
              >(you)</span>
            </p>
            <UBadge
              :label="member.status"
              :color="statusColor(member.status)"
              variant="subtle"
              size="sm"
              class="capitalize"
            />
          </div>
          <p class="text-muted truncate">
            {{ member.email }}
          </p>
        </div>
      </div>

      <div class="flex shrink-0 items-center gap-2 sm:gap-3">
        <USelect
          :model-value="member.role"
          :items="ROLE_OPTIONS"
          value-key="value"
          color="neutral"
          class="w-32"
          :disabled="member.status === 'invited'"
          @update:model-value="(value) => onRoleChange(member, value as MemberRole)"
        />

        <UDropdownMenu
          :items="menuItems(member)"
          :content="{ align: 'end' }"
        >
          <UButton
            icon="i-lucide-ellipsis-vertical"
            color="neutral"
            variant="ghost"
          />
        </UDropdownMenu>
      </div>
    </li>
  </ul>
</template>
