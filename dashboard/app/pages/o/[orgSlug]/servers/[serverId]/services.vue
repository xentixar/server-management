<script setup lang="ts">
import { ApiError } from '~/utils/api'
import { Ability } from '~/utils/auth/abilities'
import type { AgentJob } from '~/utils/runtimes'
import {
  isValidServiceName,
  serviceStatusColor,
  type ServerService,
  type ServiceControlAction
} from '~/utils/processes'

const route = useRoute()
const toast = useToast()
const { can } = useGate()
const { getById, fetchAll: fetchServers } = useServers()
const serverId = computed(() => String(route.params.serverId || ''))

const {
  services,
  fetchAll,
  create,
  update,
  remove,
  control,
  pollJob
} = useServerServices(serverId)

await fetchServers()
const server = computed(() => getById(serverId.value))
const canManage = computed(() => can(Ability.ServersUpdate) && server.value?.agent === 'online')

const busy = ref(false)
const activeJob = ref<AgentJob | null>(null)
const expanded = ref<string | null>(null)
const formOpen = ref(false)
const editing = ref<ServerService | null>(null)
const deleteOpen = ref(false)
const toDelete = ref<ServerService | null>(null)

const form = reactive({
  name: '',
  command: '',
  directory: '',
  user: '',
  processes: 1,
  autoRestart: true,
  autostart: true,
  enabled: true
})

try {
  await fetchAll()
} catch (error) {
  const detail = error instanceof ApiError ? error.message : 'Failed to load services'
  toast.add({ title: detail, color: 'error' })
}

function resetForm() {
  form.name = ''
  form.command = ''
  form.directory = ''
  form.user = server.value?.managedUser || 'sm'
  form.processes = 1
  form.autoRestart = true
  form.autostart = true
  form.enabled = true
}

function openCreate() {
  editing.value = null
  resetForm()
  formOpen.value = true
}

function openEdit(svc: ServerService) {
  editing.value = svc
  form.name = svc.name
  form.command = svc.command
  form.directory = svc.directory
  form.user = svc.user
  form.processes = svc.processes
  form.autoRestart = svc.autoRestart
  form.autostart = svc.autostart
  form.enabled = svc.enabled
  formOpen.value = true
}

async function track(job: AgentJob) {
  busy.value = true
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
    busy.value = false
  }
}

