import { createSharedComposable } from '@vueuse/core'
import type { AuthUser } from '~/types'
import { ApiError } from '~/utils/api'

interface LoginData {
  access_token: string
}

const _useSession = () => {
  const { api, token } = useApi()
  const user = useState<AuthUser | null>('auth-user', () => null)
  const ready = useState('auth-ready', () => false)
  const pending = useState('auth-pending', () => false)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  async function fetchUser() {
    if (!token.value) {
      user.value = null
      return null
    }

    try {
      const res = await api<AuthUser>('/api/v1/auth/user')
      user.value = res.data
      return user.value
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        clearSession()
      }
      throw error
    }
  }

  async function ensureSession() {
    if (ready.value) {
      return user.value
    }

    pending.value = true
    try {
      if (token.value) {
        await fetchUser().catch(() => null)
      }
    } finally {
      ready.value = true
      pending.value = false
    }

    return user.value
  }

  async function login(email: string, password: string) {
    pending.value = true
    try {
      const res = await api<LoginData>('/api/v1/auth/login', {
        method: 'POST',
        body: { email, password }
      })
      token.value = res.data.access_token
      await fetchUser()
      ready.value = true
      return user.value
    } finally {
      pending.value = false
    }
  }

  async function register(name: string, email: string, password: string) {
    pending.value = true
    try {
      await api('/api/v1/auth/register', {
        method: 'POST',
        body: { name, email, password }
      })
    } finally {
      pending.value = false
    }
  }

  function clearSession() {
    token.value = null
    user.value = null
    useOrganizations().clearOrganizations()
  }

  async function logout() {
    clearSession()
    ready.value = true
    await navigateTo('/login')
  }

  return {
    user,
    token,
    ready,
    pending,
    isAuthenticated,
    ensureSession,
    fetchUser,
    login,
    register,
    logout,
    clearSession
  }
}

export const useSession = createSharedComposable(_useSession)
