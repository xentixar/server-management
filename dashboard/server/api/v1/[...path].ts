export default defineEventHandler((event) => {
  const config = useRuntimeConfig()
  const path = getRouterParam(event, 'path')

  return proxyRequest(event, `${config.apiBase}/api/v1/${path}`)
})
