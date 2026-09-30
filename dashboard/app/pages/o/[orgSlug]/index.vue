<script setup lang="ts">
import { formatDate } from '~/utils/billing'
import type { OverviewActivity } from '~/composables/useOrgOverview'
import { agentBadgeColor } from '~/utils/servers'

const { isNotificationsSlideoverOpen } = useDashboard()
const { selected, orgHref } = useOrganizations()
const { overview, isEmpty } = useOrgOverview()
const { currentPlan, subscription } = useBilling()
const toast = useToast()

const stats = computed(() => [{
  label: 'Servers',
  value: overview.value.servers,
  hint: isEmpty.value ? 'None yet' : 'In this organization'
}, {
  label: 'Sites',
  value: overview.value.sites,
  hint: isEmpty.value ? 'None yet' : 'Across all servers'
}, {
  label: 'Deploys (7d)',
  value: overview.value.deploys7d,
  hint: 'Last seven days'
}, {
  label: 'Open alerts',
  value: overview.value.openAlerts,
  hint: overview.value.openAlerts ? 'Needs attention' : 'All clear'
}])

function activityColor(status: OverviewActivity['status']) {
  if (status === 'success') {
    return 'success' as const
  }
  if (status === 'failed') {
    return 'error' as const
  }
  return 'primary' as const
}

function onTerraformSoon() {
  toast.add({
    title: 'Coming later',
    description: 'Terraform provisioning lands in a later phase.',
    color: 'neutral'
  })
}
</script>

