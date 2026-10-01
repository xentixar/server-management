<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

import { Ability } from '~/utils/auth/abilities'

const open = ref(false)
const { orgHref } = useOrganizations()
const { can } = useGate()

const links = computed(() => {
  const orgChildren: NavigationMenuItem[] = [{
    label: 'General',
    to: orgHref('/organization'),
    exact: true,
    onSelect: () => {
      open.value = false
    }
  }]

  if (can(Ability.MembersView)) {
    orgChildren.push({
      label: 'Members',
      to: orgHref('/organization/members'),
      onSelect: () => {
        open.value = false
      }
    })
  }

  if (can(Ability.BillingView)) {
    orgChildren.push({
      label: 'Billing',
      to: orgHref('/organization/billing'),
      onSelect: () => {
        open.value = false
      }
    })
  }

  return [[{
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
    type: 'trigger',
    children: orgChildren
  }]] satisfies NavigationMenuItem[][]
})

const groups = computed(() => [{
  id: 'links',
  label: 'Go to',
  items: links.value.flat().flatMap((item) => {
    const leaves = item.children?.length ? item.children : [item]
    return leaves.map(leaf => ({
      label: leaf.label,
      to: leaf.to,
      icon: leaf.icon,
      onSelect: leaf.onSelect
    }))
  })
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

    <main class="flex min-h-0 min-w-0 flex-1 flex-col">
      <slot />
    </main>

    <NotificationsSlideover />
  </UDashboardGroup>
</template>
