<script setup lang="ts">
import { formatDate } from '~/utils/billing'
import { ApiError } from '~/utils/api'
import { Ability } from '~/utils/auth/abilities'
import {
  ADDON_OPTIONS,
  DEFAULT_NODE_VERSIONS,
  DEFAULT_PHP_VERSIONS,
  provisionBadgeColor,
  selectionFromJobPayload,
  selectionFromRuntimes,
  type AgentJob,
  type ProvisionAddons,
  type ProvisionInput,
  type RuntimeCatalog
} from '~/utils/runtimes'

const route = useRoute()
const toast = useToast()
const { can } = useGate()
const { orgHref } = useOrganizations()
const serverId = computed(() => String(route.params.serverId || ''))

const { getById, fetchAll, bootstrap } = useServers()
const {
  provision,
  getCatalog,
  getLatestJob,
  getProvisionSecrets,
  dismissProvisionSecrets,
  pollJob
} = useServerRuntimes(serverId)

await fetchAll()

const server = computed(() => getById(serverId.value))
const bootstrapping = ref(false)
const provisioning = ref(false)
const activeJob = ref<AgentJob | null>(null)
const catalog = ref<RuntimeCatalog>({ php: DEFAULT_PHP_VERSIONS, node: DEFAULT_NODE_VERSIONS })
const secretsOpen = ref(false)
const secrets = ref<Record<string, Record<string, string>> | null>(null)
const secretsLoading = ref(false)

const secretRows = computed(() => {
  if (!secrets.value) {
    return []
  }
  return Object.entries(secrets.value).map(([service, fields]) => ({
    service,
    fields: Object.entries(fields).map(([key, value]) => ({ key, value }))
  }))
})

async function revealSecrets() {
  secretsLoading.value = true
  try {
    secrets.value = await getProvisionSecrets()
    secretsOpen.value = true
  } catch (error) {
    if (!(error instanceof ApiError && error.status === 404)) {
      toast.add({
        title: 'Could not load credentials',
        description: error instanceof Error ? error.message : undefined,
        color: 'error'
      })
    }
  } finally {
    secretsLoading.value = false
  }
}

async function onDismissSecrets() {
  try {
    await dismissProvisionSecrets()
    secretsOpen.value = false
    secrets.value = null
    await fetchAll(true)
  } catch (error) {
    toast.add({
      title: 'Could not dismiss credentials',
      description: error instanceof Error ? error.message : undefined,
      color: 'error'
    })
  }
}

async function maybeRevealSecrets() {
  await fetchAll(true)
  if (server.value?.hasProvisionSecrets) {
    await revealSecrets()
  }
}

function copySecret(value: string) {
  navigator.clipboard?.writeText(value)
  toast.add({ title: 'Copied', color: 'success' })
}

function applyJob(job: AgentJob | null | undefined) {
  activeJob.value = job ?? null
}

function applyProvisionInput(sel: ProvisionInput | null | undefined) {
  if (!sel) {
    return false
  }
  phpSelected.value = sel.phpVersions
  nodeSelected.value = sel.nodeVersions
  Object.assign(addons, sel.addons)
  return true
}

/** Prefer installed inventory; fall back to last provision job payload. */
function syncProvisionForm(job?: AgentJob | null) {
  if (server.value?.provisionStatus === 'ready') {
    if (applyProvisionInput(selectionFromRuntimes(server.value.runtimes))) {
      return
    }
  }
  if (job && applyProvisionInput(selectionFromJobPayload(job.payload))) {
    return
  }
  if (!phpSelected.value.length) {
    phpSelected.value = [...DEFAULT_PHP_VERSIONS]
  }
  if (!nodeSelected.value.length) {
    nodeSelected.value = [...DEFAULT_NODE_VERSIONS]
  }
}

function jobSummary(job: AgentJob) {
  const head = `Job #${job.id} · ${job.status}`
  return job.message ? `${head}\n${job.message}` : head
}

const phpSelected = ref<string[]>([])
const nodeSelected = ref<string[]>([])
const addons = reactive<ProvisionAddons>({
  redis: false,
  mysql: false,
  mariadb: false,
  postgres: false
})

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
    label: 'Agent',
    value: s.agentVersion || s.agent
  }, {
    label: 'Stack',
    value: s.provisionStatus
  }, {
    label: 'Added',
    value: formatDate(s.createdAt)
  }]
})

const canUpdate = computed(() => can(Ability.ServersUpdate))

const canProvision = computed(() =>
  !!server.value
  && server.value.agent === 'online'
  && canUpdate.value
  && !['pending', 'running'].includes(server.value.provisionStatus)
)

const provisionLabel = computed(() => {
  const status = server.value?.provisionStatus
  if (status === 'ready' || status === 'failed') {
    return 'Re-provision'
  }
  return 'Provision server'
})

if (server.value?.agent === 'online') {
  try {
    catalog.value = await getCatalog()
  } catch {
    // keep defaults
  }
}

const latest = await getLatestJob('server.provision').catch(() => null)
if (latest) {
  applyJob(latest)
}
syncProvisionForm(latest)

