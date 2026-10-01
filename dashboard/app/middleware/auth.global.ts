const guestPaths = new Set([
  '/login',
  '/register',
  '/forgot-password',
  '/reset-password',
  '/verify-email'
])

function isInvitePath(path: string) {
  return /^\/invites\/[^/]+$/.test(path)
}

export default defineNuxtRouteMiddleware(async (to) => {
  const session = useSession()
  await session.ensureSession()

  const isGuestRoute = guestPaths.has(to.path)
  const inviteRoute = isInvitePath(to.path)

  if (isGuestRoute) {
    if (session.token.value) {
      const redirect = typeof to.query.redirect === 'string' ? to.query.redirect : '/'
      return navigateTo(redirect)
    }
    return
  }

  if (inviteRoute) {
    if (session.token.value) {
      await ensureOrganizations()
    }
    return
  }

  if (!session.token.value) {
    return navigateTo({
      path: '/login',
      query: to.path === '/' ? undefined : { redirect: to.fullPath }
    })
  }

  await ensureOrganizations()
})
