<script setup lang="ts">
import { ApiError } from '~/utils/api'
import { Ability } from '~/utils/auth/abilities'
import {
  DEFAULT_NODE_VERSIONS,
  DEFAULT_PHP_VERSIONS,
  RUNTIME_INSTALL_KIND,
  addonLabel,
  jobLabel,
  runtimeStatusColor,
  type AgentJob,
  type RuntimeCatalog,
  type RuntimeName,
  type RuntimeVersion
} from '~/utils/runtimes'

const route = useRoute()
const toast = useToast()
const { can } = useGate()
const { orgHref } = useOrganizations()
const serverId = computed(() => String(route.params.serverId || ''))

const { getById, fetchAll } = useServers()
const {
  getCatalog,
  getLatestJob,
  installRuntime,
  setDefault,
  pollJob
} = useServerRuntimes(serverId)

await fetchAll()

const server = computed(() => getById(serverId.value))
const catalog = ref<RuntimeCatalog>({ php: DEFAULT_PHP_VERSIONS, node: DEFAULT_NODE_VERSIONS })
const phpToAdd = ref('')
const nodeToAdd = ref('')
const busy = reactive<Record<RuntimeName, boolean>>({ php: false, node: false })
const activeJob = ref<AgentJob | null>(null)
const canManage = computed(() => can(Ability.ServersUpdate) && server.value?.agent === 'online')

function takenVersions(runtime: RuntimeName) {
  const family = runtime === 'php' ? server.value?.runtimes.php : server.value?.runtimes.node
  return new Set(
    (family?.versions ?? [])
      .filter(v => v.status === 'installed' || v.status === 'installing')
      .map(v => v.version)
  )
}

const phpAvailable = computed(() => catalog.value.php.filter(v => !takenVersions('php').has(v)))
const nodeAvailable = computed(() => catalog.value.node.filter(v => !takenVersions('node').has(v)))

const phpSelectItems = computed(() => phpAvailable.value.map(v => ({ label: v, value: v })))
const nodeSelectItems = computed(() => nodeAvailable.value.map(v => ({ label: v, value: v })))

const addonItems = computed(() =>
  (server.value?.runtimes.addons ?? []).map(key => ({
    key,
    label: addonLabel(key)
  }))
)

function jobSummary(job: AgentJob) {
  const head = jobLabel(job)
  return job.message ? `${head}\n${job.message}` : head
}

function sortVersions(versions: RuntimeVersion[]) {
  return [...versions].sort((a, b) => b.version.localeCompare(a.version, undefined, { numeric: true }))
}

watch(phpAvailable, (list) => {
  if (!list.includes(phpToAdd.value)) {
    phpToAdd.value = list[0] || ''
  }
}, { immediate: true })

watch(nodeAvailable, (list) => {
  if (!list.includes(nodeToAdd.value)) {
    nodeToAdd.value = list[0] || ''
  }
}, { immediate: true })

try {
  catalog.value = await getCatalog()
} catch {
  // defaults
}

async function trackJob(runtime: RuntimeName, job: AgentJob) {
  busy[runtime] = true
  activeJob.value = job
  try {
    const done = await pollJob(job.id, {
      onTick: (j) => {
        activeJob.value = j
      }
    })
    activeJob.value = done ?? null
    return done
  } finally {
    busy[runtime] = false
    await fetchAll(true)
  }
}

async function trackInstall(runtime: RuntimeName, job: AgentJob) {
  const done = await trackJob(runtime, job)
  const version = typeof done?.payload?.version === 'string' ? done.payload.version : ''
  toast.add({
    title: done?.status === 'succeeded'
      ? (version ? `${runtime} ${version} installed` : `${runtime} installed`)
      : 'Install failed',
    description: done?.message || done?.status,
    color: done?.status === 'succeeded' ? 'success' : 'error'
  })
}

for (const runtime of ['php', 'node'] as RuntimeName[]) {
  const latest = await getLatestJob(RUNTIME_INSTALL_KIND[runtime]).catch(() => null)
  if (latest && (latest.status === 'pending' || latest.status === 'running')) {
    void trackInstall(runtime, latest)
  }
}

async function onInstall(runtime: RuntimeName, version: string) {
  if (!version || !canManage.value || busy[runtime]) {
    return
  }
  busy[runtime] = true
  activeJob.value = null
  try {
    const job = await installRuntime(runtime, version)
    await fetchAll(true)
    await trackInstall(runtime, job)
  } catch (error) {
    busy[runtime] = false
    const detail = error instanceof ApiError && error.data?.message
      ? String(error.data.message)
      : error instanceof Error ? error.message : 'Install failed'
    toast.add({ title: 'Install failed', description: detail, color: 'error' })
    await fetchAll(true)
  }
}

