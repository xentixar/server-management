<script setup lang="ts">
import type { Notification } from '~/types'
import { formatDate } from '~/utils/billing'

const { isNotificationsSlideoverOpen } = useDashboard()

const { data: notifications } = await useFetch<Notification[]>('/api/notifications')
</script>

<template>
  <USlideover
    v-model:open="isNotificationsSlideoverOpen"
    title="Notifications"
  >
    <template #body>
      <p
        v-if="!notifications?.length"
        class="text-muted text-sm"
      >
        No notifications yet.
      </p>

      <template v-else>
        <div
          v-for="notification in notifications"
          :key="notification.id"
          class="relative -mx-3 flex items-start gap-3 rounded-md px-3 py-2.5 first:-mt-3 last:-mb-3 hover:bg-elevated/50"
        >
          <span
            class="mt-1.5 size-2 shrink-0 rounded-full"
            :class="notification.unread ? 'bg-primary' : 'bg-transparent'"
            aria-hidden="true"
          />

          <div class="min-w-0 flex-1 text-sm">
            <p class="flex items-baseline justify-between gap-3">
              <span class="text-highlighted font-medium">{{ notification.title }}</span>
              <time
                :datetime="notification.date"
                class="text-muted shrink-0 text-xs"
              >
                {{ formatDate(notification.date) }}
              </time>
            </p>
            <p
              v-if="notification.body"
              class="text-muted mt-0.5"
            >
              {{ notification.body }}
            </p>
          </div>
        </div>
      </template>
    </template>
  </USlideover>
</template>
