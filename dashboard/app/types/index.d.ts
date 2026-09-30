export interface AuthUser {
  id: number
  name: string
  email: string
}

export interface Organization {
  id: string
  name: string
  slug: string
  avatar?: {
    src?: string
    alt?: string
  }
}

export interface Notification {
  id: number
  unread?: boolean
  title: string
  body?: string
  date: string
}
