const guestPaths = new Set([
  '/login',
  '/register',
  '/forgot-password',
  '/reset-password',
  '/verify-email'
])

export default defineNuxtRouteMiddleware(async (to) => {
  const session = useSession()
  await session.ensureSession()

  const isGuestRoute = guestPaths.has(to.path)

  if (isGuestRoute) {
    if (session.token.value) {
      return navigateTo('/')
    }
    return
  }

  if (!session.token.value) {
    return navigateTo({
      path: '/login',
      query: to.path === '/' ? undefined : { redirect: to.fullPath }
    })
  }
})
