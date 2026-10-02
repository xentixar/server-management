<script setup lang="ts">
import { ApiError } from '~/utils/api'
import { Ability } from '~/utils/auth/abilities'
import type { AgentJob } from '~/utils/runtimes'
import {
  cronStatusColor,
  looksLikeCronSchedule,
  type ServerCron
} from '~/utils/processes'

const route = useRoute()
const toast = useToast()
const { can } = useGate()
const { getById, fetchAll: fetchServers } = useServers()
const serverId = computed(() => String(route.params.serverId || ''))

const {
  crons,
  fetchAll,
  create,
  update,
  remove,
  pollJob
} = useServerCrons(serverId)

await fetchServers()
const server = computed(() => getById(serverId.value))
const canManage = computed(() => can(Ability.ServersUpdate) && server.value?.agent === 'online')

const busy = ref(false)
const activeJob = ref<AgentJob | null>(null)
const formOpen = ref(false)
const editing = ref<ServerCron | null>(null)
const deleteOpen = ref(false)
const toDelete = ref<ServerCron | null>(null)

const form = reactive({
  schedule: '*/5 * * * *',
  command: '',
  user: '',
  timezone: 'UTC',
  enabled: true
})

try {
  await fetchAll()
} catch (error) {
  const detail = error instanceof ApiError ? error.message : 'Failed to load cron jobs'
  toast.add({ title: detail, color: 'error' })
}

function resetForm() {
  form.schedule = '*/5 * * * *'
  form.command = ''
  form.user = server.value?.managedUser || 'sm'
  form.timezone = 'UTC'
  form.enabled = true
}

function openCreate() {
  editing.value = null
  resetForm()
  formOpen.value = true
}

function openEdit(cron: ServerCron) {
  editing.value = cron
  form.schedule = cron.schedule
  form.command = cron.command
  form.user = cron.user
  form.timezone = cron.timezone
  form.enabled = cron.enabled
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
  if (!looksLikeCronSchedule(form.schedule)) {
    toast.add({
      title: 'Invalid schedule',
      description: 'Use a 5-field cron expression (e.g. */5 * * * *)',
      color: 'error'
    })
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
          schedule: form.schedule.trim(),
          command: form.command.trim(),
          user: form.user.trim() || server.value?.managedUser || 'sm',
          timezone: form.timezone.trim() || 'UTC',
          enabled: form.enabled
        })
      : await create({
          schedule: form.schedule.trim(),
          command: form.command.trim(),
          user: form.user.trim() || server.value?.managedUser || 'sm',
          timezone: form.timezone.trim() || 'UTC',
          enabled: form.enabled
        })
    formOpen.value = false
    await fetchAll()
    const done = await track(res.job)
    toast.add({
      title: done?.status === 'succeeded' ? (editing.value ? 'Cron updated' : 'Cron created') : 'Cron sync failed',
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

function confirmDelete(cron: ServerCron) {
  toDelete.value = cron
  deleteOpen.value = true
}

async function onDelete() {
  if (!toDelete.value || !canManage.value || busy.value) {
    return
  }
  const cron = toDelete.value
  deleteOpen.value = false
  busy.value = true
  try {
    const res = await remove(cron.id)
    await fetchAll()
    const done = await track(res.job)
    toast.add({
      title: done?.status === 'succeeded' ? 'Cron removed' : 'Remove failed',
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
          Cron
        </h2>
        <p class="text-muted text-sm">
          Server-wide scheduled jobs written to <span class="font-mono">/etc/cron.d/sa</span>.
        </p>
      </div>
      <UButton
        label="Add cron"
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
      description="Cron changes need a live agent on this server."
    />

    <div
      v-if="activeJob && (activeJob.status === 'pending' || activeJob.status === 'running' || activeJob.status === 'failed')"
      class="border-default bg-elevated/40 rounded-lg border px-4 py-3"
    >
      <p class="text-highlighted text-xs font-medium">
        Cron job
      </p>
      <pre class="text-muted mt-1 max-h-40 overflow-auto whitespace-pre-wrap font-mono text-xs">{{ activeJob.kind }}: {{ activeJob.message || activeJob.status }}</pre>
    </div>

    <template v-if="!crons.length">
      <section class="border-default rounded-lg border px-5 py-10 sm:px-8">
        <div class="flex max-w-xl flex-col gap-4">
          <div class="bg-primary/10 text-primary flex size-10 items-center justify-center rounded-lg">
            <UIcon
              name="i-lucide-clock"
              class="size-5"
            />
          </div>
          <div class="flex flex-col gap-2">
            <h3 class="text-highlighted text-lg font-semibold">
              No cron jobs yet
            </h3>
            <p class="text-muted text-sm leading-relaxed">
              Schedule recurring commands with a standard 5-field cron expression and timezone.
            </p>
          </div>
          <UButton
            label="Add cron"
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
        v-for="cron in crons"
        :key="cron.id"
        class="flex flex-wrap items-start justify-between gap-3 px-4 py-3.5 sm:px-5"
      >
        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-highlighted font-mono text-sm font-semibold">
              {{ cron.schedule }}
            </span>
            <UBadge
              :label="cron.status"
              :color="cronStatusColor(cron.status)"
              variant="subtle"
              size="sm"
              class="capitalize"
            />
            <UBadge
              v-if="!cron.enabled"
              label="disabled"
              color="neutral"
              variant="subtle"
              size="sm"
            />
            <UBadge
              :label="cron.timezone"
              color="neutral"
              variant="subtle"
              size="sm"
            />
          </div>
          <p class="text-muted mt-1 truncate font-mono text-xs">
            {{ cron.command }}
          </p>
          <p class="text-muted mt-0.5 text-xs">
            user {{ cron.user }}
            <span v-if="cron.message"> · {{ cron.message }}</span>
          </p>
        </div>
        <div class="flex shrink-0 items-center gap-1.5">
          <UButton
            icon="i-lucide-pencil"
            size="xs"
            color="neutral"
            variant="ghost"
            :disabled="!canManage || busy"
            @click="openEdit(cron)"
          />
          <UButton
            icon="i-lucide-trash-2"
            size="xs"
            color="error"
            variant="ghost"
            :disabled="!canManage || busy"
            @click="confirmDelete(cron)"
          />
        </div>
      </li>
    </ul>

    <UModal v-model:open="formOpen">
      <template #content>
        <UCard>
          <template #header>
            <p class="text-highlighted font-semibold">
              {{ editing ? 'Edit cron' : 'Add cron' }}
            </p>
          </template>
          <form
            class="flex flex-col gap-3"
            @submit.prevent="onSubmit"
          >
            <UFormField
              label="Schedule"
              hint="minute hour day month weekday"
            >
              <UInput
                v-model="form.schedule"
                placeholder="*/5 * * * *"
                class="w-full font-mono"
              />
            </UFormField>
            <UFormField label="Command">
              <UInput
                v-model="form.command"
                placeholder="php /var/www/app/artisan schedule:run"
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
              <UFormField label="Timezone">
                <UInput
                  v-model="form.timezone"
                  placeholder="UTC"
                  class="w-full"
                />
              </UFormField>
            </div>
            <UCheckbox
              v-model="form.enabled"
              label="Enabled"
            />
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
              Remove cron
            </p>
          </template>
          <p class="text-muted text-sm">
            Remove <span class="text-highlighted font-mono">{{ toDelete?.schedule }}</span>
            (<span class="font-mono"> {{ toDelete?.command }}</span>?
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
