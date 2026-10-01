export interface AuthUser {
  id: number
  name: string
  email: string
}

export interface Organization {
  id: number
  name: string
  slug: string
  role?: 'owner' | 'admin' | 'viewer'
  permissions?: string[]
}

export interface Notification {
  id: number
  unread?: boolean
  title: string
  body?: string
  date: string
}
