import { toApiError, type ApiResponse } from '~/utils/api'

type ApiMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

interface ApiOptions {
  method?: ApiMethod
  body?: Record<string, unknown> | BodyInit | null
  headers?: HeadersInit
}

export function useApi() {
  const token = useCookie<string | null>('outpost_token', {
    sameSite: 'lax',
    maxAge: 60 * 60 * 24 * 30
  })

  async function api<T>(path: string, options: ApiOptions = {}) {
    const headers = new Headers(options.headers)

    if (token.value) {
      headers.set('Authorization', `Bearer ${token.value}`)
    }

    try {
      // same-origin Nitro proxy → Go API (no CORS)
      return await $fetch<ApiResponse<T>>(path, {
        method: options.method,
        body: options.body,
        headers
      })
    } catch (error) {
      throw toApiError(error)
    }
  }

  return { api, token }
}
