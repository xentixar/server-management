<script setup lang="ts">
import type { ProvisionLogLine, ProvisionStep } from '~/composables/useServerProvision'

const props = defineProps<{
  steps: ProvisionStep[]
  logs: ProvisionLogLine[]
  status: 'idle' | 'running' | 'done' | 'failed'
  progress: number
  autoStart?: boolean
}>()

const emit = defineEmits<{
  start: []
  replay: []
}>()

const logEl = ref<HTMLElement | null>(null)

watch(() => props.logs.length, async () => {
  await nextTick()
  if (logEl.value) {
    logEl.value.scrollTop = logEl.value.scrollHeight
  }
})

onMounted(() => {
  if (props.autoStart && props.status === 'idle') {
    emit('start')
  }
})

function stepIcon(status: ProvisionStep['status']) {
  if (status === 'done') {
    return 'i-lucide-check'
  }
  if (status === 'running') {
    return 'i-lucide-loader-circle'
  }
  if (status === 'failed') {
    return 'i-lucide-x'
  }
  return 'i-lucide-circle'
}

function formatLogTime(at: number) {
  return new Date(at).toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}
</script>

<template>
  <div class="border-default overflow-hidden rounded-lg border">
    <div class="border-default flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3 sm:px-5">
      <div class="min-w-0">
        <p class="text-highlighted text-sm font-semibold">
          Agent provisioning
        </p>
        <p class="text-muted text-xs">
          <template v-if="status === 'running'">Installing and registering the agent...</template>
          <template v-else-if="status === 'done'">Agent is online.</template>
          <template v-else-if="status === 'failed'">Provisioning failed.</template>
          <template v-else>Ready to install the agent on this server.</template>
        </p>
      </div>
      <div class="flex items-center gap-2">
        <UBadge
          :label="status"
          :color="status === 'done' ? 'success' : status === 'failed' ? 'error' : status === 'running' ? 'primary' : 'neutral'"
          variant="subtle"
          class="capitalize"
        />
        <UButton
          v-if="status === 'idle'"
          label="Start"
          size="sm"
          @click="emit('start')"
        />
        <UButton
          v-else-if="status === 'done' || status === 'failed'"
          label="Reinstall agent"
          color="neutral"
          variant="outline"
          size="sm"
          @click="emit('replay')"
        />
      </div>
    </div>

    <div class="border-default border-b px-4 py-3 sm:px-5">
      <div class="mb-2 flex items-center justify-between text-xs">
        <span class="text-muted">Progress</span>
        <span class="text-highlighted tabular-nums font-medium">{{ progress }}%</span>
      </div>
      <UProgress
        :model-value="progress"
        size="sm"
        :color="status === 'failed' ? 'error' : 'primary'"
      />
    </div>

    <div class="grid lg:grid-cols-5">
      <ol class="border-default flex flex-col gap-0 border-b p-4 lg:col-span-2 lg:border-b-0 lg:border-r sm:p-5">
        <li
          v-for="(step, index) in steps"
          :key="step.id"
          class="relative flex gap-3 py-2.5"
          :class="step.status === 'pending' ? 'opacity-50' : 'opacity-100'"
        >
          <div class="relative z-10 flex flex-col items-center">
            <div
              class="flex size-7 items-center justify-center rounded-full border text-xs transition-colors duration-300"
              :class="{
                'border-primary bg-primary text-inverted': step.status === 'running',
                'border-success bg-success text-inverted': step.status === 'done',
                'border-error bg-error text-inverted': step.status === 'failed',
                'border-default bg-default text-muted': step.status === 'pending'
              }"
            >
              <UIcon
                :name="stepIcon(step.status)"
                class="size-3.5"
                :class="step.status === 'running' ? 'animate-spin' : ''"
              />
            </div>
            <div
              v-if="index < steps.length - 1"
              class="bg-default mt-1 w-px flex-1"
              :class="step.status === 'done' ? 'bg-success/40' : 'bg-default'"
            />
          </div>
          <div class="min-w-0 pt-1">
            <p
              class="text-sm font-medium transition-colors duration-300"
              :class="step.status === 'pending' ? 'text-muted' : 'text-highlighted'"
            >
              {{ step.label }}
            </p>
            <p class="text-muted text-xs capitalize">
              {{ step.status }}
            </p>
          </div>
        </li>
      </ol>

      <div class="flex flex-col lg:col-span-3">
        <div class="border-default text-muted border-b px-4 py-2 text-xs font-medium tracking-wide uppercase">
          Live log
        </div>
        <div
          ref="logEl"
          class="h-64 overflow-y-auto bg-zinc-950 px-4 py-3 font-mono text-sm leading-relaxed text-zinc-100 sm:h-80 dark:bg-black"
        >
          <p
            v-if="!logs.length"
            class="text-zinc-500"
          >
            Waiting for output...
          </p>
          <div
            v-for="line in logs"
            :key="line.id"
            class="log-line flex gap-3 py-0.5"
          >
            <span class="shrink-0 tabular-nums text-zinc-500">{{ formatLogTime(line.at) }}</span>
            <span
              class="min-w-0 break-all"
              :class="{
                'text-amber-300': line.stream === 'stderr',
                'text-sky-300': line.stream === 'system',
                'text-emerald-300': line.stream === 'stdout'
              }"
            >{{ line.text }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.log-line {
  animation: log-in 280ms ease-out;
}

@keyframes log-in {
  from {
    opacity: 0;
    transform: translateY(4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
