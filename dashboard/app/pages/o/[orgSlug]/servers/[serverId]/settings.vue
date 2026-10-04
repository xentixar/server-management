<script setup lang="ts">
import { ApiError } from '~/utils/api'
import { Ability } from '~/utils/auth/abilities'
import {
  looksLikeHostHostname,
  type SshAuthMethod
} from '~/utils/servers'

const route = useRoute()
const toast = useToast()
const { can } = useGate()
const { orgHref } = useOrganizations()
const serverId = computed(() => String(route.params.serverId || ''))

const {
  getById,
  fetchAll,
  update,
  testServerConnection,
  bootstrap,
  pollJob,
  remove
} = useServers()

await fetchAll()
const server = computed(() => getById(serverId.value))

const canUpdate = computed(() => can(Ability.ServersUpdate))
const canDelete = computed(() => can(Ability.ServersDelete))

const busy = ref(false)
const testing = ref(false)
const bootstrapping = ref(false)
const deleting = ref(false)
const deleteOpen = ref(false)
const showReinstallHint = ref(false)

const privateKey = ref('')
const keyFileName = ref('')
const keyInput = ref<HTMLInputElement>()

const form = reactive({
  hostname: '',
  ip: '',
  sshPort: 22,
  sshUser: 'root',
  authMethod: 'key' as SshAuthMethod,
  passphrase: '',
  password: '',
  region: '',
  timezone: 'UTC',
  labels: '',
  notes: ''
})

const initialHost = ref({ hostname: '', timezone: 'UTC' })

const authItems = [{
  label: 'SSH private key',
  value: 'key' as SshAuthMethod
}, {
  label: 'Password',
  value: 'password' as SshAuthMethod
}]

function syncForm() {
  const s = server.value
  if (!s) {
    return
  }
  form.hostname = s.hostname
  form.ip = s.ip
  form.sshPort = s.sshPort || 22
  form.sshUser = s.sshUser
  form.authMethod = s.authMethod
  form.passphrase = ''
  form.password = ''
  form.region = s.region || ''
  form.timezone = s.timezone || 'UTC'
  form.labels = s.labels.join(', ')
  form.notes = s.notes || ''
  initialHost.value = { hostname: s.hostname, timezone: s.timezone || 'UTC' }
  privateKey.value = ''
  keyFileName.value = ''
}

syncForm()
watch(server, () => {
  if (!busy.value) {
    syncForm()
  }
})

function parseLabels(raw: string): string[] {
  return raw
    .split(',')
    .map(l => l.trim())
    .filter(Boolean)
}

function buildInput() {
  return {
    hostname: form.hostname.trim(),
    ip: form.ip.trim(),
    sshPort: Number(form.sshPort) || 22,
    sshUser: form.sshUser.trim() || 'root',
    authMethod: form.authMethod,
    privateKey: privateKey.value,
    passphrase: form.passphrase,
    password: form.password,
    region: form.region.trim(),
    timezone: form.timezone.trim() || 'UTC',
    labels: parseLabels(form.labels),
    notes: form.notes.trim()
  }
}

async function onKeyFile(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    keyFileName.value = ''
    privateKey.value = ''
    return
  }
  keyFileName.value = file.name
  privateKey.value = await file.text()
}

async function onTest() {
  if (!canUpdate.value || testing.value || busy.value || !server.value) {
    return
  }
  if (!form.ip.trim() || !form.sshUser.trim()) {
    toast.add({ title: 'IP and SSH user are required', color: 'error' })
    return
  }
  testing.value = true
  try {
    const result = await testServerConnection(server.value.id, buildInput())
    toast.add({
      title: 'SSH connection ok',
      description: result.message,
      color: 'success'
    })
  } catch (error) {
    let detail = error instanceof ApiError ? error.message : 'SSH failed'
    if (error instanceof ApiError && error.data) {
      const field = error.data.private_key || error.data.password || error.data.message
      if (field) {
        detail = field
      }
    }
    toast.add({ title: 'SSH connection failed', description: detail, color: 'error' })
  } finally {
    testing.value = false
  }
}

