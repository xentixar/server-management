import { ApiError } from '~/utils/api'
import {
  mapServerMetrics,
  type MetricsRange,
  type ServerMetrics,
  type ServerMetricsApi
} from '~/utils/metrics'

export function useServerMetrics(serverId: MaybeRefOrGetter<string | number>) {
  const { api } = useApi()
  const { selected } = useOrganizations()
  const metrics = ref<ServerMetrics | null>(null)
  const range = ref<MetricsRange>('1h')
  const loading = ref(false)

  function basePath() {
    const org = selected.value
    if (!org) {
      throw new ApiError('No organization selected', 400)
    }
    return `/api/v1/organizations/${org.id}/servers/${toValue(serverId)}`
  }

  async function fetchMetrics(nextRange: MetricsRange = range.value) {
    loading.value = true
    try {
      range.value = nextRange
      const res = await api<ServerMetricsApi>(`${basePath()}/metrics?range=${encodeURIComponent(nextRange)}`)
      metrics.value = mapServerMetrics(res.data)
      return metrics.value
    } finally {
      loading.value = false
    }
  }

  return {
    metrics,
    range,
    loading,
    fetchMetrics
  }
}
