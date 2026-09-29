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

export interface OverviewServer {
  id: string
  hostname: string
  region: string
  agent: 'online' | 'offline' | 'degraded'
  sites: number
}

export interface OrgOverview {
  servers: number
  sites: number
  deploys7d: number
  openAlerts: number
  serversList: OverviewServer[]
  alerts: OverviewAlert[]
  activity: OverviewActivity[]
}

function emptyOverview(): OrgOverview {
  return {
    servers: 0,
    sites: 0,
    deploys7d: 0,
    openAlerts: 0,
    serversList: [],
    alerts: [],
    activity: []
  }
}

function sampleOverview(): OrgOverview {
  return {
    servers: 3,
    sites: 8,
    deploys7d: 14,
    openAlerts: 2,
    serversList: [{
      id: 'srv_1',
      hostname: 'edge-01.acme.dev',
      region: 'nyc1',
      agent: 'online',
      sites: 4
    }, {
      id: 'srv_2',
      hostname: 'app-02.acme.dev',
      region: 'sfo3',
      agent: 'degraded',
      sites: 3
    }, {
      id: 'srv_3',
      hostname: 'worker-03.acme.dev',
      region: 'ams3',
      agent: 'offline',
      sites: 1
    }],
    alerts: [{
      id: 'al_1',
      severity: 'error',
      title: 'Agent offline',
      detail: 'worker-03.acme.dev has not heartbeated in 18 minutes.'
    }, {
      id: 'al_2',
      severity: 'warning',
      title: 'Certificate expiring',
      detail: 'shop.acme.dev expires in 9 days.'
    }],
    activity: [{
      id: 'act_1',
      label: 'Deploy succeeded',
      target: 'api.acme.dev → main',
      status: 'success',
      at: new Date(Date.now() - 1000 * 60 * 18).toISOString()
    }, {
      id: 'act_2',
      label: 'Deploy failed',
      target: 'shop.acme.dev → release/2.4',
      status: 'failed',
      at: new Date(Date.now() - 1000 * 60 * 55).toISOString()
    }, {
      id: 'act_3',
      label: 'Deploy running',
      target: 'docs.acme.dev → main',
      status: 'running',
      at: new Date(Date.now() - 1000 * 60 * 4).toISOString()
    }, {
      id: 'act_4',
      label: 'Server provisioned',
      target: 'edge-01.acme.dev',
      status: 'success',
      at: new Date(Date.now() - 1000 * 60 * 60 * 26).toISOString()
    }]
  }
}

// ponytail: stub overview until servers/deploys API exists
const byOrg = reactive<Record<string, OrgOverview>>({})

function ensure(orgId: string): OrgOverview {
  if (!byOrg[orgId]) {
    // Outpost (id 1) starts empty; others get sample fleet for UI review
    byOrg[orgId] = orgId === '1' ? emptyOverview() : sampleOverview()
  }
  return byOrg[orgId]!
}

export function useOrgOverview() {
  const { selected } = useOrganizations()

  const overview = computed(() => ensure(selected.value.id))

  const isEmpty = computed(() => overview.value.servers === 0)

  return {
    overview,
    isEmpty
  }
}
