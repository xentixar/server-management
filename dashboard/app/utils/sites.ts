export type SiteStack = 'static' | 'php' | 'laravel' | 'node' | 'nextjs' | 'nuxt' | 'custom'
export type SiteStatus = 'active' | 'deploying' | 'failed' | 'disabled'

export interface Site {
  id: string
  serverId: string
  name: string
  domain: string
  stack: SiteStack
  webRoot: string
  status: SiteStatus
  gitRepo?: string
  gitBranch?: string
  createdAt: string
}

export interface CreateSiteInput {
  serverId: string
  name: string
  domain: string
  stack: SiteStack
  webRoot: string
  gitRepo?: string
  gitBranch?: string
}

export const SITE_STACKS: { value: SiteStack, label: string, description: string }[] = [{
  value: 'static',
  label: 'Static',
  description: 'HTML and assets, optional build step'
}, {
  value: 'php',
  label: 'PHP',
  description: 'Generic PHP with document root'
}, {
  value: 'laravel',
  label: 'Laravel',
  description: 'PHP + Composer, public/ root'
}, {
  value: 'node',
  label: 'Node',
  description: 'Generic Node app'
}, {
  value: 'nextjs',
  label: 'Next.js',
  description: 'Node SSR or static export'
}, {
  value: 'nuxt',
  label: 'Nuxt',
  description: 'Nitro output'
}, {
  value: 'custom',
  label: 'Custom',
  description: 'Your own build and start commands'
}]

export const STACK_OPTIONS = SITE_STACKS.map(s => ({
  label: s.label,
  value: s.value
}))

export function stackLabel(stack: SiteStack) {
  return SITE_STACKS.find(s => s.value === stack)?.label ?? stack
}

export function siteStatusColor(status: SiteStatus) {
  if (status === 'active') {
    return 'success' as const
  }
  if (status === 'deploying') {
    return 'primary' as const
  }
  if (status === 'failed') {
    return 'error' as const
  }
  return 'neutral' as const
}

export function defaultWebRoot(stack: SiteStack) {
  if (stack === 'laravel') {
    return 'public'
  }
  if (stack === 'nextjs' || stack === 'nuxt') {
    return '.output/public'
  }
  return 'public'
}
