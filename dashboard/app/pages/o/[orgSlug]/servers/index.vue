<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import type { Server } from '~/utils/servers'
import { agentBadgeColor } from '~/utils/servers'

const { servers, atCapacity, serverLimit, remove, fetchAll } = useServers()
const { orgHref } = useOrganizations()
const { currentPlan } = useBilling()
const toast = useToast()

await fetchAll()

const q = ref('')
const deleteOpen = ref(false)
const serverToDelete = ref<Server | null>(null)

const filtered = computed(() => {
  const query = q.value.trim().toLowerCase()
  if (!query) {
    return servers.value
  }
  return servers.value.filter((server) => {
    return server.hostname.toLowerCase().includes(query)
      || server.ip.toLowerCase().includes(query)
      || (server.region || '').toLowerCase().includes(query)
  })
})

function menuItems(server: Server): DropdownMenuItem[][] {
  return [[{
    label: 'Open',
    icon: 'i-lucide-arrow-up-right',
    onSelect: () => navigateTo(orgHref(`/servers/${server.id}`))
  }, {
    label: 'Delete server',
    icon: 'i-lucide-trash-2',
    color: 'error',
    onSelect: () => {
      serverToDelete.value = server
      deleteOpen.value = true
    }
  }]]
}

async function onConfirmDelete() {
  if (!serverToDelete.value) {
    return
  }
  try {
    const removed = await remove(serverToDelete.value.id)
    toast.add({
      title: 'Server deleted',
      description: removed.hostname,
      color: 'neutral'
    })
    serverToDelete.value = null
  } catch (error) {
    toast.add({
      title: error instanceof Error ? error.message : 'Failed to delete server',
      color: 'error'
    })
  }
}

function onAddClick() {
  if (atCapacity.value) {
    toast.add({
      title: 'Plan limit reached',
      description: `${currentPlan.value.name} allows ${serverLimit.value} servers. Upgrade in billing.`,
      color: 'error'
    })
  }
}
</script>

<template>
  <UDashboardPanel id="servers">
    <template #header>
      <UDashboardNavbar
        title="Servers"
        :ui="{ right: 'gap-2' }"
      >
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #right>
          <UButton
            :to="atCapacity ? undefined : orgHref('/servers/new')"
            label="Add server"
            icon="i-lucide-plus"
            @click="onAddClick"
          />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="mx-auto flex w-full max-w-6xl flex-col gap-6 p-4 sm:p-6 lg:p-8">
        <template v-if="!servers.length">
          <section class="border-default from-elevated/40 to-default rounded-lg border bg-linear-to-br px-5 py-10 sm:px-8">
            <div class="flex max-w-xl flex-col gap-4">
              <div class="bg-primary/10 text-primary flex size-10 items-center justify-center rounded-lg">
                <UIcon
                  name="i-lucide-server"
                  class="size-5"
                />
              </div>
              <div class="flex flex-col gap-2">
                <h2 class="text-highlighted text-lg font-semibold">
                  No servers yet
                </h2>
                <p class="text-muted text-sm leading-relaxed">
                  Add an existing machine over SSH. Once the agent heartbeats, you can provision the stack and manage sites from here.
                </p>
              </div>
              <div class="flex flex-wrap gap-2">
                <UButton
                  :to="orgHref('/servers/new')"
                  label="Add server"
                  icon="i-lucide-terminal"
                />
              </div>
            </div>
          </section>
        </template>

        <template v-else>
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <p class="text-muted text-sm">
              {{ servers.length }} / {{ serverLimit }} on {{ currentPlan.name }}
            </p>
            <UInput
              v-model="q"
              icon="i-lucide-search"
              placeholder="Search hostname or IP"
              class="w-full sm:max-w-xs"
            />
          </div>

          <ul class="border-default divide-y divide-default overflow-hidden rounded-lg border">
            <li
              v-for="server in filtered"
              :key="server.id"
              class="hover:bg-elevated/40 flex flex-wrap items-center justify-between gap-3 px-4 py-3.5 transition-colors sm:px-5"
            >
              <NuxtLink
                :to="orgHref(`/servers/${server.id}`)"
                class="min-w-0 flex-1"
              >
                <div class="flex flex-wrap items-center gap-2">
                  <p class="text-highlighted truncate text-sm font-medium">
                    {{ server.hostname }}
                  </p>
                  <UBadge
                    :label="server.agent"
                    :color="agentBadgeColor(server.agent)"
                    variant="subtle"
                    size="sm"
                    class="capitalize"
                  />
                </div>
                <p class="text-muted mt-0.5 text-xs">
                  {{ server.ip }}:{{ server.sshPort }}
                  · {{ server.sshUser }}
                  · {{ server.region || 'n/a' }}
                  · {{ server.sites }} {{ server.sites === 1 ? 'site' : 'sites' }}
                </p>
              </NuxtLink>

              <UDropdownMenu
                :items="menuItems(server)"
                :content="{ align: 'end' }"
              >
                <UButton
                  icon="i-lucide-ellipsis-vertical"
                  color="neutral"
                  variant="ghost"
                  aria-label="Server actions"
                />
              </UDropdownMenu>
            </li>
          </ul>

          <p
            v-if="!filtered.length"
            class="text-muted text-sm"
          >
            No servers match your search.
          </p>
        </template>
      </div>

      <OrganizationDeleteServerModal
        v-model:open="deleteOpen"
        :server="serverToDelete"
        @confirm="onConfirmDelete"
      />
    </template>
  </UDashboardPanel>
</template>
