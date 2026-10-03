<script setup lang="ts">
import { ApiError } from '~/utils/api'
import { Ability } from '~/utils/auth/abilities'
import type { AgentJob } from '~/utils/runtimes'
import {
  looksLikePublicKey,
  sshKeyStatusColor,
  truncatePublicKey,
  type ServerSshKey
} from '~/utils/ssh-keys'

const route = useRoute()
const toast = useToast()
const { can } = useGate()
const { getById, fetchAll: fetchServers } = useServers()
const serverId = computed(() => String(route.params.serverId || ''))

const {
  keys,
  fetchAll,
  create,
  update,
  remove,
  sync,
  pollJob
} = useServerSshKeys(serverId)

await fetchServers()
const server = computed(() => getById(serverId.value))
const canManage = computed(() => can(Ability.SshKeysManage) && server.value?.agent === 'online')
const managedUser = computed(() => server.value?.managedUser || 'sm')

const busy = ref(false)
const formOpen = ref(false)
const editing = ref<ServerSshKey | null>(null)
const deleteOpen = ref(false)
const toDelete = ref<ServerSshKey | null>(null)

const form = reactive({
  name: '',
  publicKey: ''
})

try {
  await fetchAll()
} catch (error) {
  const detail = error instanceof ApiError ? error.message : 'Failed to load SSH keys'
  toast.add({ title: detail, color: 'error' })
}

function resetForm() {
  form.name = ''
  form.publicKey = ''
}

function openCreate() {
  editing.value = null
  resetForm()
  formOpen.value = true
}

function openEdit(key: ServerSshKey) {
  editing.value = key
  form.name = key.name
  form.publicKey = key.publicKey
  formOpen.value = true
}

async function track(job: AgentJob) {
  busy.value = true
  try {
    return await pollJob(job.id)
  } finally {
    busy.value = false
  }
}

async function onSubmit() {
  if (!canManage.value || busy.value) {
    return
  }
  const name = form.name.trim()
  const publicKey = form.publicKey.trim()
  if (!/^[a-z0-9][a-z0-9_-]{0,62}$/.test(name)) {
    toast.add({
      title: 'Invalid name',
      description: 'Use lowercase letters, numbers, hyphens, or underscores.',
      color: 'error'
    })
    return
  }
  if (!looksLikePublicKey(publicKey)) {
    toast.add({
      title: 'Invalid public key',
      description: 'Paste a single OpenSSH public key line.',
      color: 'error'
    })
    return
  }
  busy.value = true
  try {
    const res = editing.value
      ? await update(editing.value.id, { name, publicKey })
      : await create({ name, publicKey })
    formOpen.value = false
    await fetchAll()
    const done = await track(res.job)
    toast.add({
      title: done?.status === 'succeeded'
        ? (editing.value ? 'SSH key updated' : 'SSH key created')
        : 'SSH key sync failed',
      description: done?.message || done?.status,
      color: done?.status === 'succeeded' ? 'success' : 'error'
    })
  } catch (error) {
    busy.value = false
    const detail = error instanceof ApiError ? error.message : 'Request failed'
    toast.add({ title: detail, color: 'error' })
    await fetchAll()
  }
}

function confirmDelete(key: ServerSshKey) {
  toDelete.value = key
  deleteOpen.value = true
}

async function onDelete() {
  if (!toDelete.value || !canManage.value || busy.value) {
    return
  }
  const key = toDelete.value
  deleteOpen.value = false
  busy.value = true
  try {
    const res = await remove(key.id)
    await fetchAll()
    const done = await track(res.job)
    toast.add({
      title: done?.status === 'succeeded' ? 'SSH key removed' : 'Remove failed',
      description: done?.message || done?.status,
      color: done?.status === 'succeeded' ? 'success' : 'error'
    })
  } catch (error) {
    busy.value = false
    const detail = error instanceof ApiError ? error.message : 'Delete failed'
    toast.add({ title: detail, color: 'error' })
    await fetchAll()
  } finally {
    toDelete.value = null
  }
}

async function onSync() {
  if (!canManage.value || busy.value) {
    return
  }
  busy.value = true
  try {
    const res = await sync()
    await fetchAll()
    const done = await track(res.job)
    toast.add({
      title: done?.status === 'succeeded' ? 'SSH keys synced' : 'SSH key sync failed',
      description: done?.message || done?.status,
      color: done?.status === 'succeeded' ? 'success' : 'error'
    })
  } catch (error) {
    busy.value = false
    const detail = error instanceof ApiError ? error.message : 'Sync failed'
    toast.add({ title: detail, color: 'error' })
    await fetchAll()
  }
}
</script>

