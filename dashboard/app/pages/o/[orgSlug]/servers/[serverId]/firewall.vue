<script setup lang="ts">
import { ApiError } from '~/utils/api'
import { Ability } from '~/utils/auth/abilities'
import type { AgentJob } from '~/utils/runtimes'
import {
  firewallStatusColor,
  formatFirewallPorts,
  looksLikeCIDR,
  type FirewallAction,
  type FirewallProtocol,
  type ServerFirewallRule
} from '~/utils/firewall'

const route = useRoute()
const toast = useToast()
const { can } = useGate()
const { getById, fetchAll: fetchServers } = useServers()
const serverId = computed(() => String(route.params.serverId || ''))

const {
  rules,
  fetchAll,
  create,
  update,
  remove,
  sync,
  pollJob
} = useServerFirewall(serverId)

await fetchServers()
const server = computed(() => getById(serverId.value))
const canManage = computed(() => can(Ability.FirewallManage) && server.value?.agent === 'online')

const busy = ref(false)
const formOpen = ref(false)
const editing = ref<ServerFirewallRule | null>(null)
const deleteOpen = ref(false)
const toDelete = ref<ServerFirewallRule | null>(null)

const actionItems = [
  { label: 'Allow', value: 'allow' },
  { label: 'Deny', value: 'deny' }
]
const protocolItems = [
  { label: 'TCP', value: 'tcp' },
  { label: 'UDP', value: 'udp' },
  { label: 'ICMP', value: 'icmp' },
  { label: 'Any', value: 'any' }
]

const form = reactive({
  name: '',
  action: 'allow' as FirewallAction,
  protocol: 'tcp' as FirewallProtocol,
  portStart: '80',
  portEnd: '',
  source: '0.0.0.0/0',
  enabled: true
})

const needsPorts = computed(() => form.protocol === 'tcp' || form.protocol === 'udp')

try {
  await fetchAll()
} catch (error) {
  const detail = error instanceof ApiError ? error.message : 'Failed to load firewall rules'
  toast.add({ title: detail, color: 'error' })
}

function resetForm() {
  form.name = ''
  form.action = 'allow'
  form.protocol = 'tcp'
  form.portStart = '80'
  form.portEnd = ''
  form.source = '0.0.0.0/0'
  form.enabled = true
}

function openCreate() {
  editing.value = null
  resetForm()
  formOpen.value = true
}

