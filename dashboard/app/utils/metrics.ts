export type MetricsRange = '1h' | '6h' | '24h' | '7d'

export const METRICS_RANGES: { value: MetricsRange, label: string }[] = [
  { value: '1h', label: '1h' },
  { value: '6h', label: '6h' },
  { value: '24h', label: '24h' },
  { value: '7d', label: '7d' }
]

export interface MetricPointApi {
  t: string
  cpu_percent?: number | null
  mem_used_bytes?: number | null
  mem_total_bytes?: number | null
  disk_used_bytes?: number | null
  disk_total_bytes?: number | null
  load1?: number | null
  load5?: number | null
  load15?: number | null
}

export interface ServerMetricsApi {
  range: MetricsRange
  from: string
  to: string
  latest: MetricPointApi | null
  series: MetricPointApi[]
}

export interface MetricPoint {
  t: Date
  cpuPercent: number | null
  memUsedBytes: number | null
  memTotalBytes: number | null
  diskUsedBytes: number | null
  diskTotalBytes: number | null
  load1: number | null
  load5: number | null
  load15: number | null
}

export interface ServerMetrics {
  range: MetricsRange
  from: Date
  to: Date
  latest: MetricPoint | null
  series: MetricPoint[]
}

export function mapMetricPoint(p: MetricPointApi | null | undefined): MetricPoint | null {
  if (!p) {
    return null
  }
  return {
    t: new Date(p.t),
    cpuPercent: p.cpu_percent ?? null,
    memUsedBytes: p.mem_used_bytes ?? null,
    memTotalBytes: p.mem_total_bytes ?? null,
    diskUsedBytes: p.disk_used_bytes ?? null,
    diskTotalBytes: p.disk_total_bytes ?? null,
    load1: p.load1 ?? null,
    load5: p.load5 ?? null,
    load15: p.load15 ?? null
  }
}

export function mapServerMetrics(data: ServerMetricsApi): ServerMetrics {
  return {
    range: data.range,
    from: new Date(data.from),
    to: new Date(data.to),
    latest: mapMetricPoint(data.latest),
    series: (data.series ?? []).map(p => mapMetricPoint(p)!).filter(Boolean)
  }
}

export function usagePercent(used: number | null, total: number | null): number | null {
  if (used == null || total == null || total <= 0) {
    return null
  }
  return Math.min(100, Math.max(0, (used / total) * 100))
}

export function formatBytes(bytes: number | null | undefined): string {
  if (bytes == null || !Number.isFinite(bytes)) {
    return '-'
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let n = bytes
  let i = 0
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024
    i++
  }
  const digits = i === 0 ? 0 : n >= 10 ? 1 : 2
  return `${n.toFixed(digits)} ${units[i]}`
}

export function formatPercent(value: number | null | undefined): string {
  if (value == null || !Number.isFinite(value)) {
    return '-'
  }
  return `${value.toFixed(1)}%`
}

export function formatLoad(value: number | null | undefined): string {
  if (value == null || !Number.isFinite(value)) {
    return '-'
  }
  return value.toFixed(2)
}

export type MetricChartSeries<T = Record<string, unknown>> = {
  key: string
  color: string
  label: string
  y: (d: T) => number | null | undefined
  area?: boolean
}