if (latest && (latest.status === 'pending' || latest.status === 'running')) {
  provisioning.value = true
  pollJob(latest.id, {
    onTick: applyJob
  }).then(async (done) => {
    applyJob(done)
    if (done?.status === 'succeeded') {
      toast.add({ title: 'Server provisioned', description: done.message, color: 'success' })
      await maybeRevealSecrets()
    } else if (done?.status === 'failed') {
      toast.add({ title: 'Provision failed', description: done.message || done.status, color: 'error' })
    }
  }).finally(async () => {
    provisioning.value = false
    await fetchAll(true)
    syncProvisionForm(activeJob.value)
  })
} else if (server.value?.hasProvisionSecrets) {
  await revealSecrets()
}

// Keep checkboxes aligned with inventory after runtime installs / refresh.
watch(
  () => {
    const s = server.value
    if (!s || s.provisionStatus !== 'ready' || provisioning.value) {
      return ''
    }
    return [
      s.runtimes.php.versions.map(v => `${v.version}:${v.status}`).join(','),
      s.runtimes.node.versions.map(v => `${v.version}:${v.status}`).join(','),
      s.runtimes.addons.join(',')
    ].join('|')
  },
  (key, prev) => {
    if (key && key !== prev) {
      syncProvisionForm(activeJob.value)
    }
  }
)

function togglePhp(ver: string, on: boolean) {
  if (on) {
    phpSelected.value = [...new Set([...phpSelected.value, ver])]
  } else {
    phpSelected.value = phpSelected.value.filter(v => v !== ver)
  }
}

function toggleNode(ver: string, on: boolean) {
  if (on) {
    nodeSelected.value = [...new Set([...nodeSelected.value, ver])]
  } else {
    nodeSelected.value = nodeSelected.value.filter(v => v !== ver)
  }
}

async function onReinstall() {
  const s = server.value
  if (!s) {
    return
  }
  bootstrapping.value = true
  try {
    const result = await bootstrap(s.id)
    toast.add({
      title: 'Agent reinstalled',
      description: result.message,
      color: 'success'
    })
    await fetchAll(true)
  } catch (error) {
    const detail = error instanceof ApiError && error.data?.message
      ? String(error.data.message)
      : error instanceof Error ? error.message : 'Bootstrap failed'
    toast.add({
      title: 'Agent install failed',
      description: detail,
      color: 'error'
    })
  } finally {
    bootstrapping.value = false
  }
}

async function onProvision() {
  if (!canProvision.value) {
    return
  }
  if (!phpSelected.value.length || !nodeSelected.value.length) {
    toast.add({
      title: 'Select versions',
      description: 'Pick at least one PHP and one Node version.',
      color: 'warning'
    })
    return
  }

  provisioning.value = true
  activeJob.value = null
  try {
    const job = await provision({
      phpVersions: phpSelected.value,
      nodeVersions: nodeSelected.value,
      addons: { ...addons }
    })
    applyJob(job)
    await fetchAll(true)
    const done = await pollJob(job.id, { onTick: applyJob })
    applyJob(done)
    toast.add({
      title: done?.status === 'succeeded' ? 'Server provisioned' : 'Provision failed',
      description: done?.message || done?.status,
      color: done?.status === 'succeeded' ? 'success' : 'error'
    })
    if (done?.status === 'succeeded') {
      await maybeRevealSecrets()
    }
  } catch (error) {
    const detail = error instanceof ApiError && error.data?.message
      ? String(error.data.message)
      : error instanceof Error ? error.message : 'Provision failed'
    activeJob.value = {
      id: 0,
      kind: 'server.provision',
      payload: {},
      status: 'failed',
      message: detail,
      exitCode: 1,
      createdAt: new Date().toISOString()
    }
    toast.add({
      title: 'Provision failed',
      description: detail,
      color: 'error'
    })
  } finally {
    provisioning.value = false
    await fetchAll(true)
    syncProvisionForm(activeJob.value)
  }
}
</script>