<template>
  <UDashboardPanel id="home">
    <template #header>
      <UDashboardNavbar
        title="Dashboard"
        :ui="{ right: 'gap-3' }"
      >
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #right>
          <UTooltip
            text="Notifications"
            :shortcuts="['N']"
          >
            <UChip
              color="error"
              :show="overview.openAlerts > 0"
              inset
            >
              <UButton
                color="neutral"
                variant="ghost"
                square
                icon="i-lucide-bell"
                aria-label="Notifications"
                @click="isNotificationsSlideoverOpen = true"
              />
            </UChip>
          </UTooltip>

          <UButton
            :to="orgHref('/servers/new')"
            label="Add server"
            icon="i-lucide-plus"
          />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="mx-auto flex w-full max-w-6xl flex-col gap-8 p-4 sm:p-6 lg:p-8">
        <header class="flex flex-col gap-1">
          <h2 class="text-highlighted text-2xl font-semibold tracking-tight sm:text-3xl">
            {{ selected?.name }}
          </h2>
          <p class="text-muted max-w-2xl text-sm">
            Operations overview for this organization. Status, deploys, and alerts surface here first.
          </p>
        </header>

        <section
          aria-label="Key metrics"
          class="border-default grid grid-cols-2 gap-px overflow-hidden rounded-lg border bg-default sm:grid-cols-4"
        >
          <div
            v-for="stat in stats"
            :key="stat.label"
            class="bg-default flex flex-col gap-1 px-4 py-4 sm:px-5 sm:py-5"
          >
            <p class="text-muted text-xs font-medium tracking-wide uppercase">
              {{ stat.label }}
            </p>
            <p class="text-highlighted text-2xl font-semibold tabular-nums">
              {{ stat.value }}
            </p>
            <p class="text-muted text-xs">
              {{ stat.hint }}
            </p>
          </div>
        </section>

        <section
          v-if="isEmpty"
          class="border-default from-elevated/40 to-default relative overflow-hidden rounded-lg border bg-linear-to-br px-5 py-8 sm:px-8 sm:py-10"
        >
          <div class="relative z-10 flex max-w-xl flex-col gap-4">
            <div class="bg-primary/10 text-primary flex size-10 items-center justify-center rounded-lg">
              <UIcon
                name="i-lucide-server"
                class="size-5"
              />
            </div>
            <div class="flex flex-col gap-2">
              <h2 class="text-highlighted text-lg font-semibold">
                Add your first server
              </h2>
              <p class="text-muted text-sm leading-relaxed">
                Connect a machine over SSH or provision one with Terraform. Once the agent heartbeats, sites and deploys show up on this dashboard.
              </p>
            </div>
            <div class="flex flex-wrap gap-2">
              <UButton
                :to="orgHref('/servers/new')"
                label="Add server manually"
                icon="i-lucide-terminal"
              />
              <UButton
                label="Provision with Terraform"
                color="neutral"
                variant="outline"
                icon="i-lucide-cloud"
                @click="onTerraformSoon"
              />
            </div>
          </div>
        </section>

        <div
          v-else
          class="grid gap-6 lg:grid-cols-5"
        >
          <section
            aria-label="Servers"
            class="flex flex-col gap-3 lg:col-span-3"
          >
            <div class="flex items-end justify-between gap-3">
              <h2 class="text-highlighted text-sm font-semibold">
                Servers
              </h2>
              <UButton
                :to="orgHref('/servers')"
                label="View all"
                color="neutral"
                variant="link"
                size="sm"
                class="px-0"
              />
            </div>

            <ul class="border-default divide-y divide-default overflow-hidden rounded-lg border">
              <li
                v-for="server in overview.serversList"
                :key="server.id"
              >
                <NuxtLink
                  :to="orgHref(`/servers/${server.id}`)"
                  class="hover:bg-elevated/40 flex items-center justify-between gap-4 px-4 py-3.5 transition-colors"
                >
                  <div class="min-w-0">
                    <p class="text-highlighted truncate text-sm font-medium">
                      {{ server.hostname }}
                    </p>
                    <p class="text-muted text-xs">
                      {{ server.region }} · {{ server.sites }} {{ server.sites === 1 ? 'site' : 'sites' }}
                    </p>
                  </div>
                  <UBadge
                    :label="server.agent"
                    :color="agentBadgeColor(server.agent)"
                    variant="outline"
                    class="capitalize"
                  />
                </NuxtLink>
              </li>
            </ul>
          </section>

          <section
            aria-label="Alerts"
            class="flex flex-col gap-3 lg:col-span-2"
          >
            <h2 class="text-highlighted text-sm font-semibold">
              Alerts
            </h2>

            <ul
              v-if="overview.alerts.length"
              class="border-default divide-y divide-default overflow-hidden rounded-lg border"
            >
              <li
                v-for="alert in overview.alerts"
                :key="alert.id"
                class="flex gap-3 px-4 py-3.5"
              >
                <UIcon
                  :name="alert.severity === 'error' ? 'i-lucide-circle-alert' : 'i-lucide-triangle-alert'"
                  class="mt-0.5 size-4 shrink-0"
                  :class="alert.severity === 'error' ? 'text-error' : 'text-warning'"
                />
                <div class="min-w-0">
                  <p class="text-highlighted text-sm font-medium">
                    {{ alert.title }}
                  </p>
                  <p class="text-muted mt-0.5 text-xs leading-relaxed">
                    {{ alert.detail }}
                  </p>
                </div>
              </li>
            </ul>
            <div
              v-else
              class="border-default text-muted rounded-lg border border-dashed px-4 py-8 text-center text-sm"
            >
              No open alerts
            </div>
          </section>
        </div>

        <section
          v-if="!isEmpty"
          aria-label="Recent activity"
          class="flex flex-col gap-3"
        >
          <h2 class="text-highlighted text-sm font-semibold">
            Recent activity
          </h2>

          <ul class="border-default divide-y divide-default overflow-hidden rounded-lg border">
            <li
              v-for="item in overview.activity"
              :key="item.id"
              class="flex flex-wrap items-center justify-between gap-3 px-4 py-3.5"
            >
              <div class="flex min-w-0 items-center gap-3">
                <UBadge
                  :label="item.status"
                  :color="activityColor(item.status)"
                  variant="outline"
                  class="capitalize"
                />
                <div class="min-w-0 text-sm">
                  <p class="text-highlighted font-medium">
                    {{ item.label }}
                  </p>
                  <p class="text-muted truncate text-xs">
                    {{ item.target }}
                  </p>
                </div>
              </div>
              <time
                class="text-muted text-xs tabular-nums"
                :datetime="item.at"
              >
                {{ formatDate(item.at) }}
              </time>
            </li>
          </ul>
        </section>

        <section
          aria-label="Plan"
          class="border-default flex flex-col gap-4 rounded-lg border px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5"
        >
          <div class="min-w-0">
            <p class="text-highlighted text-sm font-medium">
              {{ currentPlan.name }} plan
            </p>
            <p class="text-muted text-xs">
              {{ subscription.usage.servers }} / {{ currentPlan.limits.find(l => l.key === 'servers')?.max }} servers
              · {{ subscription.usage.sites }} / {{ currentPlan.limits.find(l => l.key === 'sites')?.max }} sites
              · {{ subscription.usage.seats }} / {{ currentPlan.limits.find(l => l.key === 'seats')?.max }} seats
            </p>
          </div>
          <UButton
            :to="orgHref('/organization/billing')"
            label="Manage billing"
            color="neutral"
            variant="outline"
            size="sm"
          />
        </section>
      </div>
    </template>
  </UDashboardPanel>
</template>
