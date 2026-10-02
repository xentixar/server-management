import type {
  BillingInterval,
  Invoice,
  PaymentMethod,
  PlanId,
  SubscriptionState
} from '~/utils/billing'
import { planById } from '~/utils/billing'

// ponytail: fixed clock so SSR stub HTML matches client; swap for API later
const ANCHOR = Date.UTC(2026, 8, 30, 12, 0, 0)

function daysFromAnchor(days: number) {
  return new Date(ANCHOR + days * 86400000).toISOString()
}

function defaultSubscription(orgId: string): SubscriptionState {
  if (orgId === '1') {
    return {
      planId: 'free',
      interval: 'monthly',
      status: 'active',
      renewsAt: null,
      cancelAtPeriodEnd: false,
      usage: { servers: 0, sites: 0, seats: 1 },
      paymentMethod: null,
      invoices: [{
        id: 'inv_1001',
        number: 'INV-1001',
        date: '2026-08-01T00:00:00.000Z',
        amount: 0,
        status: 'paid'
      }, {
        id: 'inv_1002',
        number: 'INV-1002',
        date: '2026-09-01T00:00:00.000Z',
        amount: 0,
        status: 'paid'
      }]
    }
  }

  // Acme sample fleet (3 servers) needs Pro room
  return {
    planId: 'pro',
    interval: 'monthly',
    status: 'active',
    renewsAt: daysFromAnchor(20),
    cancelAtPeriodEnd: false,
    usage: { servers: 3, sites: 8, seats: 3 },
    paymentMethod: {
      brand: 'Visa',
      last4: '4242',
      expMonth: 12,
      expYear: 2028
    },
    invoices: [{
      id: 'inv_2001',
      number: 'INV-2001',
      date: '2026-09-01T00:00:00.000Z',
      amount: 29,
      status: 'paid'
    }]
  }
}

function nextRenewal(interval: BillingInterval) {
  const d = new Date(ANCHOR)
  if (interval === 'yearly') {
    d.setUTCFullYear(d.getUTCFullYear() + 1)
  } else {
    d.setUTCMonth(d.getUTCMonth() + 1)
  }
  return d.toISOString()
}

export function useBilling() {
  const { selected } = useOrganizations()
  const store = useState<Record<string, SubscriptionState>>('billing-by-org', () => ({}))

  function ensure(orgId: string | number): SubscriptionState {
    const key = String(orgId)
    if (!store.value[key]) {
      store.value = { ...store.value, [key]: defaultSubscription(key) }
    }
    return store.value[key]!
  }

  const subscription = computed(() => ensure(selected.value?.id ?? '_'))
  const currentPlan = computed(() => planById(subscription.value.planId))

  function changePlan(planId: PlanId, interval: BillingInterval) {
    const sub = ensure(selected.value?.id ?? '_')
    sub.planId = planId
    sub.interval = interval
    sub.status = 'active'
    sub.cancelAtPeriodEnd = false
    sub.renewsAt = planId === 'free' ? null : nextRenewal(interval)

    if (planId !== 'free' && !sub.paymentMethod) {
      sub.paymentMethod = {
        brand: 'Visa',
        last4: '4242',
        expMonth: 12,
        expYear: 2028
      }
    }

    if (planId !== 'free') {
      const plan = planById(planId)
      const amount = interval === 'yearly' ? plan.priceYearly : plan.priceMonthly
      const invoice: Invoice = {
        id: `inv_${Date.now()}`,
        number: `INV-${String(sub.invoices.length + 1001)}`,
        date: new Date().toISOString(),
        amount,
        status: 'paid'
      }
      sub.invoices = [invoice, ...sub.invoices]
    }
  }

  function setPaymentMethod(method: PaymentMethod) {
    ensure(selected.value?.id ?? '_').paymentMethod = method
  }

  function removePaymentMethod() {
    const sub = ensure(selected.value?.id ?? '_')
    if (sub.planId === 'free') {
      sub.paymentMethod = null
    }
  }

  function cancelAtPeriodEnd() {
    const sub = ensure(selected.value?.id ?? '_')
    if (sub.planId === 'free') {
      return
    }
    sub.cancelAtPeriodEnd = true
    if (!sub.renewsAt) {
      sub.renewsAt = nextRenewal(sub.interval)
    }
  }

  function resumeSubscription() {
    const sub = ensure(selected.value?.id ?? '_')
    sub.cancelAtPeriodEnd = false
    sub.status = 'active'
  }

  function setUsage(partial: Partial<SubscriptionState['usage']>) {
    const sub = ensure(selected.value?.id ?? '_')
    sub.usage = { ...sub.usage, ...partial }
  }

  return {
    subscription,
    currentPlan,
    changePlan,
    setPaymentMethod,
    removePaymentMethod,
    cancelAtPeriodEnd,
    resumeSubscription,
    setUsage
  }
}
