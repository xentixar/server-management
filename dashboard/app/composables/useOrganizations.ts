import type { Organization } from '~/types'

const LAST_ORG_KEY = 'outpost:last-org-slug'

function slugify(name: string) {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '') || 'org'
}

// ponytail: module singleton until real session/org API exists
const organizations = ref<Organization[]>([{
  id: '1',
  name: 'Outpost',
  slug: 'outpost',
  avatar: { alt: 'Outpost' }
}, {
  id: '2',
  name: 'Acme Labs',
  slug: 'acme-labs',
  avatar: { alt: 'Acme Labs' }
}])

function uniqueSlug(base: string, exceptId?: string) {
  let slug = slugify(base)
  let n = 2
  while (organizations.value.some(o => o.slug === slug && o.id !== exceptId)) {
    slug = `${slugify(base)}-${n}`
    n += 1
  }
  return slug
}

export function useOrganizations() {
  const route = useRoute()
  const router = useRouter()

  const routeSlug = computed(() => {
    const raw = route.params.orgSlug
    return typeof raw === 'string' ? raw : Array.isArray(raw) ? raw[0] ?? '' : ''
  })

  const selected = computed<Organization>(() => {
    return organizations.value.find(o => o.slug === routeSlug.value) ?? organizations.value[0]!
  })

  function resolveLandingSlug() {
    if (import.meta.client) {
      const last = localStorage.getItem(LAST_ORG_KEY)
      if (last && organizations.value.some(o => o.slug === last)) {
        return last
      }
    }
    return organizations.value[0]!.slug
  }

  function rememberSlug(slug: string) {
    if (import.meta.client) {
      localStorage.setItem(LAST_ORG_KEY, slug)
    }
  }

  function orgHref(path = '') {
    const base = `/o/${selected.value.slug}`
    if (!path || path === '/') {
      return base
    }
    return `${base}${path.startsWith('/') ? path : `/${path}`}`
  }

  function pathAfterOrg() {
    return route.path.replace(/^\/o\/[^/]+/, '') || ''
  }

  function switchTo(slug: string) {
    if (!organizations.value.some(o => o.slug === slug)) {
      return
    }
    rememberSlug(slug)
    return router.push(`/o/${slug}${pathAfterOrg()}`)
  }

  function create(name: string): Organization {
    const org: Organization = {
      id: String(Date.now()),
      name: name.trim(),
      slug: uniqueSlug(name),
      avatar: { alt: name.trim() }
    }
    organizations.value.push(org)
    rememberSlug(org.slug)
    return org
  }

  function updateSelected(patch: Partial<Pick<Organization, 'name' | 'slug' | 'avatar'>>) {
    const org = organizations.value.find(o => o.id === selected.value.id)
    if (!org) {
      return
    }

    const previousSlug = org.slug

    if (patch.name !== undefined) {
      org.name = patch.name
      org.avatar = { ...org.avatar, alt: patch.name }
    }
    if (patch.slug !== undefined) {
      org.slug = uniqueSlug(patch.slug, org.id)
    }
    if (patch.avatar !== undefined) {
      org.avatar = patch.avatar
    }

    if (org.slug !== previousSlug) {
      rememberSlug(org.slug)
      return router.replace(`/o/${org.slug}${pathAfterOrg()}`)
    }

    rememberSlug(org.slug)
  }

  function findBySlug(slug: string) {
    return organizations.value.find(o => o.slug === slug)
  }

  return {
    organizations,
    selected,
    routeSlug,
    resolveLandingSlug,
    rememberSlug,
    orgHref,
    switchTo,
    create,
    updateSelected,
    findBySlug,
    slugify,
    uniqueSlug
  }
}
