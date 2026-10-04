<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

const route = useRoute()
const { orgHref } = useOrganizations()

const links = computed(() => [[{
  label: 'General',
  icon: 'i-lucide-building-2',
  to: orgHref('/organization'),
  exact: true
}, {
  label: 'Members',
  icon: 'i-lucide-users',
  to: orgHref('/organization/members')
}, {
  label: 'Git',
  icon: 'i-lucide-git-branch',
  to: orgHref('/organization/git')
}, {
  label: 'Billing',
  icon: 'i-lucide-credit-card',
  to: orgHref('/organization/billing')
}]] satisfies NavigationMenuItem[][])

const bodyClass = computed(() => {
  const git = route.path.includes('/organization/git')
  return git
    ? 'mx-auto flex w-full flex-col gap-4 sm:gap-6 lg:max-w-5xl lg:gap-12'
    : 'mx-auto flex w-full flex-col gap-4 sm:gap-6 lg:max-w-2xl lg:gap-12'
})
</script>

<template>
  <UDashboardPanel
    id="organization"
    :ui="{ body: 'lg:py-12' }"
  >
    <template #header>
      <UDashboardNavbar title="Organization">
        <template #leading>
          <UDashboardSidebarCollapse />
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
      <div :class="bodyClass">
        <NuxtPage />
      </div>
    </template>
  </UDashboardPanel>
</template>