function openEdit(rule: ServerFirewallRule) {
  editing.value = rule
  form.name = rule.name
  form.action = rule.action
  form.protocol = rule.protocol
  form.portStart = rule.portStart ? String(rule.portStart) : ''
  form.portEnd = rule.portEnd && rule.portEnd !== rule.portStart ? String(rule.portEnd) : ''
  form.source = rule.source
  form.enabled = rule.enabled
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

function parsePorts(): { portStart?: number, portEnd?: number } | null {
  if (!needsPorts.value) {
    return { portStart: 0, portEnd: 0 }
  }
  const startRaw = form.portStart.trim()
  const endRaw = form.portEnd.trim()
  if (!startRaw && !endRaw) {
    return { portStart: 0, portEnd: 0 }
  }
  const start = Number(startRaw)
  if (!Number.isInteger(start) || start < 1 || start > 65535) {
    return null
  }
  if (!endRaw) {
    return { portStart: start, portEnd: start }
  }
  const end = Number(endRaw)
  if (!Number.isInteger(end) || end < start || end > 65535) {
    return null
  }
  return { portStart: start, portEnd: end }
}

async function onSubmit() {
  if (!canManage.value || busy.value) {
    return
  }
  if (!looksLikeCIDR(form.source)) {
    toast.add({ title: 'Invalid source', description: 'Use an IPv4 address or CIDR', color: 'error' })
    return
  }
  const ports = parsePorts()
  if (!ports) {
    toast.add({ title: 'Invalid port range', color: 'error' })
    return
  }
  busy.value = true
  try {
    const payload = {
      name: form.name.trim(),
      action: form.action,
      protocol: form.protocol,
      portStart: ports.portStart,
      portEnd: ports.portEnd,
      source: form.source.trim() || '0.0.0.0/0',
      enabled: form.enabled
    }
    const res = editing.value
      ? await update(editing.value.id, payload)
      : await create(payload)
    formOpen.value = false
    await fetchAll()
    const done = await track(res.job)
    toast.add({
      title: done?.status === 'succeeded'
        ? (editing.value ? 'Rule updated' : 'Rule created')
        : 'Firewall sync failed',
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

function confirmDelete(rule: ServerFirewallRule) {
  toDelete.value = rule
  deleteOpen.value = true
}

async function onDelete() {
  if (!toDelete.value || !canManage.value || busy.value) {
    return
  }
  const rule = toDelete.value
  deleteOpen.value = false
  busy.value = true
  try {
    const res = await remove(rule.id)
    await fetchAll()
    const done = await track(res.job)
    toast.add({
      title: done?.status === 'succeeded' ? 'Rule removed' : 'Remove failed',
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
      title: done?.status === 'succeeded' ? 'Firewall synced' : 'Firewall sync failed',
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

function ruleTitle(rule: ServerFirewallRule) {
  if (rule.name) {
    return rule.name
  }
  const ports = formatFirewallPorts(rule)
  return `${rule.action} ${rule.protocol}${ports !== '-' ? ` ${ports}` : ''}`
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
          Firewall
        </h2>
        <p class="text-muted text-sm">
          Inbound rules applied to iptables chain <span class="font-mono">SA-FW</span>.
          SSH port {{ server.sshPort || 22 }}/tcp stays allowed.
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
          label="Add rule"
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
      description="Firewall changes need a live agent on this server."
    />

    <UAlert
      color="neutral"
      variant="subtle"
      icon="i-lucide-info"
      title="How rules apply"
      description="SSH stays allowed. Deny rules win over open connections. Allow rules apply next. Unmatched traffic falls through to the rest of the host firewall."
    />

    <template v-if="!rules.length">
      <section class="border-default rounded-lg border px-5 py-10 sm:px-8">
        <div class="flex max-w-xl flex-col gap-4">
          <div class="bg-primary/10 text-primary flex size-10 items-center justify-center rounded-lg">
            <UIcon
              name="i-lucide-shield"
              class="size-5"
            />
          </div>
          <div class="flex flex-col gap-2">
            <h3 class="text-highlighted text-lg font-semibold">
              No firewall rules yet
            </h3>
            <p class="text-muted text-sm leading-relaxed">
              Add allow or deny rules for inbound TCP, UDP, or ICMP traffic from specific sources.
            </p>
          </div>
          <UButton
            label="Add rule"
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
        v-for="rule in rules"
        :key="rule.id"
        class="border-default rounded-lg border px-4 py-3"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-highlighted text-sm font-semibold">
                {{ ruleTitle(rule) }}
              </span>
              <UBadge
                :label="rule.action"
                :color="rule.action === 'allow' ? 'success' : 'error'"
                variant="subtle"
                size="sm"
                class="capitalize"
              />
              <UBadge
                :label="rule.protocol"
                color="neutral"
                variant="subtle"
                size="sm"
                class="uppercase"
              />
              <UBadge
                :label="rule.status"
                :color="firewallStatusColor(rule.status)"
                variant="subtle"
                size="sm"
                class="capitalize"
              />
              <UBadge
                v-if="!rule.enabled"
                label="disabled"
                color="neutral"
                variant="subtle"
                size="sm"
              />
            </div>
            <p class="text-muted mt-1 font-mono text-xs">
              {{ formatFirewallPorts(rule) }} from {{ rule.source }}
            </p>
            <p
              v-if="rule.message"
              class="text-muted mt-1 text-xs"
            >
              {{ rule.message }}
            </p>
          </div>
          <div class="flex shrink-0 items-center gap-1">
            <UButton
              icon="i-lucide-pencil"
              size="xs"
              color="neutral"
              variant="ghost"
              :disabled="!canManage || busy"
              @click="openEdit(rule)"
            />
            <UButton
              icon="i-lucide-trash-2"
              size="xs"
              color="error"
              variant="ghost"
              :disabled="!canManage || busy"
              @click="confirmDelete(rule)"
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
              {{ editing ? 'Edit rule' : 'Add rule' }}
            </p>
          </template>
          <form
            class="flex flex-col gap-3"
            @submit.prevent="onSubmit"
          >
            <UFormField
              label="Name"
              hint="Optional label"
            >
              <UInput
                v-model="form.name"
                placeholder="HTTP"
                class="w-full"
              />
            </UFormField>
            <div class="grid gap-3 sm:grid-cols-2">
              <UFormField label="Action">
                <USelect
                  v-model="form.action"
                  :items="actionItems"
                  value-key="value"
                  class="w-full"
                />
              </UFormField>
              <UFormField label="Protocol">
                <USelect
                  v-model="form.protocol"
                  :items="protocolItems"
                  value-key="value"
                  class="w-full"
                />
              </UFormField>
            </div>
            <div
              v-if="needsPorts"
              class="grid gap-3 sm:grid-cols-2"
            >
              <UFormField
                label="Port"
                hint="Leave empty for any"
              >
                <UInput
                  v-model="form.portStart"
                  type="number"
                  min="1"
                  max="65535"
                  placeholder="80"
                  class="w-full"
                />
              </UFormField>
              <UFormField
                label="Port end"
                hint="Optional range"
              >
                <UInput
                  v-model="form.portEnd"
                  type="number"
                  min="1"
                  max="65535"
                  placeholder="Same as start"
                  class="w-full"
                />
              </UFormField>
            </div>
            <UFormField
              label="Source"
              hint="IPv4 or CIDR"
            >
              <UInput
                v-model="form.source"
                placeholder="0.0.0.0/0"
                class="w-full font-mono"
              />
            </UFormField>
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
              Remove rule
            </p>
          </template>
          <p class="text-muted text-sm">
            Remove <span class="text-highlighted font-semibold">{{ toDelete ? ruleTitle(toDelete) : '' }}</span>
            ({{ toDelete?.action }} {{ toDelete?.protocol }} {{ toDelete ? formatFirewallPorts(toDelete) : '' }} from {{ toDelete?.source }})?
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
