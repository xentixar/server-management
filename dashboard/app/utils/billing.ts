export type PlanId = 'free' | 'pro' | 'business'
export type BillingInterval = 'monthly' | 'yearly'
export type InvoiceStatus = 'paid' | 'open' | 'failed'

export interface PlanLimit {
  key: 'servers' | 'sites' | 'seats'
  label: string
  max: number
}

export interface PlanDefinition {
  id: PlanId
  name: string
  description: string
  priceMonthly: number
  priceYearly: number
  limits: PlanLimit[]
  features: string[]
  highlighted?: boolean
}

export interface PlanUsage {
  servers: number
  sites: number
  seats: number
}

export interface PaymentMethod {
  brand: string
  last4: string
  expMonth: number
  expYear: number
}

export interface Invoice {
  id: string
  number: string
  date: string
  amount: number
  status: InvoiceStatus
  pdfUrl?: string
}

export interface SubscriptionState {
  planId: PlanId
  interval: BillingInterval
  status: 'active' | 'trialing' | 'canceled'
  renewsAt: string | null
  cancelAtPeriodEnd: boolean
  usage: PlanUsage
  paymentMethod: PaymentMethod | null
  invoices: Invoice[]
}

export const PLAN_CATALOG: PlanDefinition[] = [{
  id: 'free',
  name: 'Free',
  description: 'For trying Outpost on a single server.',
  priceMonthly: 0,
  priceYearly: 0,
  limits: [
    { key: 'servers', label: 'Servers', max: 1 },
    { key: 'sites', label: 'Sites', max: 3 },
    { key: 'seats', label: 'Seats', max: 3 }
  ],
  features: ['Manual server add', 'Basic deploys', 'Community support']
}, {
  id: 'pro',
  name: 'Pro',
  description: 'For small teams running production sites.',
  priceMonthly: 29,
  priceYearly: 290,
  highlighted: true,
  limits: [
    { key: 'servers', label: 'Servers', max: 5 },
    { key: 'sites', label: 'Sites', max: 25 },
    { key: 'seats', label: 'Seats', max: 10 }
  ],
  features: ['Terraform provision', 'Zero-downtime deploys', 'Priority email support']
}, {
  id: 'business',
  name: 'Business',
  description: 'For agencies and larger fleets.',
  priceMonthly: 99,
  priceYearly: 990,
  limits: [
    { key: 'servers', label: 'Servers', max: 25 },
    { key: 'sites', label: 'Sites', max: 100 },
    { key: 'seats', label: 'Seats', max: 50 }
  ],
  features: ['Audit log', 'SSO-ready seats', 'Dedicated support']
}]

export function planById(id: PlanId) {
  return PLAN_CATALOG.find(p => p.id === id)!
}

export function formatMoney(amount: number, interval?: BillingInterval) {
  if (amount === 0) {
    return 'Free'
  }
  const suffix = interval === 'yearly' ? '/yr' : interval === 'monthly' ? '/mo' : ''
  return `$${amount}${suffix}`
}

export function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    timeZone: 'UTC'
  })
}