<template>
  <div
    v-if="server"
    class="flex flex-col gap-6"
  >
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h2 class="text-highlighted text-lg font-semibold">
          SSH keys
        </h2>
        <p class="text-muted text-sm">
          Public keys for the managed user <span class="font-mono">{{ managedUser }}</span>.
          Connect with <span class="font-mono">ssh {{ managedUser }}@{{ server.ip }}</span>.
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UButton
          label="Sync"
          icon="i-lucide-refresh-cw"
          color="neutral"
          variant="soft"
          :disabled="!canManage || busy"
          @click="onSync"
        />
        <UButton
          label="Add key"
          icon="i-lucide-plus"
          :disabled="!canManage || busy"
          @click="openCreate"
        />
      </div>
    </div>

    <UAlert
      v-if="server.agent !== 'online'"
      color="warning"
      variant="subtle"
      icon="i-lucide-wifi-off"
      title="Agent not online"
      description="SSH key changes need a live agent on this server."
    />

    <UAlert
      color="neutral"
      variant="subtle"
      icon="i-lucide-info"
      title="How keys apply"
      :description="`Keys rewrite /home/${managedUser}/.ssh/authorized_keys on sync. Manual edits to that file are overwritten.`"
    />

    <template v-if="!keys.length">
      <section class="border-default rounded-lg border px-5 py-10 sm:px-8">
        <div class="flex max-w-xl flex-col gap-4">
          <div class="bg-primary/10 text-primary flex size-10 items-center justify-center rounded-lg">
            <UIcon
              name="i-lucide-key-round"
              class="size-5"
            />
          </div>
          <div class="flex flex-col gap-2">
            <h3 class="text-highlighted text-lg font-semibold">
              No SSH keys yet
            </h3>
            <p class="text-muted text-sm leading-relaxed">
              Add your OpenSSH public key so you can sign in as {{ managedUser }} from outside.
            </p>
          </div>
          <UButton
            label="Add key"
            icon="i-lucide-plus"
            class="w-fit"
            :disabled="!canManage || busy"
            @click="openCreate"
          />
        </div>
      </section>
    </template>

    <ul
      v-else
      class="flex flex-col gap-3"
    >
      <li
        v-for="key in keys"
        :key="key.id"
        class="border-default rounded-lg border px-4 py-3"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-highlighted text-sm font-semibold">
                {{ key.name }}
              </span>
              <UBadge
                :label="key.status"
                :color="sshKeyStatusColor(key.status)"
                variant="subtle"
                size="sm"
                class="capitalize"
              />
            </div>
            <p class="text-muted mt-1 font-mono text-xs break-all">
              {{ key.fingerprint }}
            </p>
            <p class="text-muted mt-1 font-mono text-xs break-all">
              {{ truncatePublicKey(key.publicKey) }}
            </p>
            <p
              v-if="key.message"
              class="text-muted mt-1 text-xs"
            >
              {{ key.message }}
            </p>
          </div>
          <div class="flex shrink-0 items-center gap-1">
            <UButton
              icon="i-lucide-pencil"
              size="xs"
              color="neutral"
              variant="ghost"
              :disabled="!canManage || busy"
              @click="openEdit(key)"
            />
            <UButton
              icon="i-lucide-trash-2"
              size="xs"
              color="error"
              variant="ghost"
              :disabled="!canManage || busy"
              @click="confirmDelete(key)"
            />
          </div>
        </div>
      </li>
    </ul>

    <UModal v-model:open="formOpen">
      <template #content>
        <UCard>
          <template #header>
            <p class="text-highlighted font-semibold">
              {{ editing ? 'Edit SSH key' : 'Add SSH key' }}
            </p>
          </template>
          <form
            class="flex flex-col gap-3"
            @submit.prevent="onSubmit"
          >
            <UFormField
              label="Name"
              hint="Lowercase label, e.g. laptop"
            >
              <UInput
                v-model="form.name"
                placeholder="laptop"
                class="w-full"
              />
            </UFormField>
            <UFormField
              label="Public key"
              hint="One OpenSSH public key line"
            >
              <UTextarea
                v-model="form.publicKey"
                :rows="4"
                autoresize
                placeholder="ssh-ed25519 AAAA... comment"
                class="w-full font-mono text-xs"
              />
            </UFormField>
            <div class="flex justify-end gap-2 pt-2">
              <UButton
                label="Cancel"
                color="neutral"
                variant="ghost"
                @click="formOpen = false"
              />
              <UButton
                type="submit"
                :label="editing ? 'Save' : 'Create'"
                :loading="busy"
                :disabled="!canManage"
              />
            </div>
          </form>
        </UCard>
      </template>
    </UModal>

    <UModal v-model:open="deleteOpen">
      <template #content>
        <UCard>
          <template #header>
            <p class="text-highlighted font-semibold">
              Remove SSH key
            </p>
          </template>
          <p class="text-muted text-sm">
            Remove <span class="text-highlighted font-semibold">{{ toDelete?.name }}</span>
            ({{ toDelete?.fingerprint }}) from {{ managedUser }}?
          </p>
          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton
                label="Cancel"
                color="neutral"
                variant="ghost"
                @click="deleteOpen = false"
              />
              <UButton
                label="Remove"
                color="error"
                :loading="busy"
                @click="onDelete"
              />
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
  </div>
</template>