<template>
  <div
    v-if="server"
    class="flex flex-col gap-6"
  >
    <section
      aria-label="Server summary"
      class="border-default grid grid-cols-2 gap-px overflow-hidden rounded-lg border sm:grid-cols-3 lg:grid-cols-7"
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
      v-else-if="server.agent === 'pending'"
      color="warning"
      variant="subtle"
      icon="i-lucide-loader-circle"
      title="Agent pending"
      description="Install did not finish or no heartbeat yet. Reinstall over SSH from here."
    />

    <div class="border-default flex flex-wrap items-center justify-between gap-3 rounded-lg border px-4 py-3 sm:px-5">
      <div class="min-w-0">
        <p class="text-highlighted text-sm font-semibold">
          Agent
        </p>
        <p class="text-muted text-xs">
          Install runs on the API over SSH.
        </p>
      </div>
      <UButton
        label="Reinstall agent"
        color="neutral"
        variant="outline"
        size="sm"
        :loading="bootstrapping"
        :disabled="!canUpdate"
        @click="onReinstall"
      />
    </div>

    <section
      v-if="server.agent === 'online'"
      class="border-default flex flex-col gap-4 rounded-lg border px-4 py-4 sm:px-5"
    >
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-2">
            <p class="text-highlighted text-sm font-semibold">
              Provision stack
            </p>
            <UBadge
              :label="server.provisionStatus"
              :color="provisionBadgeColor(server.provisionStatus)"
              variant="subtle"
              size="sm"
              class="capitalize"
            />
          </div>
          <p class="text-muted mt-1 text-xs">
            Install nginx, certbot, Composer, multi-version PHP (Sury) and Node (fnm), plus optional databases.
          </p>
        </div>
        <UButton
          :label="provisionLabel"
          color="primary"
          size="sm"
          class="shrink-0"
          :loading="provisioning"
          :disabled="!canProvision"
          @click="onProvision"
        />
      </div>

      <div class="grid gap-4 md:grid-cols-2">
        <div>
          <p class="text-highlighted mb-2 text-xs font-medium tracking-wide uppercase">
            PHP versions
          </p>
          <div class="flex flex-wrap gap-3">
            <UCheckbox
              v-for="ver in catalog.php"
              :key="`php-${ver}`"
              :model-value="phpSelected.includes(ver)"
              :label="ver"
              :disabled="provisioning || !canProvision"
              @update:model-value="togglePhp(ver, $event === true)"
            />
          </div>
        </div>
        <div>
          <p class="text-highlighted mb-2 text-xs font-medium tracking-wide uppercase">
            Node versions
          </p>
          <div class="flex flex-wrap gap-3">
            <UCheckbox
              v-for="ver in catalog.node"
              :key="`node-${ver}`"
              :model-value="nodeSelected.includes(ver)"
              :label="ver"
              :disabled="provisioning || !canProvision"
              @update:model-value="toggleNode(ver, $event === true)"
            />
          </div>
        </div>
      </div>

      <div>
        <p class="text-highlighted mb-2 text-xs font-medium tracking-wide uppercase">
          Add-ons
        </p>
        <div class="flex flex-wrap gap-3">
          <UCheckbox
            v-for="opt in ADDON_OPTIONS"
            :key="opt.key"
            v-model="addons[opt.key]"
            :label="opt.label"
            :disabled="provisioning || !canProvision"
          />
        </div>
        <p class="text-muted mt-1 text-xs">
          MariaDB covers MySQL-compatible apps.
        </p>
      </div>

      <div
        v-if="activeJob"
        class="border-default bg-elevated/40 rounded-md border px-3 py-2"
      >
        <p class="text-highlighted text-xs font-medium">
          Provision job
        </p>
        <pre class="text-muted mt-1 max-h-48 overflow-auto whitespace-pre-wrap font-mono text-xs">{{ jobSummary(activeJob) }}</pre>
      </div>
      <UAlert
        v-if="activeJob?.status === 'failed' && activeJob.message"
        color="error"
        variant="subtle"
        icon="i-lucide-circle-alert"
        title="Provision failed"
        :description="activeJob.message"
      />
      <UAlert
        v-if="server.hasProvisionSecrets && !secretsOpen"
        color="warning"
        variant="subtle"
        icon="i-lucide-key-round"
        title="Addon credentials waiting"
        description="Save the generated database passwords now. They are cleared after you dismiss this."
      >
        <template #actions>
          <UButton
            label="Show credentials"
            size="sm"
            color="neutral"
            variant="outline"
            :loading="secretsLoading"
            @click="revealSecrets"
          />
        </template>
      </UAlert>

      <p
        v-if="server.provisionStatus === 'ready'"
        class="text-muted text-xs"
      >
        Manage installed versions on the
        <NuxtLink
          :to="orgHref(`/servers/${serverId}/runtimes`)"
          class="text-highlighted underline"
        >
          Runtimes
        </NuxtLink>
        tab.
      </p>
    </section>

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

    <UModal
      v-model:open="secretsOpen"
      title="Save addon credentials"
      description="Generated during provision. Copy them now — dismiss clears them from the control plane."
    >
      <template #body>
        <div class="flex flex-col gap-4">
          <UAlert
            color="warning"
            variant="subtle"
            icon="i-lucide-triangle-alert"
            title="One-time reveal"
            description="After you dismiss, these passwords are not shown again. Reset them on the server if you lose them."
          />
          <div
            v-for="row in secretRows"
            :key="row.service"
            class="border-default rounded-md border px-3 py-2"
          >
            <p class="text-highlighted text-xs font-semibold tracking-wide uppercase">
              {{ row.service }}
            </p>
            <div
              v-for="field in row.fields"
              :key="field.key"
              class="mt-2 flex items-center justify-between gap-2"
            >
              <div class="min-w-0">
                <p class="text-muted text-xs">
                  {{ field.key }}
                </p>
                <p class="text-highlighted truncate font-mono text-sm">
                  {{ field.value }}
                </p>
              </div>
              <UButton
                icon="i-lucide-copy"
                color="neutral"
                variant="ghost"
                size="xs"
                aria-label="Copy"
                @click="copySecret(field.value)"
              />
            </div>
          </div>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton
            label="I've saved these"
            color="primary"
            @click="onDismissSecrets"
          />
        </div>
      </template>
    </UModal>
  </div>
</template>
