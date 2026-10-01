import { createSharedComposable } from '@vueuse/core'
import type { Organization } from '~/types'

const LAST_ORG_KEY = 'outpost:last-org-slug'

function orgState() {
  return {
    organizations: useState<Organization[]>('organizations', () => []),
    ready: useState('orgs-ready', () => false),
    pending: useState('orgs-pending', () => false)
  }
}

function rememberSlug(slug: string) {
  if (import.meta.client) {
    localStorage.setItem(LAST_ORG_KEY, slug)
  }
}

/** Safe to call from middleware / session — no useRoute. */
export async function ensureOrganizations() {
  const { api } = useApi()
  const { organizations, ready, pending } = orgState()

  if (ready.value && organizations.value.every(o => Array.isArray(o.permissions))) {
    return organizations.value
  }

  pending.value = true
  try {
    const res = await api<Organization[]>('/api/v1/organizations')
    organizations.value = res.data
  } finally {
    ready.value = true
    pending.value = false
  }

  return organizations.value
}

/** Safe to call from middleware / session — no useRoute. */
export function clearOrganizations() {
  const { organizations, ready } = orgState()
  organizations.value = []
  ready.value = false
}

const _useOrganizations = () => {
  const { api } = useApi()
  const route = useRoute()
  const router = useRouter()
  const { organizations, ready, pending } = orgState()

  const routeSlug = computed(() => {
    const raw = route.params.orgSlug
    return typeof raw === 'string' ? raw : Array.isArray(raw) ? raw[0] ?? '' : ''
  })

  const selected = computed(() => {
    return organizations.value.find(o => o.slug === routeSlug.value) ?? organizations.value[0]
  })

  function resolveLandingSlug() {
    if (import.meta.client) {
      const last = localStorage.getItem(LAST_ORG_KEY)
      if (last && organizations.value.some(o => o.slug === last)) {
        return last
      }
    }
    return organizations.value[0]?.slug ?? null
  }

  function findBySlug(slug: string) {
    return organizations.value.find(o => o.slug === slug)
  }

  function pathAfterOrg() {
    return route.path.replace(/^\/o\/[^/]+/, '') || ''
  }

  function orgHref(path = '') {
    const slug = selected.value?.slug ?? resolveLandingSlug()
    if (!slug) {
      return '/'
    }
    const base = `/o/${slug}`
    if (!path || path === '/') {
      return base
    }
    return `${base}${path.startsWith('/') ? path : `/${path}`}`
  }

  function switchTo(slug: string) {
    if (!organizations.value.some(o => o.slug === slug)) {
      return
    }
    rememberSlug(slug)
    return router.push(`/o/${slug}${pathAfterOrg()}`)
  }

  async function fetchOrganizations() {
    return ensureOrganizations()
  }

  async function create(name: string) {
    pending.value = true
    try {
      const res = await api<Organization>('/api/v1/organizations', {
        method: 'POST',
        body: { name: name.trim() }
      })
      organizations.value.push(res.data)
      rememberSlug(res.data.slug)
      return res.data
    } finally {
      pending.value = false
    }
  }

  async function updateSelected(name: string) {
    const org = selected.value
    if (!org) {
      return
    }

    pending.value = true
    try {
      const previousSlug = org.slug
      const res = await api<Organization>(`/api/v1/organizations/${org.id}`, {
        method: 'PATCH',
        body: { name: name.trim() }
      })

      const idx = organizations.value.findIndex(o => o.id === org.id)
      if (idx !== -1) {
        organizations.value[idx] = res.data
      }

      if (res.data.slug !== previousSlug) {
        rememberSlug(res.data.slug)
        return router.replace(`/o/${res.data.slug}${pathAfterOrg()}`)
      }

      rememberSlug(res.data.slug)
      return res.data
    } finally {
      pending.value = false
    }
  }

  async function removeSelected() {
    const org = selected.value
    if (!org) {
      return
    }

    pending.value = true
    try {
      await api(`/api/v1/organizations/${org.id}`, { method: 'DELETE' })
      organizations.value = organizations.value.filter(o => o.id !== org.id)

      const next = organizations.value[0]
      if (!next) {
        if (import.meta.client) {
          localStorage.removeItem(LAST_ORG_KEY)
        }
        return navigateTo('/', { replace: true })
      }

      rememberSlug(next.slug)
      return navigateTo(`/o/${next.slug}${pathAfterOrg()}`, { replace: true })
    } finally {
      pending.value = false
    }
  }

  return {
    organizations,
    selected,
    routeSlug,
    ready,
    pending,
    ensureOrganizations,
    fetchOrganizations,
    clearOrganizations,
    resolveLandingSlug,
    rememberSlug,
    orgHref,
    switchTo,
    create,
    updateSelected,
    removeSelected,
    findBySlug
  }
}

export const useOrganizations = createSharedComposable(_useOrganizations)
