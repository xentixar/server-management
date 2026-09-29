import type { Organization } from '~/types'

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

const selectedId = ref('1')

export function useOrganizations() {
  const selected = computed<Organization>(() => {
    return organizations.value.find(o => o.id === selectedId.value) ?? organizations.value[0]!
  })

  function select(id: string) {
    if (organizations.value.some(o => o.id === id)) {
      selectedId.value = id
    }
  }

  function create(name: string): Organization {
    const org: Organization = {
      id: String(Date.now()),
      name: name.trim(),
      slug: slugify(name),
      avatar: { alt: name.trim() }
    }
    organizations.value.push(org)
    selectedId.value = org.id
    return org
  }

  function updateSelected(patch: Partial<Pick<Organization, 'name' | 'slug' | 'avatar'>>) {
    const org = organizations.value.find(o => o.id === selectedId.value)
    if (!org) {
      return
    }
    if (patch.name !== undefined) {
      org.name = patch.name
      org.avatar = { ...org.avatar, alt: patch.name }
    }
    if (patch.slug !== undefined) {
      org.slug = patch.slug
    }
    if (patch.avatar !== undefined) {
      org.avatar = patch.avatar
    }
  }

  return {
    organizations,
    selected,
    selectedId,
    select,
    create,
    updateSelected,
    slugify
  }
}