async function onSetDefault(runtime: RuntimeName, version: string) {
  if (!canManage.value || busy[runtime]) {
    return
  }
  const family = runtime === 'php' ? server.value?.runtimes.php : server.value?.runtimes.node
  if (family?.default === version) {
    return
  }
  busy[runtime] = true
  activeJob.value = null
  try {
    const job = await setDefault(runtime, version)
    const done = await trackJob(runtime, job)
    toast.add({
      title: done?.status === 'succeeded' ? 'Default updated' : 'Set default failed',
      description: done?.status === 'succeeded'
        ? `${runtime.toUpperCase()} ${version} is now the CLI default on the server`
        : (done?.message || done?.status),
      color: done?.status === 'succeeded' ? 'success' : 'error'
    })
  } catch (error) {
    busy[runtime] = false
    const detail = error instanceof ApiError && error.data?.message
      ? String(error.data.message)
      : error instanceof Error ? error.message : 'Update failed'
    toast.add({ title: 'Update failed', description: detail, color: 'error' })
    await fetchAll(true)
  }
}

async function onRetry(runtime: RuntimeName, version: string) {
  await onInstall(runtime, version)
}
</script>

<template>
  <div
    v-if="server"
    class="flex flex-col gap-6"
  >
    <div class="flex flex-col gap-1">
      <h2 class="text-highlighted text-lg font-semibold">
        Runtimes
      </h2>
      <p class="text-muted text-sm">
        Installed PHP and Node versions, defaults for new sites, and provisioned add-ons.
      </p>
    </div>

    <UAlert
      v-if="server.agent !== 'online'"
      color="warning"
      variant="subtle"
      icon="i-lucide-wifi-off"
      title="Agent not online"
      description="Install and default changes need a live agent on this server."
    />

    <div
      v-if="activeJob && (activeJob.status === 'pending' || activeJob.status === 'running' || activeJob.status === 'failed')"
      class="border-default bg-elevated/40 rounded-lg border px-4 py-3"
    >
      <p class="text-highlighted text-xs font-medium">
        Runtime job
      </p>
      <pre class="text-muted mt-1 max-h-40 overflow-auto whitespace-pre-wrap font-mono text-xs">{{ jobSummary(activeJob) }}</pre>
    </div>

    <div class="grid gap-4 lg:grid-cols-2">
      <section class="border-default flex flex-col gap-4 rounded-lg border px-4 py-4 sm:px-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <div class="flex items-center gap-2">
              <UIcon
                name="i-lucide-file-code-2"
                class="text-muted size-4"
              />
              <h3 class="text-highlighted text-sm font-semibold">
                PHP
              </h3>
            </div>
            <p class="text-muted mt-1 text-xs">
              Default for new sites:
              <span class="text-highlighted font-medium">{{ server.runtimes.php.default || 'none' }}</span>
            </p>
          </div>
          <UBadge
            :label="`${server.runtimes.php.versions.filter(v => v.status === 'installed').length} installed`"
            color="neutral"
            variant="subtle"
            size="sm"
          />
        </div>

        <ul
          v-if="server.runtimes.php.versions.length"
          class="divide-default border-default divide-y rounded-md border"
        >
          <li
            v-for="ver in sortVersions(server.runtimes.php.versions)"
            :key="`php-${ver.version}`"
            class="flex items-center justify-between gap-3 px-3 py-2.5"
          >
            <div class="flex min-w-0 items-center gap-2">
              <span class="text-highlighted font-mono text-sm font-medium">
                {{ ver.version }}
              </span>
              <UBadge
                v-if="ver.version === server.runtimes.php.default"
                label="default"
                color="primary"
                variant="subtle"
                size="sm"
              />
              <UBadge
                :label="ver.status"
                :color="runtimeStatusColor(ver.status)"
                variant="subtle"
                size="sm"
                class="capitalize"
              />
            </div>
            <div class="flex shrink-0 items-center gap-1.5">
              <UButton
                v-if="ver.status === 'failed' && canManage"
                label="Retry"
                size="xs"
                color="neutral"
                variant="outline"
                :loading="busy.php"
                @click="onRetry('php', ver.version)"
              />
              <UButton
                v-else-if="ver.status === 'installed' && canManage && ver.version !== server.runtimes.php.default"
                label="Set default"
                size="xs"
                color="neutral"
                variant="ghost"
                :disabled="busy.php"
                @click="onSetDefault('php', ver.version)"
              />
              <UIcon
                v-else-if="ver.status === 'installing'"
                name="i-lucide-loader-circle"
                class="text-muted size-4 animate-spin"
              />
            </div>
          </li>
        </ul>
        <p
          v-else
          class="text-muted text-sm"
        >
          No PHP versions installed yet.
        </p>

        <div
          v-if="canManage && phpSelectItems.length"
          class="border-default flex flex-wrap items-end gap-2 border-t pt-4"
        >
          <UFormField
            label="Add version"
            class="min-w-28 flex-1"
          >
            <USelect
              v-model="phpToAdd"
              :items="phpSelectItems"
              value-key="value"
              class="w-full"
              size="sm"
              :disabled="busy.php"
            />
          </UFormField>
          <UButton
            label="Install"
            icon="i-lucide-download"
            size="sm"
            :loading="busy.php"
            :disabled="!phpToAdd"
            @click="onInstall('php', phpToAdd)"
          />
        </div>
        <p
          v-else-if="canManage && !phpSelectItems.length"
          class="text-muted border-default border-t pt-4 text-xs"
        >
          All catalog PHP versions are installed.
        </p>
      </section>

      <section class="border-default flex flex-col gap-4 rounded-lg border px-4 py-4 sm:px-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <div class="flex items-center gap-2">
              <UIcon
                name="i-lucide-hexagon"
                class="text-muted size-4"
              />
              <h3 class="text-highlighted text-sm font-semibold">
                Node
              </h3>
            </div>
            <p class="text-muted mt-1 text-xs">
              Default for new sites:
              <span class="text-highlighted font-medium">{{ server.runtimes.node.default || 'none' }}</span>
            </p>
          </div>
          <UBadge
            :label="`${server.runtimes.node.versions.filter(v => v.status === 'installed').length} installed`"
            color="neutral"
            variant="subtle"
            size="sm"
          />
        </div>

        <ul
          v-if="server.runtimes.node.versions.length"
          class="divide-default border-default divide-y rounded-md border"
        >
          <li
            v-for="ver in sortVersions(server.runtimes.node.versions)"
            :key="`node-${ver.version}`"
            class="flex items-center justify-between gap-3 px-3 py-2.5"
          >
            <div class="flex min-w-0 items-center gap-2">
              <span class="text-highlighted font-mono text-sm font-medium">
                {{ ver.version }}
              </span>
              <UBadge
                v-if="ver.version === server.runtimes.node.default"
                label="default"
                color="primary"
                variant="subtle"
                size="sm"
              />
              <UBadge
                :label="ver.status"
                :color="runtimeStatusColor(ver.status)"
                variant="subtle"
                size="sm"
                class="capitalize"
              />
            </div>
            <div class="flex shrink-0 items-center gap-1.5">
              <UButton
                v-if="ver.status === 'failed' && canManage"
                label="Retry"
                size="xs"
                color="neutral"
                variant="outline"
                :loading="busy.node"
                @click="onRetry('node', ver.version)"
              />
              <UButton
                v-else-if="ver.status === 'installed' && canManage && ver.version !== server.runtimes.node.default"
                label="Set default"
                size="xs"
                color="neutral"
                variant="ghost"
                :disabled="busy.node"
                @click="onSetDefault('node', ver.version)"
              />
              <UIcon
                v-else-if="ver.status === 'installing'"
                name="i-lucide-loader-circle"
                class="text-muted size-4 animate-spin"
              />
            </div>
          </li>
        </ul>
        <p
          v-else
          class="text-muted text-sm"
        >
          No Node versions installed yet.
        </p>

        <div
          v-if="canManage && nodeSelectItems.length"
          class="border-default flex flex-wrap items-end gap-2 border-t pt-4"
        >
          <UFormField
            label="Add version"
            class="min-w-28 flex-1"
          >
            <USelect
              v-model="nodeToAdd"
              :items="nodeSelectItems"
              value-key="value"
              class="w-full"
              size="sm"
              :disabled="busy.node"
            />
          </UFormField>
          <UButton
            label="Install"
            icon="i-lucide-download"
            size="sm"
            :loading="busy.node"
            :disabled="!nodeToAdd"
            @click="onInstall('node', nodeToAdd)"
          />
        </div>
        <p
          v-else-if="canManage && !nodeSelectItems.length"
          class="text-muted border-default border-t pt-4 text-xs"
        >
          All catalog Node versions are installed.
        </p>
      </section>
    </div>

    <section class="border-default rounded-lg border px-4 py-4 sm:px-5">
      <div class="mb-3 flex items-center gap-2">
        <UIcon
          name="i-lucide-database"
          class="text-muted size-4"
        />
        <h3 class="text-highlighted text-sm font-semibold">
          Add-ons
        </h3>
      </div>
      <div
        v-if="addonItems.length"
        class="flex flex-wrap gap-2"
      >
        <div
          v-for="addon in addonItems"
          :key="addon.key"
          class="border-default bg-elevated/40 inline-flex items-center gap-2 rounded-md border px-3 py-1.5"
        >
          <span class="bg-success size-1.5 rounded-full" />
          <span class="text-highlighted text-sm">
            {{ addon.label }}
          </span>
        </div>
      </div>
      <p
        v-else
        class="text-muted text-sm"
      >
        No database or cache add-ons from the last provision.
      </p>
    </section>
  </div>
</template>
