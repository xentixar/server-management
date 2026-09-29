import type { AvatarProps } from '@nuxt/ui'

export type UserStatus = 'subscribed' | 'unsubscribed' | 'bounced'

export interface User {
  id: number
  name: string
  email: string
  avatar?: AvatarProps
  status: UserStatus
  location: string
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
  sender: User
  body: string
  date: string
}
