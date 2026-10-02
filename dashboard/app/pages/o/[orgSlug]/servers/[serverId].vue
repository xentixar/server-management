<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import { agentBadgeColor } from '~/utils/servers'

const route = useRoute()
const { orgHref } = useOrganizations()
const { getById, fetchAll } = useServers()

await fetchAll()

const serverId = computed(() => String(route.params.serverId || ''))
const server = computed(() => getById(serverId.value))
const provisioned = computed(() => server.value?.provisionStatus === 'ready')

if (!server.value) {
  await navigateTo(orgHref('/servers'), { replace: true })
}

const links = computed(() => {
  const base = orgHref(`/servers/${serverId.value}`)
  const ready = provisioned.value
  return [[{
    label: 'Overview',
    icon: 'i-lucide-layout-dashboard',
    to: base,
    exact: true
  }, {
    label: 'Sites',
    icon: 'i-lucide-globe',
    to: ready ? `${base}/sites` : undefined,
    disabled: !ready
  }, {
    label: 'Runtimes',
    icon: 'i-lucide-boxes',
    to: ready ? `${base}/runtimes` : undefined,
    disabled: !ready
  }]] satisfies NavigationMenuItem[][]
})

watch([() => route.path, provisioned], ([path, ready]) => {
  if (ready || !server.value) {
    return
  }
  const prefix = orgHref(`/servers/${serverId.value}`)
  if (path.startsWith(`${prefix}/sites`) || path.startsWith(`${prefix}/runtimes`)) {
    navigateTo(prefix, { replace: true })
  }
}, { immediate: true })
</script>

<template>
  <UDashboardPanel
    v-if="server"
    id="server-detail"
  >
    <template #header>
      <UDashboardNavbar :ui="{ right: 'gap-2' }">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #title>
          <!-- phrasing-only: UDashboardNavbar wraps #title in <h1> -->
          <span class="inline-flex min-w-0 max-w-full flex-wrap items-center gap-2">
            <NuxtLink
              :to="orgHref('/servers')"
              class="text-muted hover:text-highlighted text-sm font-normal"
            >
              Servers
            </NuxtLink>
            <span class="text-muted text-sm font-normal">/</span>
            <span class="truncate text-sm font-semibold">
              {{ server.hostname }}
            </span>
          </span>
        </template>

        <template #right>
          <UBadge
            :label="server.agent"
            :color="agentBadgeColor(server.agent)"
            variant="subtle"
            size="sm"
            class="capitalize"
          />
        </template>
      </UDashboardNavbar>

      <UDashboardToolbar>
        <UNavigationMenu
          :items="links"
          highlight
          class="-mx-1 flex-1"
        />
      </UDashboardToolbar>
    </template>

    <template #body>
      <div class="mx-auto flex w-full max-w-6xl flex-col gap-6 p-4 sm:p-6 lg:p-8">
        <NuxtPage />
      </div>
    </template>
  </UDashboardPanel>
</template>
