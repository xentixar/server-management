export type ProvisionStepStatus = 'pending' | 'running' | 'done' | 'failed'

export interface ProvisionStep {
  id: string
  label: string
  status: ProvisionStepStatus
}

export interface ProvisionLogLine {
  id: string
  at: number
  stream: 'stdout' | 'stderr' | 'system'
  text: string
}

export type ProvisionRunStatus = 'idle' | 'running' | 'done' | 'failed'

export interface ProvisionRun {
  status: ProvisionRunStatus
  steps: ProvisionStep[]
  logs: ProvisionLogLine[]
  startedAt: number | null
  finishedAt: number | null
}

const STEP_DEFS = [{
  id: 'ssh',
  label: 'Connect over SSH'
}, {
  id: 'upload',
  label: 'Upload bootstrap script'
}, {
  id: 'install',
  label: 'Install agent'
}, {
  id: 'register',
  label: 'Register with control plane'
}, {
  id: 'heartbeat',
  label: 'Confirm agent heartbeat'
}] as const

function freshSteps(): ProvisionStep[] {
  return STEP_DEFS.map(s => ({ id: s.id, label: s.label, status: 'pending' as const }))
}

// ponytail: fixed clock so SSR stub HTML matches client; swap for job stream API later
const ANCHOR = Date.UTC(2026, 8, 30, 12, 0, 0)

function completedRun(): ProvisionRun {
  return {
    status: 'done',
    steps: STEP_DEFS.map(s => ({ id: s.id, label: s.label, status: 'done' as const })),
    logs: [{
      id: 'seed_1',
      at: ANCHOR - 60000,
      stream: 'system',
      text: 'Provisioning completed earlier.'
    }, {
      id: 'seed_2',
      at: ANCHOR - 59000,
      stream: 'stdout',
      text: 'agent: heartbeat ok'
    }],
    startedAt: ANCHOR - 120000,
    finishedAt: ANCHOR - 58000
  }
}

function emptyRun(): ProvisionRun {
  return {
    status: 'idle',
    steps: freshSteps(),
    logs: [],
    startedAt: null,
    finishedAt: null
  }
}

const timers = new Map<string, ReturnType<typeof setTimeout>[]>()

function ensureRun(serverId: string, agent: string): ProvisionRun {
  const store = useState<Record<string, ProvisionRun>>('provision-runs', () => ({}))
  if (!store.value[serverId]) {
    store.value = {
      ...store.value,
      [serverId]: agent === 'pending' ? emptyRun() : completedRun()
    }
  }
  return store.value[serverId]!
}

function clearTimers(serverId: string) {
  const list = timers.get(serverId) || []
  for (const t of list) {
    clearTimeout(t)
  }
  timers.set(serverId, [])
}

function pushTimer(serverId: string, fn: () => void, delay: number) {
  const id = setTimeout(fn, delay)
  const list = timers.get(serverId) || []
  list.push(id)
  timers.set(serverId, list)
}

function appendLog(run: ProvisionRun, stream: ProvisionLogLine['stream'], text: string) {
  run.logs.push({
    id: `log_${Date.now()}_${run.logs.length}`,
    at: Date.now(),
    stream,
    text
  })
}

const STEP_LOGS: Record<string, { stream: ProvisionLogLine['stream'], text: string }[]> = {
  ssh: [
    { stream: 'system', text: 'Opening SSH session...' },
    { stream: 'stdout', text: 'Authenticated as deploy user' }
  ],
  upload: [
    { stream: 'stdout', text: 'Uploading /tmp/outpost-bootstrap.sh (12 KB)' },
    { stream: 'stdout', text: 'Checksum ok' }
  ],
  install: [
    { stream: 'stdout', text: 'Installing outpost-agent...' },
    { stream: 'stdout', text: 'Created systemd unit outpost-agent.service' },
    { stream: 'stdout', text: 'Started outpost-agent.service' }
  ],
  register: [
    { stream: 'stdout', text: 'Requesting enrollment token...' },
    { stream: 'stdout', text: 'Agent registered with control plane' }
  ],
  heartbeat: [
    { stream: 'stdout', text: 'Waiting for first heartbeat...' },
    { stream: 'stdout', text: 'Heartbeat received (latency 84ms)' },
    { stream: 'system', text: 'Server is online' }
  ]
}

export function useServerProvision(serverId: MaybeRefOrGetter<string>) {
  const { getById, setAgent } = useServers()

  const id = computed(() => toValue(serverId))

  const server = computed(() => getById(id.value))

  const run = computed(() => {
    const s = server.value
    if (!s) {
      return emptyRun()
    }
    return ensureRun(s.id, s.agent)
  })

  const progress = computed(() => {
    const steps = run.value.steps
    if (!steps.length) {
      return 0
    }
    const done = steps.filter(s => s.status === 'done').length
    const running = steps.some(s => s.status === 'running') ? 0.45 : 0
    return Math.min(100, Math.round(((done + running) / steps.length) * 100))
  })

  function start(force = false) {
    const s = server.value
    if (!s) {
      return
    }

    const existing = ensureRun(s.id, s.agent)
    if (existing.status === 'running') {
      return
    }
    if (existing.status === 'done' && !force) {
      return
    }

    clearTimers(s.id)
    setAgent(s.id, 'pending')

    const store = useState<Record<string, ProvisionRun>>('provision-runs', () => ({}))
    const next = emptyRun()
    next.status = 'running'
    next.startedAt = Date.now()
    store.value = { ...store.value, [s.id]: next }
    appendLog(next, 'system', `Provisioning ${s.hostname} (${s.ip})`)

    let delay = 400
    STEP_DEFS.forEach((def, index) => {
      pushTimer(s.id, () => {
        const current = store.value[s.id]
        if (!current || current.status !== 'running') {
          return
        }
        current.steps[index]!.status = 'running'
        for (const line of STEP_LOGS[def.id] || []) {
          appendLog(current, line.stream, line.text)
        }
      }, delay)

      delay += 900 + index * 200

      pushTimer(s.id, () => {
        const current = store.value[s.id]
        if (!current || current.status !== 'running') {
          return
        }
        current.steps[index]!.status = 'done'

        if (index === STEP_DEFS.length - 1) {
          current.status = 'done'
          current.finishedAt = Date.now()
          setAgent(s.id, 'online')
        }
      }, delay)

      delay += 350
    })
  }

  function replay() {
    start(true)
  }

  onBeforeUnmount(() => {
    // leave timers running so background complete still updates agent
  })

  return {
    server,
    run,
    progress,
    start,
    replay
  }
}