async function onSave() {
  if (!canUpdate.value || busy.value || !server.value) {
    return
  }
  const hostname = form.hostname.trim()
  const timezone = form.timezone.trim() || 'UTC'
  const hostChanged = hostname !== initialHost.value.hostname || timezone !== initialHost.value.timezone
  if (hostChanged && !looksLikeHostHostname(hostname)) {
    toast.add({
      title: 'Invalid hostname',
      description: 'Use a DNS-safe name (letters, numbers, hyphens, dots).',
      color: 'error'
    })
    return
  }
  if (hostChanged && server.value.agent !== 'online') {
    toast.add({
      title: 'Agent offline',
      description: 'Hostname and timezone need an online agent to apply on the host.',
      color: 'error'
    })
    return
  }
  if (!form.ip.trim() || !form.sshUser.trim()) {
    toast.add({ title: 'IP and SSH user are required', color: 'error' })
    return
  }

  busy.value = true
  showReinstallHint.value = false
  try {
    const res = await update(server.value.id, buildInput())
    syncForm()
    if (res.job) {
      const done = await pollJob(server.value.id, res.job.id)
      toast.add({
        title: done?.status === 'succeeded' ? 'Settings applied' : 'Settings sync failed',
        description: done?.message || done?.status,
        color: done?.status === 'succeeded' ? 'success' : 'error'
      })
    } else {
      toast.add({ title: 'Server updated', color: 'success' })
      showReinstallHint.value = true
    }
  } catch (error) {
    const detail = error instanceof ApiError ? error.message : 'Update failed'
    toast.add({ title: detail, color: 'error' })
    await fetchAll(true)
    syncForm()
  } finally {
    busy.value = false
  }
}

async function onReinstall() {
  if (!canUpdate.value || !server.value || bootstrapping.value) {
    return
  }
  bootstrapping.value = true
  try {
    const result = await bootstrap(server.value.id)
    toast.add({
      title: 'Agent reinstall started',
      description: result.message,
      color: 'success'
    })
    await fetchAll(true)
  } catch (error) {
    const detail = error instanceof ApiError ? error.message : 'Bootstrap failed'
    toast.add({ title: 'Agent install failed', description: detail, color: 'error' })
  } finally {
    bootstrapping.value = false
  }
}

