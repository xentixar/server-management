<script setup lang="ts">
import { formatDate } from '~/utils/billing'

const route = useRoute()
const serverId = computed(() => String(route.params.serverId || ''))

const { server, run, progress, start, replay } = useServerProvision(serverId)

const facts = computed(() => {
  const s = server.value
  if (!s) {
    return []
  }
  return [{
    label: 'Address',
    value: `${s.ip}:${s.sshPort}`
  }, {
    label: 'SSH user',
    value: s.sshUser
  }, {
    label: 'Auth',
    value: s.authMethod === 'key' ? (s.hasKeyUploaded ? 'SSH key' : 'SSH key (missing)') : 'Password'
  }, {
    label: 'Region',
    value: s.region || 'n/a'
  }, {
    label: 'Sites',
    value: String(s.sites)
  }, {
    label: 'Added',
    value: formatDate(s.createdAt)
  }]
})
</script>

<template>
  <div
    v-if="server"
    class="flex flex-col gap-6"
  >
    <section
      aria-label="Server summary"
      class="border-default grid grid-cols-2 gap-px overflow-hidden rounded-lg border sm:grid-cols-3 lg:grid-cols-6"
    >
      <div
        v-for="fact in facts"
        :key="fact.label"
        class="bg-default px-4 py-3"
      >
        <p class="text-muted text-xs font-medium tracking-wide uppercase">
          {{ fact.label }}
        </p>
        <p class="text-highlighted mt-1 truncate text-sm font-medium">
          {{ fact.value }}
        </p>
      </div>
    </section>

    <UAlert
      v-if="server.agent === 'offline'"
      color="error"
      variant="subtle"
      icon="i-lucide-circle-alert"
      title="Agent offline"
      description="No recent heartbeat. Reinstall the agent or check the host."
    />
    <UAlert
      v-else-if="server.agent === 'degraded'"
      color="warning"
      variant="subtle"
      icon="i-lucide-triangle-alert"
      title="Agent degraded"
      description="Heartbeats are intermittent. Check load and network on the host."
    />
    <UAlert
      v-else-if="server.agent === 'online'"
      color="success"
      variant="subtle"
      icon="i-lucide-check"
      title="Agent online"
      :description="`Last known status for ${server.hostname}.`"
    />

    <OrganizationServerProvisionPanel
      :steps="run.steps"
      :logs="run.logs"
      :status="run.status"
      :progress="progress"
      :auto-start="server.agent === 'pending'"
      @start="start()"
      @replay="replay()"
    />

    <section
      v-if="server.labels.length || server.notes"
      class="border-default rounded-lg border px-4 py-4 sm:px-5"
    >
      <h2 class="text-highlighted mb-3 text-sm font-semibold">
        Metadata
      </h2>
      <div
        v-if="server.labels.length"
        class="mb-3 flex flex-wrap gap-2"
      >
        <UBadge
          v-for="label in server.labels"
          :key="label"
          :label="label"
          color="neutral"
          variant="subtle"
        />
      </div>
      <p
        v-if="server.notes"
        class="text-muted text-sm"
      >
        {{ server.notes }}
      </p>
    </section>
  </div>
</template>