async function onSubmit() {
  if (!canManage.value || busy.value) {
    return
  }
  if (!editing.value && !isValidServiceName(form.name.trim().toLowerCase())) {
    toast.add({ title: 'Invalid name', description: 'Use lowercase letters, numbers, _ or -', color: 'error' })
    return
  }
  if (!form.command.trim()) {
    toast.add({ title: 'Command required', color: 'error' })
    return
  }
  busy.value = true
  try {
    const res = editing.value
      ? await update(editing.value.id, {
          command: form.command.trim(),
          directory: form.directory.trim(),
          user: form.user.trim() || server.value?.managedUser || 'sm',
          processes: form.processes,
          autoRestart: form.autoRestart,
          autostart: form.autostart,
          enabled: form.enabled
        })
      : await create({
          name: form.name.trim().toLowerCase(),
          command: form.command.trim(),
          directory: form.directory.trim(),
          user: form.user.trim() || server.value?.managedUser || 'sm',
          processes: form.processes,
          autoRestart: form.autoRestart,
          autostart: form.autostart,
          enabled: form.enabled
        })
    formOpen.value = false
    await fetchAll()
    const done = await track(res.job)
    toast.add({
      title: done?.status === 'succeeded' ? (editing.value ? 'Service updated' : 'Service created') : 'Service job failed',
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

async function onControl(svc: ServerService, action: ServiceControlAction) {
  if (!canManage.value || busy.value) {
    return
  }
  busy.value = true
  try {
    const res = await control(svc.id, action)
    await fetchAll()
    const done = await track(res.job)
    toast.add({
      title: done?.status === 'succeeded' ? `${action} ok` : `${action} failed`,
      description: done?.message || done?.status,
      color: done?.status === 'succeeded' ? 'success' : 'error'
    })
    if (done?.status === 'succeeded') {
      expanded.value = svc.id
    }
  } catch (error) {
    busy.value = false
    const detail = error instanceof ApiError ? error.message : 'Control failed'
    toast.add({ title: detail, color: 'error' })
    await fetchAll()
  }
}

async function toggleLogs(svc: ServerService) {
  if (expanded.value === svc.id) {
    expanded.value = null
    return
  }
  expanded.value = svc.id
  if (canManage.value) {
    await onControl(svc, 'status')
  }
}

function confirmDelete(svc: ServerService) {
  toDelete.value = svc
  deleteOpen.value = true
}

async function onDelete() {
  if (!toDelete.value || !canManage.value || busy.value) {
    return
  }
  const svc = toDelete.value
  deleteOpen.value = false
  busy.value = true
  try {
    const res = await remove(svc.id)
    await fetchAll()
    const done = await track(res.job)
    toast.add({
      title: done?.status === 'succeeded' ? 'Service removed' : 'Remove failed',
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
</script>

<template>
  <div
    v-if="server"
    class="flex flex-col gap-6"
  >
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h2 class="text-highlighted text-lg font-semibold">
          Services
        </h2>
        <p class="text-muted text-sm">
          Supervisor-backed background processes on this server.
        </p>
      </div>
      <UButton
        label="Add service"
        icon="i-lucide-plus"
        :disabled="!canManage || busy"
        @click="openCreate"
      />
    </div>

    <UAlert
      v-if="server.agent !== 'online'"
      color="warning"
      variant="subtle"
      icon="i-lucide-wifi-off"
      title="Agent not online"
      description="Service changes need a live agent on this server."
    />

    <div
      v-if="activeJob && (activeJob.status === 'pending' || activeJob.status === 'running' || activeJob.status === 'failed')"
      class="border-default bg-elevated/40 rounded-lg border px-4 py-3"
    >
      <p class="text-highlighted text-xs font-medium">
        Service job
      </p>
      <pre class="text-muted mt-1 max-h-40 overflow-auto whitespace-pre-wrap font-mono text-xs">{{ activeJob.kind }}: {{ activeJob.message || activeJob.status }}</pre>
    </div>

    <template v-if="!services.length">
      <section class="border-default rounded-lg border px-5 py-10 sm:px-8">
        <div class="flex max-w-xl flex-col gap-4">
          <div class="bg-primary/10 text-primary flex size-10 items-center justify-center rounded-lg">
            <UIcon
              name="i-lucide-cog"
              class="size-5"
            />
          </div>
          <div class="flex flex-col gap-2">
            <h3 class="text-highlighted text-lg font-semibold">
              No services yet
            </h3>
            <p class="text-muted text-sm leading-relaxed">
              Add a queue worker or long-running command. The agent installs Supervisor and keeps the process managed.
            </p>
          </div>
          <UButton
            label="Add service"
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
      class="border-default divide-default divide-y overflow-hidden rounded-lg border"
    >
      <li
        v-for="svc in services"
        :key="svc.id"
        class="flex flex-col gap-3 px-4 py-3.5 sm:px-5"
      >
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-highlighted font-mono text-sm font-semibold">
                {{ svc.name }}
              </span>
              <UBadge
                :label="svc.status"
                :color="serviceStatusColor(svc.status)"
                variant="subtle"
                size="sm"
                class="capitalize"
              />
              <UBadge
                v-if="!svc.enabled"
                label="disabled"
                color="neutral"
                variant="subtle"
                size="sm"
              />
              <UBadge
                v-if="svc.autoRestart"
                label="auto-restart"
                color="neutral"
                variant="subtle"
                size="sm"
              />
            </div>
            <p class="text-muted mt-1 truncate font-mono text-xs">
              {{ svc.command }}
            </p>
            <p class="text-muted mt-0.5 text-xs">
              user {{ svc.user }} · {{ svc.processes }} process{{ svc.processes === 1 ? '' : 'es' }}
              <span v-if="svc.directory"> · {{ svc.directory }}</span>
            </p>
          </div>
          <div class="flex flex-wrap items-center gap-1.5">
            <UButton
              label="Start"
              size="xs"
              color="neutral"
              variant="outline"
              :disabled="!canManage || busy"
              @click="onControl(svc, 'start')"
            />
            <UButton
              label="Stop"
              size="xs"
              color="neutral"
              variant="outline"
              :disabled="!canManage || busy"
              @click="onControl(svc, 'stop')"
            />
            <UButton
              label="Restart"
              size="xs"
              color="neutral"
              variant="outline"
              :disabled="!canManage || busy"
              @click="onControl(svc, 'restart')"
            />
            <UButton
              :label="expanded === svc.id ? 'Hide logs' : 'Logs'"
              size="xs"
              color="neutral"
              variant="ghost"
              :disabled="busy"
              @click="toggleLogs(svc)"
            />
            <UButton
              icon="i-lucide-pencil"
              size="xs"
              color="neutral"
              variant="ghost"
              :disabled="!canManage || busy"
              @click="openEdit(svc)"
            />
            <UButton
              icon="i-lucide-trash-2"
              size="xs"
              color="error"
              variant="ghost"
              :disabled="!canManage || busy"
              @click="confirmDelete(svc)"
            />
          </div>
        </div>
        <div
          v-if="expanded === svc.id"
          class="grid gap-2 sm:grid-cols-2"
        >
          <div class="border-default bg-elevated/30 rounded-md border p-2">
            <p class="text-muted mb-1 text-xs font-medium">
              stdout
            </p>
            <pre class="text-muted max-h-40 overflow-auto whitespace-pre-wrap font-mono text-[11px]">{{ svc.stdoutTail || '(empty)' }}</pre>
          </div>
          <div class="border-default bg-elevated/30 rounded-md border p-2">
            <p class="text-muted mb-1 text-xs font-medium">
              stderr
            </p>
            <pre class="text-muted max-h-40 overflow-auto whitespace-pre-wrap font-mono text-[11px]">{{ svc.stderrTail || '(empty)' }}</pre>
          </div>
          <p
            v-if="svc.message"
            class="text-muted sm:col-span-2 text-xs"
          >
            {{ svc.message }}
          </p>
        </div>
      </li>
    </ul>

    <UModal v-model:open="formOpen">
      <template #content>
        <UCard>
          <template #header>
            <p class="text-highlighted font-semibold">
              {{ editing ? 'Edit service' : 'Add service' }}
            </p>
          </template>
          <form
            class="flex flex-col gap-3"
            @submit.prevent="onSubmit"
          >
            <UFormField
              label="Name"
              hint="lowercase slug"
            >
              <UInput
                v-model="form.name"
                :disabled="!!editing"
                placeholder="queue-worker"
                class="w-full"
              />
            </UFormField>
            <UFormField label="Command">
              <UInput
                v-model="form.command"
                placeholder="php /var/www/app/artisan queue:work"
                class="w-full"
              />
            </UFormField>
            <UFormField label="Directory">
              <UInput
                v-model="form.directory"
                placeholder="/var/www/app"
                class="w-full"
              />
            </UFormField>
            <div class="grid gap-3 sm:grid-cols-2">
              <UFormField label="User">
                <UInput
                  v-model="form.user"
                  class="w-full"
                />
              </UFormField>
              <UFormField label="Processes">
                <UInput
                  v-model.number="form.processes"
                  type="number"
                  :min="1"
                  :max="64"
                  class="w-full"
                />
              </UFormField>
            </div>
            <div class="flex flex-wrap gap-4">
              <UCheckbox
                v-model="form.autoRestart"
                label="Auto-restart"
              />
              <UCheckbox
                v-model="form.autostart"
                label="Autostart"
              />
              <UCheckbox
                v-model="form.enabled"
                label="Enabled"
              />
            </div>
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
              Remove service
            </p>
          </template>
          <p class="text-muted text-sm">
            Remove <span class="text-highlighted font-mono">{{ toDelete?.name }}</span> from Supervisor on this host?
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