async function onConfirmDelete() {
  if (!server.value || !canDelete.value || deleting.value) {
    return
  }
  deleting.value = true
  try {
    await remove(server.value.id)
    toast.add({ title: 'Server deleted', color: 'success' })
    await navigateTo(orgHref('/servers'), { replace: true })
  } catch (error) {
    const detail = error instanceof ApiError ? error.message : 'Delete failed'
    toast.add({ title: detail, color: 'error' })
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div
    v-if="server"
    class="flex flex-col gap-6"
  >
    <div>
      <h2 class="text-highlighted text-lg font-semibold">
        Settings
      </h2>
      <p class="text-muted text-sm">
        Connection details for the panel, plus host hostname and timezone when the agent is online.
      </p>
    </div>

    <section class="border-default flex flex-col gap-4 rounded-lg border p-4 sm:p-5">
      <div>
        <h3 class="text-highlighted text-sm font-semibold">
          General
        </h3>
        <p class="text-muted text-xs">
          Hostname is the panel name and is applied on the host when it changes (agent online).
        </p>
      </div>
      <div class="grid items-start gap-4 sm:grid-cols-2">
        <UFormField
          label="Hostname"
          required
          class="sm:col-span-2"
        >
          <UInput
            v-model="form.hostname"
            :disabled="!canUpdate || busy"
            class="w-full"
          />
        </UFormField>
        <UFormField
          label="Timezone"
          description="IANA name, e.g. UTC"
        >
          <UInput
            v-model="form.timezone"
            placeholder="UTC"
            :disabled="!canUpdate || busy"
            class="w-full font-mono"
          />
        </UFormField>
        <UFormField
          label="Region"
          description="Optional label for sorting"
        >
          <UInput
            v-model="form.region"
            placeholder="nyc1"
            :disabled="!canUpdate || busy"
            class="w-full"
          />
        </UFormField>
        <UFormField
          label="Labels"
          description="Comma-separated"
          class="sm:col-span-2"
        >
          <UInput
            v-model="form.labels"
            placeholder="prod, edge"
            :disabled="!canUpdate || busy"
            class="w-full"
          />
        </UFormField>
        <UFormField
          label="Notes"
          class="sm:col-span-2"
        >
          <UTextarea
            v-model="form.notes"
            :rows="3"
            autoresize
            :disabled="!canUpdate || busy"
            class="w-full"
          />
        </UFormField>
      </div>
    </section>

    <section class="border-default flex flex-col gap-4 rounded-lg border p-4 sm:p-5">
      <div>
        <h3 class="text-highlighted text-sm font-semibold">
          Connection
        </h3>
        <p class="text-muted text-xs">
          Used by the panel for bootstrap, agent reinstall, and delete cleanup. Does not change sshd on the host.
          Leave key/password blank to keep the current secret.
        </p>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <UFormField
          label="IP or host"
          required
          class="sm:col-span-2"
        >
          <UInput
            v-model="form.ip"
            :disabled="!canUpdate || busy"
            class="w-full"
          />
        </UFormField>
        <UFormField
          label="SSH port"
          required
        >
          <UInput
            v-model.number="form.sshPort"
            type="number"
            min="1"
            max="65535"
            :disabled="!canUpdate || busy"
            class="w-full"
          />
        </UFormField>
        <UFormField
          label="SSH user"
          required
        >
          <UInput
            v-model="form.sshUser"
            :disabled="!canUpdate || busy"
            class="w-full"
          />
        </UFormField>
        <UFormField
          label="Auth method"
          class="sm:col-span-2"
        >
          <USelect
            v-model="form.authMethod"
            :items="authItems"
            value-key="value"
            :disabled="!canUpdate || busy"
            class="w-full"
          />
        </UFormField>
        <template v-if="form.authMethod === 'key'">
          <UFormField
            label="Private key"
            :description="server.hasKeyUploaded ? 'A key is stored. Choose a file only to replace it.' : 'No key stored yet.'"
            class="sm:col-span-2"
          >
            <div class="flex flex-wrap items-center gap-3">
              <UButton
                label="Choose file"
                color="neutral"
                :disabled="!canUpdate || busy"
                @click="keyInput?.click()"
              />
              <span class="text-muted text-sm">
                {{ keyFileName || 'No file chosen' }}
              </span>
              <input
                ref="keyInput"
                type="file"
                class="hidden"
                @change="onKeyFile"
              >
            </div>
          </UFormField>
          <UFormField
            label="Passphrase"
            description="Optional. Leave blank to keep the current passphrase."
            class="sm:col-span-2"
          >
            <UInput
              v-model="form.passphrase"
              type="password"
              autocomplete="off"
              :disabled="!canUpdate || busy"
              class="w-full"
            />
          </UFormField>
        </template>
        <template v-else>
          <UAlert
            class="sm:col-span-2"
            color="warning"
            variant="subtle"
            icon="i-lucide-triangle-alert"
            title="Password auth is discouraged"
            description="Prefer key-based SSH. Leave blank to keep the current password."
          />
          <UFormField
            label="Password"
            class="sm:col-span-2"
          >
            <UInput
              v-model="form.password"
              type="password"
              autocomplete="new-password"
              :disabled="!canUpdate || busy"
              class="w-full"
            />
          </UFormField>
        </template>
      </div>
    </section>

    <div class="flex flex-wrap gap-2">
      <UButton
        label="Test connection"
        color="neutral"
        variant="soft"
        icon="i-lucide-plug"
        :loading="testing"
        :disabled="!canUpdate || busy"
        @click="onTest"
      />
      <UButton
        label="Save"
        icon="i-lucide-save"
        :loading="busy"
        :disabled="!canUpdate || testing"
        @click="onSave"
      />
    </div>

    <UAlert
      v-if="showReinstallHint"
      color="neutral"
      variant="subtle"
      icon="i-lucide-info"
      title="Connection saved"
      description="If this is a new or wiped host, reinstall the agent. Re-provision from Overview only if the stack is missing."
    >
      <template #actions>
        <UButton
          label="Reinstall agent"
          size="sm"
          color="neutral"
          variant="soft"
          :loading="bootstrapping"
          :disabled="!canUpdate"
          @click="onReinstall"
        />
      </template>
    </UAlert>

    <section class="border-error/40 flex flex-col gap-3 rounded-lg border p-4 sm:p-5">
      <div>
        <h3 class="text-highlighted text-sm font-semibold">
          Danger zone
        </h3>
        <p class="text-muted text-xs">
          Removes this server from the organization and uninstalls the agent over SSH.
        </p>
      </div>
      <UButton
        label="Delete server"
        color="error"
        variant="soft"
        icon="i-lucide-trash-2"
        class="w-fit"
        :disabled="!canDelete || busy || deleting"
        @click="deleteOpen = true"
      />
    </section>

    <OrganizationDeleteServerModal
      v-model:open="deleteOpen"
      :server="server"
      @confirm="onConfirmDelete"
    />
  </div>
</template>
