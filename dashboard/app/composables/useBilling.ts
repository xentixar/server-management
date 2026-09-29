import type {
  BillingInterval,
  Invoice,
  PaymentMethod,
  PlanId,
  SubscriptionState
} from '~/utils/billing'
import { planById } from '~/utils/billing'

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
    renewsAt: new Date(Date.now() + 1000 * 60 * 60 * 24 * 20).toISOString(),
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

// ponytail: local stub until billing API exists; keyed by org id
const byOrg = reactive<Record<string, SubscriptionState>>({})

function ensure(orgId: string): SubscriptionState {
  if (!byOrg[orgId]) {
    byOrg[orgId] = defaultSubscription(orgId)
  }
  return byOrg[orgId]!
}

function nextRenewal(interval: BillingInterval) {
  const d = new Date()
  if (interval === 'yearly') {
    d.setFullYear(d.getFullYear() + 1)
  } else {
    d.setMonth(d.getMonth() + 1)
  }
  return d.toISOString()
}

export function useBilling() {
  const { selected } = useOrganizations()

  const subscription = computed(() => ensure(selected.value.id))
  const currentPlan = computed(() => planById(subscription.value.planId))

  function changePlan(planId: PlanId, interval: BillingInterval) {
    const sub = ensure(selected.value.id)
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
    ensure(selected.value.id).paymentMethod = method
  }

  function removePaymentMethod() {
    const sub = ensure(selected.value.id)
    if (sub.planId === 'free') {
      sub.paymentMethod = null
    }
  }

  function cancelAtPeriodEnd() {
    const sub = ensure(selected.value.id)
    if (sub.planId === 'free') {
      return
    }
    sub.cancelAtPeriodEnd = true
    if (!sub.renewsAt) {
      sub.renewsAt = nextRenewal(sub.interval)
    }
  }

  function resumeSubscription() {
    const sub = ensure(selected.value.id)
    sub.cancelAtPeriodEnd = false
    sub.status = 'active'
  }

  function setUsage(partial: Partial<SubscriptionState['usage']>) {
    const sub = ensure(selected.value.id)
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
