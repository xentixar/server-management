export type AlertSeverity = 'warning' | 'error' | 'info'

export interface OverviewAlert {
  id: string
  severity: AlertSeverity
  title: string
  detail: string
  href?: string
}

export interface OverviewActivity {
  id: string
  label: string
  target: string
  status: 'success' | 'failed' | 'running'
  at: string
}

const ANCHOR = Date.UTC(2026, 8, 30, 12, 0, 0)

function ago(ms: number) {
  return new Date(ANCHOR - ms).toISOString()
}

function sampleMeta(orgId: string) {
  if (orgId === '1') {
    return {
      deploys7d: 0,
      alerts: [] as OverviewAlert[],
      activity: [] as OverviewActivity[]
    }
  }

  return {
    deploys7d: 14,
    alerts: [{
      id: 'al_1',
      severity: 'error' as const,
      title: 'Agent offline',
      detail: 'worker-03.acme.dev has not heartbeated in 18 minutes.'
    }, {
      id: 'al_2',
      severity: 'warning' as const,
      title: 'Certificate expiring',
      detail: 'shop.acme.dev expires in 9 days.'
    }] satisfies OverviewAlert[],
    activity: [{
      id: 'act_1',
      label: 'Deploy succeeded',
      target: 'api.acme.dev -> main',
      status: 'success' as const,
      at: ago(1000 * 60 * 18)
    }, {
      id: 'act_2',
      label: 'Deploy failed',
      target: 'shop.acme.dev -> release/2.4',
      status: 'failed' as const,
      at: ago(1000 * 60 * 55)
    }, {
      id: 'act_3',
      label: 'Deploy running',
      target: 'docs.acme.dev -> main',
      status: 'running' as const,
      at: ago(1000 * 60 * 4)
    }, {
      id: 'act_4',
      label: 'Server provisioned',
      target: 'edge-01.acme.dev',
      status: 'success' as const,
      at: ago(1000 * 60 * 60 * 26)
    }] satisfies OverviewActivity[]
  }
}

function ensureMeta(orgId: string | number) {
  const key = String(orgId)
  const store = useState<Record<string, ReturnType<typeof sampleMeta>>>('org-overview-meta', () => ({}))
  if (!store.value[key]) {
    store.value = { ...store.value, [key]: sampleMeta(key) }
  }
  return store.value[key]!
}

export function useOrgOverview() {
  const { selected } = useOrganizations()
  const { servers } = useServers()

  const overview = computed(() => {
    const meta = ensureMeta(selected.value?.id ?? '_')
    const list = servers.value
    return {
      servers: list.length,
      sites: list.reduce((sum, s) => sum + s.sites, 0),
      deploys7d: meta.deploys7d,
      openAlerts: meta.alerts.length,
      serversList: list.map(s => ({
        id: s.id,
        hostname: s.hostname,
        region: s.region || 'n/a',
        agent: s.agent,
        sites: s.sites
      })),
      alerts: meta.alerts,
      activity: meta.activity
    }
  })

  const isEmpty = computed(() => overview.value.servers === 0)

  return {
    overview,
    isEmpty
  }
}
