<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import type { Member } from '~/types'

defineProps<{
  members: Member[]
}>()

const toast = useToast()

const items = [{
  label: 'Edit member',
  onSelect: () => toast.add({ title: 'Not wired yet', color: 'neutral' })
}, {
  label: 'Remove member',
  color: 'error' as const,
  onSelect: () => toast.add({ title: 'Not wired yet', color: 'neutral' })
}] satisfies DropdownMenuItem[]
</script>

<template>
  <ul
    role="list"
    class="divide-y divide-default"
  >
    <li
      v-for="(member, index) in members"
      :key="index"
      class="flex items-center justify-between gap-3 px-4 py-3 sm:px-6"
    >
      <div class="flex min-w-0 items-center gap-3">
        <UAvatar
          v-bind="member.avatar"
          size="md"
        />

        <div class="min-w-0 text-sm">
          <p class="text-highlighted truncate font-medium">
            {{ member.name }}
          </p>
          <p class="text-muted truncate">
            {{ member.username }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <USelect
          :model-value="member.role"
          :items="['member', 'owner']"
          color="neutral"
          :ui="{ value: 'capitalize', item: 'capitalize' }"
        />

        <UDropdownMenu
          :items="items"
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
