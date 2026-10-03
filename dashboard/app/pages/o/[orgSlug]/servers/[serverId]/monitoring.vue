<script setup lang="ts">
import { ApiError } from '~/utils/api'
import { Ability } from '~/utils/auth/abilities'
import {
  formatBytes,
  formatLoad,
  formatPercent,
  METRICS_RANGES,
  usagePercent,
  type MetricChartSeries,
  type MetricPoint,
  type MetricsRange
} from '~/utils/metrics'

const route = useRoute()
const toast = useToast()
const { can } = useGate()
const { getById, fetchAll: fetchServers } = useServers()
const serverId = computed(() => String(route.params.serverId || ''))

const { metrics, range, loading, fetchMetrics } = useServerMetrics(serverId)

await fetchServers()
const server = computed(() => getById(serverId.value))
const canView = computed(() => can(Ability.ServersView))

try {
  if (canView.value) {
    await fetchMetrics()
  }
} catch (error) {
  const detail = error instanceof ApiError ? error.message : 'Failed to load metrics'
  toast.add({ title: detail, color: 'error' })
}

const latest = computed(() => metrics.value?.latest ?? null)
const series = computed(() => metrics.value?.series ?? [])

const memPct = computed(() => usagePercent(latest.value?.memUsedBytes ?? null, latest.value?.memTotalBytes ?? null))
const diskPct = computed(() => usagePercent(latest.value?.diskUsedBytes ?? null, latest.value?.diskTotalBytes ?? null))

const stats = computed(() => [{
  label: 'CPU',
  value: formatPercent(latest.value?.cpuPercent),
  hint: 'Current utilization'
}, {
  label: 'Memory',
  value: formatPercent(memPct.value),
  hint: latest.value
    ? `${formatBytes(latest.value.memUsedBytes)} / ${formatBytes(latest.value.memTotalBytes)}`
    : 'No sample yet'
}, {
  label: 'Disk',
  value: formatPercent(diskPct.value),
  hint: latest.value
    ? `${formatBytes(latest.value.diskUsedBytes)} / ${formatBytes(latest.value.diskTotalBytes)}`
    : 'Root filesystem'
}, {
  label: 'Load',
  value: formatLoad(latest.value?.load1),
  hint: latest.value
    ? `5m ${formatLoad(latest.value.load5)} / 15m ${formatLoad(latest.value.load15)}`
    : '1 / 5 / 15 minute'
}])

type ChartRow = {
  t: number
  cpu: number | null
  mem: number | null
  disk: number | null
  load1: number | null
  load5: number | null
  load15: number | null
}

/** Carry forward last known values so early heartbeats without CPU still chart cleanly. */
const chartRows = computed<ChartRow[]>(() => {
  let cpu: number | null = null
  let mem: number | null = null
  let disk: number | null = null
  let load1: number | null = null
  let load5: number | null = null
  let load15: number | null = null
  const rows: ChartRow[] = []
  for (const p of series.value as MetricPoint[]) {
    if (p.cpuPercent != null) cpu = p.cpuPercent
    const nextMem = usagePercent(p.memUsedBytes, p.memTotalBytes)
    if (nextMem != null) mem = nextMem
    const nextDisk = usagePercent(p.diskUsedBytes, p.diskTotalBytes)
    if (nextDisk != null) disk = nextDisk
    if (p.load1 != null) load1 = p.load1
    if (p.load5 != null) load5 = p.load5
    if (p.load15 != null) load15 = p.load15
    rows.push({ t: p.t.getTime(), cpu, mem, disk, load1, load5, load15 })
  }
  return rows
})

/** Fit the plot to samples. Expand to the selected window only once data covers most of it. */
const xDomain = computed<[number, number]>(() => {
  const rows = chartRows.value
  const rangeFrom = metrics.value?.from?.getTime()
  const rangeTo = metrics.value?.to?.getTime()
  const rangeSpan = rangeFrom != null && rangeTo != null ? rangeTo - rangeFrom : 0

  if (!rows.length) {
    const now = Date.now()
    return [now - (rangeSpan || 3_600_000), now]
  }

  const dataMin = rows[0]!.t
  const dataMax = rows[rows.length - 1]!.t
  const dataSpan = Math.max(0, dataMax - dataMin)

  // Sparse history: zoom to samples so the line fills the chart (not a 1px spike on the right).
  if (!rangeSpan || dataSpan < rangeSpan * 0.35) {
    const pad = Math.max(30_000, dataSpan * 0.08, 60_000)
    const midPad = dataSpan === 0 ? 60_000 : pad
    return [dataMin - midPad, dataMax + midPad]
  }

  return [rangeFrom!, rangeTo!]
})

