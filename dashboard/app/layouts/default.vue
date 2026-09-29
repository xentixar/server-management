<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

const open = ref(false)
const { orgHref } = useOrganizations()

const links = computed(() => [[{
  label: 'Home',
  icon: 'i-lucide-house',
  to: orgHref(),
  exact: true,
  onSelect: () => {
    open.value = false
  }
}, {
  label: 'Servers',
  icon: 'i-lucide-server',
  to: orgHref('/servers'),
  onSelect: () => {
    open.value = false
  }
}, {
  label: 'Organization',
  to: orgHref('/organization'),
  icon: 'i-lucide-building-2',
  defaultOpen: true,
  type: 'trigger',
  children: [{
    label: 'General',
    to: orgHref('/organization'),
    exact: true,
    onSelect: () => {
      open.value = false
    }
  }, {
    label: 'Members',
    to: orgHref('/organization/members'),
    onSelect: () => {
      open.value = false
    }
  }, {
    label: 'Billing',
    to: orgHref('/organization/billing'),
    onSelect: () => {
      open.value = false
    }
  }]
}]] satisfies NavigationMenuItem[][])

const groups = computed(() => [{
  id: 'links',
  label: 'Go to',
  items: links.value.flat()
}])
</script>

<template>
  <UDashboardGroup unit="rem">
    <UDashboardSidebar
      id="default"
      v-model:open="open"
      collapsible
      resizable
      class="bg-elevated/25"
      :ui="{ footer: 'lg:border-t lg:border-default' }"
    >
      <template #header="{ collapsed }">
        <OrganizationsMenu :collapsed="collapsed" />
      </template>

      <template #default="{ collapsed }">
        <UDashboardSearchButton
          :collapsed="collapsed"
          class="bg-transparent ring-default"
        />

        <UNavigationMenu
          :collapsed="collapsed"
          :items="links[0]"
          orientation="vertical"
          tooltip
          popover
        />
      </template>

      <template #footer="{ collapsed }">
        <UserMenu :collapsed="collapsed" />
      </template>
    </UDashboardSidebar>

    <UDashboardSearch :groups="groups" />

    <slot />

    <NotificationsSlideover />
  </UDashboardGroup>
</template>
