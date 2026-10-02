<script setup lang="ts">
import { ApiError } from '~/utils/api'
import { Ability } from '~/utils/auth/abilities'
import {
  DEFAULT_NODE_VERSIONS,
  DEFAULT_PHP_VERSIONS,
  type RuntimeCatalog,
  type RuntimeName
} from '~/utils/runtimes'

const route = useRoute()
const toast = useToast()
const { can } = useGate()
const { orgHref } = useOrganizations()
const serverId = computed(() => String(route.params.serverId || ''))

const { getById, fetchAll } = useServers()
const {
  getCatalog,
  installRuntime,
  setDefault,
  pollJob
} = useServerRuntimes(serverId)

await fetchAll()

const server = computed(() => getById(serverId.value))
const catalog = ref<RuntimeCatalog>({ php: DEFAULT_PHP_VERSIONS, node: DEFAULT_NODE_VERSIONS })
const busy = ref(false)
const phpToAdd = ref('')
const nodeToAdd = ref('')

const phpAvailable = computed(() => {
  const installed = new Set(server.value?.runtimes.php.versions.map(v => v.version) ?? [])
  return catalog.value.php.filter(v => !installed.has(v))
})
const nodeAvailable = computed(() => {
  const installed = new Set(server.value?.runtimes.node.versions.map(v => v.version) ?? [])
  return catalog.value.node.filter(v => !installed.has(v))
})

onMounted(async () => {
  try {
    catalog.value = await getCatalog()
    if (!phpToAdd.value && phpAvailable.value[0]) {
      phpToAdd.value = phpAvailable.value[0]
    }
    if (!nodeToAdd.value && nodeAvailable.value[0]) {
      nodeToAdd.value = nodeAvailable.value[0]
    }
  } catch {
    // defaults
  }
})

async function onInstall(runtime: RuntimeName, version: string) {
  if (!version || !can(Ability.ServersUpdate)) {
    return
  }
  busy.value = true
  try {
    const job = await installRuntime(runtime, version)
    const done = await pollJob(job.id)
    toast.add({
      title: done?.status === 'succeeded' ? `${runtime} ${version} installed` : 'Install failed',
      description: done?.message || done?.status,
      color: done?.status === 'succeeded' ? 'success' : 'error'
    })
  } catch (error) {
    const detail = error instanceof ApiError && error.data?.message
      ? String(error.data.message)
      : error instanceof Error ? error.message : 'Install failed'
    toast.add({ title: 'Install failed', description: detail, color: 'error' })
  } finally {
    busy.value = false
    await fetchAll(true)
  }
}

async function onSetDefault(runtime: RuntimeName, version: string) {
  if (!can(Ability.ServersUpdate)) {
    return
  }
  busy.value = true
  try {
    await setDefault(runtime, version)
    await fetchAll(true)
    toast.add({
      title: 'Default updated',
      description: `${runtime} default is now ${version}`,
      color: 'success'
    })
  } catch (error) {
    const detail = error instanceof ApiError && error.data?.message
      ? String(error.data.message)
      : error instanceof Error ? error.message : 'Update failed'
    toast.add({ title: 'Update failed', description: detail, color: 'error' })
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div
    v-if="server"
    class="flex flex-col gap-6"
  >
    <UAlert
      v-if="server.provisionStatus !== 'ready'"
      color="warning"
      variant="subtle"
      icon="i-lucide-package"
      title="Stack not provisioned"
      description="Provision the server stack from Overview before managing runtimes."
    >
      <template #actions>
        <UButton
          label="Go to Overview"
          size="sm"
          color="neutral"
          variant="outline"
          :to="orgHref(`/servers/${serverId}`)"
        />
      </template>
    </UAlert>

    <template v-else>
      <section class="border-default rounded-lg border px-4 py-4 sm:px-5">
        <h2 class="text-highlighted mb-1 text-sm font-semibold">
          PHP
        </h2>
        <p class="text-muted mb-3 text-xs">
          Default for new sites: {{ server.runtimes.php.default || 'none' }}
        </p>
        <div class="mb-4 flex flex-wrap gap-2">
          <UBadge
            v-for="ver in server.runtimes.php.versions"
            :key="`php-${ver.version}`"
            :label="`${ver.version}${ver.version === server.runtimes.php.default ? ' (default)' : ''} · ${ver.status}`"
            :color="ver.status === 'installed' ? 'success' : ver.status === 'failed' ? 'error' : 'warning'"
            variant="subtle"
            class="cursor-pointer"
            @click="ver.status === 'installed' && onSetDefault('php', ver.version)"
          />
          <p
            v-if="!server.runtimes.php.versions.length"
            class="text-muted text-sm"
          >
            No PHP versions installed.
          </p>
        </div>
        <div
          v-if="can(Ability.ServersUpdate) && phpAvailable.length"
          class="flex flex-wrap items-center gap-2"
        >
          <USelect
            v-model="phpToAdd"
            :items="phpAvailable"
            class="w-32"
            size="sm"
          />
          <UButton
            label="Install PHP"
            size="sm"
            :loading="busy"
            @click="onInstall('php', phpToAdd)"
          />
        </div>
      </section>

      <section class="border-default rounded-lg border px-4 py-4 sm:px-5">
        <h2 class="text-highlighted mb-1 text-sm font-semibold">
          Node
        </h2>
        <p class="text-muted mb-3 text-xs">
          Default for new sites: {{ server.runtimes.node.default || 'none' }}
        </p>
        <div class="mb-4 flex flex-wrap gap-2">
          <UBadge
            v-for="ver in server.runtimes.node.versions"
            :key="`node-${ver.version}`"
            :label="`${ver.version}${ver.version === server.runtimes.node.default ? ' (default)' : ''} · ${ver.status}`"
            :color="ver.status === 'installed' ? 'success' : ver.status === 'failed' ? 'error' : 'warning'"
            variant="subtle"
            class="cursor-pointer"
            @click="ver.status === 'installed' && onSetDefault('node', ver.version)"
          />
          <p
            v-if="!server.runtimes.node.versions.length"
            class="text-muted text-sm"
          >
            No Node versions installed.
          </p>
        </div>
        <div
          v-if="can(Ability.ServersUpdate) && nodeAvailable.length"
          class="flex flex-wrap items-center gap-2"
        >
          <USelect
            v-model="nodeToAdd"
            :items="nodeAvailable"
            class="w-32"
            size="sm"
          />
          <UButton
            label="Install Node"
            size="sm"
            :loading="busy"
            @click="onInstall('node', nodeToAdd)"
          />
        </div>
      </section>

      <section
        v-if="server.runtimes.addons.length"
        class="border-default rounded-lg border px-4 py-4 sm:px-5"
      >
        <h2 class="text-highlighted mb-3 text-sm font-semibold">
          Add-ons
        </h2>
        <div class="flex flex-wrap gap-2">
          <UBadge
            v-for="addon in server.runtimes.addons"
            :key="addon"
            :label="addon"
            color="neutral"
            variant="subtle"
            class="capitalize"
          />
        </div>
      </section>
    </template>
  </div>
</template>