const xSpanMs = computed(() => xDomain.value[1] - xDomain.value[0])

const x = (d: ChartRow) => d.t

function xTickFormat(v: number | Date) {
  const d = typeof v === 'number' ? new Date(v) : v
  if (range.value === '7d' && xSpanMs.value >= 2 * 24 * 3_600_000) {
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
  }
  // Short windows need seconds so ticks are not all "11:06 AM"
  if (xSpanMs.value <= 15 * 60_000) {
    return d.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit', second: '2-digit' })
  }
  return d.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' })
}

function pctTick(v: number) {
  return `${Math.round(v)}%`
}

function tipTime(d: ChartRow) {
  return new Date(d.t).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    second: '2-digit'
  })
}

function tipBox(title: string, value: string) {
  // Inline styles: Unovis tooltip HTML sits outside Vue scoped/Tailwind class scope.
  return `<div style="font:12px/1.4 ui-sans-serif,system-ui,sans-serif;padding:6px 8px">
    <div style="opacity:.65;margin-bottom:2px">${title}</div>
    <div style="font-weight:600">${value}</div>
  </div>`
}

const colors = {
  cpu: '#10b981',
  mem: '#34d399',
  disk: '#f59e0b',
  load1: '#10b981',
  load5: '#38bdf8',
  load15: '#94a3b8'
}

const cpuSeries = computed<MetricChartSeries<ChartRow>[]>(() => [{
  key: 'cpu',
  label: 'CPU',
  color: colors.cpu,
  area: true,
  y: d => d.cpu
}])

const memSeries = computed<MetricChartSeries<ChartRow>[]>(() => [{
  key: 'mem',
  label: 'Memory',
  color: colors.mem,
  area: true,
  y: d => d.mem
}])

const diskSeries = computed<MetricChartSeries<ChartRow>[]>(() => [{
  key: 'disk',
  label: 'Disk',
  color: colors.disk,
  area: true,
  y: d => d.disk
}])

const loadSeries = computed<MetricChartSeries<ChartRow>[]>(() => [{
  key: 'load1',
  label: '1m',
  color: colors.load1,
  y: d => d.load1
}, {
  key: 'load5',
  label: '5m',
  color: colors.load5,
  y: d => d.load5
}, {
  key: 'load15',
  label: '15m',
  color: colors.load15,
  y: d => d.load15
}])

const loadYDomain = computed<[number, number]>(() => {
  const vals = chartRows.value.flatMap(r => [r.load1, r.load5, r.load15])
    .filter((v): v is number => v != null && Number.isFinite(v))
  const peak = vals.length ? Math.max(...vals) : 1
  return [0, Math.max(1, peak * 1.25)]
})

function tipCpu(d: ChartRow) {
  return tipBox(tipTime(d), `CPU ${formatPercent(d.cpu)}`)
}
function tipMem(d: ChartRow) {
  return tipBox(tipTime(d), `Memory ${formatPercent(d.mem)}`)
}
function tipDisk(d: ChartRow) {
  return tipBox(tipTime(d), `Disk ${formatPercent(d.disk)}`)
}
function tipLoad(d: ChartRow) {
  return tipBox(tipTime(d), `Load ${formatLoad(d.load1)} / ${formatLoad(d.load5)} / ${formatLoad(d.load15)}`)
}

/** Zoom Y when usage is low so a 0.7% CPU line is not glued to the floor. */
function pctYDomain(values: Array<number | null>): [number, number] {
  const nums = values.filter((v): v is number => v != null && Number.isFinite(v))
  const peak = nums.length ? Math.max(...nums) : 0
  if (peak >= 50) {
    return [0, 100]
  }
  const top = Math.min(100, Math.max(10, Math.ceil((peak * 1.4) / 5) * 5))
  return [0, top]
}

const cpuYDomain = computed(() => pctYDomain(chartRows.value.map(r => r.cpu)))
const memYDomain = computed(() => pctYDomain(chartRows.value.map(r => r.mem)))
const diskYDomain = computed(() => pctYDomain(chartRows.value.map(r => r.disk)))

async function setRange(next: MetricsRange) {
  if (next === range.value && metrics.value) {
    return
  }
  try {
    await fetchMetrics(next)
  } catch (error) {
    const detail = error instanceof ApiError ? error.message : 'Failed to load metrics'
    toast.add({ title: detail, color: 'error' })
  }
}

let timer: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  timer = setInterval(async () => {
    if (!canView.value || !import.meta.client) {
      return
    }
    try {
      await fetchMetrics(range.value)
    } catch {
      // keep last good sample
    }
  }, 30_000)
})
onBeforeUnmount(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
      <div class="flex flex-col gap-1">
        <h2 class="text-highlighted text-xl font-semibold tracking-tight">
          Monitoring
        </h2>
        <p class="text-muted max-w-2xl text-sm">
          Host CPU, memory, disk, and load sampled from the agent heartbeat.
        </p>
      </div>
      <UFieldGroup>
        <UButton
          v-for="item in METRICS_RANGES"
          :key="item.value"
          size="sm"
          :variant="range === item.value ? 'soft' : 'ghost'"
          :color="range === item.value ? 'primary' : 'neutral'"
          :label="item.label"
          :disabled="loading"
          @click="setRange(item.value)"
        />
      </UFieldGroup>
    </div>

    <UAlert
      v-if="server && server.agent !== 'online'"
      color="warning"
      variant="subtle"
      icon="i-lucide-wifi-off"
      title="Agent is not online"
      description="Charts show stored samples. New data arrives when the agent heartbeats again."
    />

    <section
      v-if="!latest && !loading"
      class="border-default flex flex-col items-center gap-3 rounded-lg border border-dashed px-6 py-12 text-center"
    >
      <UIcon
        name="i-lucide-activity"
        class="text-muted size-8"
      />
      <div class="flex flex-col gap-1">
        <p class="text-highlighted text-sm font-medium">
          No metrics yet
        </p>
        <p class="text-muted text-sm">
          Samples appear after the agent sends a heartbeat with host metrics. Redeploy the agent if this stays empty.
        </p>
      </div>
    </section>

    <template v-else>
      <section
        aria-label="Current gauges"
        class="border-default grid grid-cols-2 gap-px overflow-hidden rounded-lg border bg-elevated/50 sm:grid-cols-4"
      >
        <div
          v-for="stat in stats"
          :key="stat.label"
          class="bg-default flex flex-col gap-1 px-4 py-4 sm:px-5 sm:py-5"
        >
          <p class="text-muted text-xs font-medium tracking-wide uppercase">
            {{ stat.label }}
          </p>
          <p class="text-highlighted text-2xl font-semibold tracking-tight tabular-nums">
            {{ stat.value }}
          </p>
          <p class="text-muted text-xs">
            {{ stat.hint }}
          </p>
        </div>
      </section>

      <ClientOnly>
        <div
          v-if="chartRows.length"
          class="grid gap-4 lg:grid-cols-2"
        >
          <ServerMetricLineChart
            title="CPU"
            :data="chartRows"
            :x="x"
            :x-domain="xDomain"
            :y-domain="cpuYDomain"
            :y-tick-format="pctTick"
            :x-tick-format="xTickFormat"
            :series="cpuSeries"
            :format-tooltip="tipCpu"
          />
          <ServerMetricLineChart
            title="Memory"
            :data="chartRows"
            :x="x"
            :x-domain="xDomain"
            :y-domain="memYDomain"
            :y-tick-format="pctTick"
            :x-tick-format="xTickFormat"
            :series="memSeries"
            :format-tooltip="tipMem"
          />
          <ServerMetricLineChart
            title="Disk"
            :data="chartRows"
            :x="x"
            :x-domain="xDomain"
            :y-domain="diskYDomain"
            :y-tick-format="pctTick"
            :x-tick-format="xTickFormat"
            :series="diskSeries"
            :format-tooltip="tipDisk"
          />
          <ServerMetricLineChart
            title="Load average"
            :data="chartRows"
            :x="x"
            :x-domain="xDomain"
            :y-domain="loadYDomain"
            :x-tick-format="xTickFormat"
            :series="loadSeries"
            :format-tooltip="tipLoad"
          />
        </div>
        <div
          v-else
          class="text-muted text-sm"
        >
          Waiting for enough samples to draw charts.
        </div>
      </ClientOnly>
    </template>
  </div>
</template>
