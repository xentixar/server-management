<script setup lang="ts">
import { formatDate, formatMoney } from '~/utils/billing'
import type { Invoice } from '~/utils/billing'

const { selected } = useOrganizations()
const {
  subscription,
  currentPlan,
  changePlan,
  setPaymentMethod,
  cancelAtPeriodEnd,
  resumeSubscription
} = useBilling()

const toast = useToast()
const changeOpen = ref(false)
const paymentOpen = ref(false)
const cancelOpen = ref(false)

const usageRows = computed(() =>
  currentPlan.value.limits.map(limit => ({
    ...limit,
    used: subscription.value.usage[limit.key],
    pct: Math.min(100, Math.round((subscription.value.usage[limit.key] / limit.max) * 100))
  }))
)

const renewalLabel = computed(() => {
  const sub = subscription.value
  if (sub.planId === 'free') {
    return 'No renewal on the Free plan'
  }
  if (sub.cancelAtPeriodEnd && sub.renewsAt) {
    return `Cancels on ${formatDate(sub.renewsAt)}`
  }
  if (sub.renewsAt) {
    return `Renews on ${formatDate(sub.renewsAt)}`
  }
  return 'Active'
})

const priceLabel = computed(() => {
  const plan = currentPlan.value
  const amount = subscription.value.interval === 'yearly' ? plan.priceYearly : plan.priceMonthly
  return formatMoney(amount, subscription.value.interval)
})

function onPlanConfirm(planId: Parameters<typeof changePlan>[0], interval: Parameters<typeof changePlan>[1]) {
  changePlan(planId, interval)
  toast.add({
    title: 'Plan updated',
    description: `Switched to ${planId} (${interval}). Local stub until Stripe is wired.`,
    color: 'success'
  })
}

function onPaymentSave(method: Parameters<typeof setPaymentMethod>[0]) {
  setPaymentMethod(method)
  toast.add({
    title: 'Payment method saved',
    description: `${method.brand} ending in ${method.last4}`,
    color: 'success'
  })
}

function onDownload(invoice: Invoice) {
  toast.add({
    title: 'Download ready',
    description: `${invoice.number} PDF will come from the API later.`,
    color: 'neutral'
  })
}

function onConfirmCancel() {
  cancelAtPeriodEnd()
  cancelOpen.value = false
  toast.add({
    title: 'Cancellation scheduled',
    description: renewalLabel.value,
    color: 'neutral'
  })
}

function onResume() {
  resumeSubscription()
  toast.add({
    title: 'Subscription resumed',
    color: 'success'
  })
}

function statusColor(status: Invoice['status']) {
  if (status === 'paid') {
    return 'success' as const
  }
  if (status === 'failed') {
    return 'error' as const
  }
  return 'warning' as const
}
</script>

<template>
  <div class="flex flex-col gap-4 sm:gap-6">
    <LazyUAlert
      v-if="subscription.cancelAtPeriodEnd"
      color="warning"
      variant="subtle"
      icon="i-lucide-triangle-alert"
      title="Cancellation scheduled"
      :description="`Access continues until ${subscription.renewsAt ? formatDate(subscription.renewsAt) : 'period end'}.`"
    >
      <template #actions>
        <UButton
          label="Keep plan"
          color="neutral"
          variant="outline"
          size="sm"
          @click="onResume"
        />
      </template>
    </LazyUAlert>

    <div>
      <UPageCard
        title="Current plan"
        :description="`Billing for ${selected?.name}.`"
        variant="naked"
        orientation="horizontal"
        class="mb-4"
      >
        <UButton
          label="Change plan"
          color="neutral"
          class="w-fit lg:ms-auto"
          @click="changeOpen = true"
        />
      </UPageCard>

      <UPageCard variant="subtle">
        <div class="flex flex-col gap-5">
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div class="flex flex-col gap-1">
              <div class="flex items-center gap-2">
                <p class="text-highlighted text-lg font-semibold">
                  {{ currentPlan.name }}
                </p>
                <UBadge
                  :label="subscription.status"
                  color="neutral"
                  variant="outline"
                  class="capitalize"
                />
              </div>
              <p class="text-muted text-sm">
                {{ priceLabel }} · {{ renewalLabel }}
              </p>
              <p class="text-muted text-xs">
                {{ currentPlan.description }}
              </p>
            </div>

            <UButton
              v-if="subscription.planId !== 'free' && !subscription.cancelAtPeriodEnd"
              label="Cancel plan"
              color="error"
              variant="ghost"
              @click="cancelOpen = true"
            />
          </div>

          <USeparator />

          <div class="flex flex-col gap-4">
            <p class="text-highlighted text-sm font-medium">
              Usage
            </p>
            <div
              v-for="row in usageRows"
              :key="row.key"
              class="flex flex-col gap-1.5"
            >
              <div class="flex items-center justify-between text-sm">
                <span class="text-muted">{{ row.label }}</span>
                <span class="text-highlighted font-medium">{{ row.used }} / {{ row.max }}</span>
              </div>
              <UProgress
                :model-value="row.pct"
                size="sm"
                :color="row.pct >= 100 ? 'error' : row.pct >= 80 ? 'warning' : 'primary'"
              />
            </div>
          </div>

          <USeparator />

          <ul class="text-muted grid gap-1.5 text-sm sm:grid-cols-2">
            <li
              v-for="feature in currentPlan.features"
              :key="feature"
              class="flex items-center gap-2"
            >
              <UIcon
                name="i-lucide-check"
                class="text-primary size-4 shrink-0"
              />
              {{ feature }}
            </li>
          </ul>
        </div>
      </UPageCard>
    </div>

    <div>
      <UPageCard
        title="Payment method"
        description="Used for renewals and upgrades."
        variant="naked"
        orientation="horizontal"
        class="mb-4"
      >
        <UButton
          :label="subscription.paymentMethod ? 'Update' : 'Add card'"
          color="neutral"
          class="w-fit lg:ms-auto"
          @click="paymentOpen = true"
        />
      </UPageCard>

      <UPageCard variant="subtle">
        <div
          v-if="subscription.paymentMethod"
          class="flex flex-wrap items-center justify-between gap-3"
        >
          <div class="flex items-center gap-3">
            <div class="bg-elevated flex size-10 items-center justify-center rounded-md">
              <UIcon
                name="i-lucide-credit-card"
                class="size-5"
              />
            </div>
            <div class="text-sm">
              <p class="text-highlighted font-medium">
                {{ subscription.paymentMethod.brand }} ···· {{ subscription.paymentMethod.last4 }}
              </p>
              <p class="text-muted">
                Expires {{ String(subscription.paymentMethod.expMonth).padStart(2, '0') }}/{{ subscription.paymentMethod.expYear }}
              </p>
            </div>
          </div>
        </div>
        <p
          v-else
          class="text-muted text-sm"
        >
          No payment method on file. Required for paid plans.
        </p>
      </UPageCard>
    </div>

    <div>
      <UPageCard
        title="Invoices"
        description="History for this organization."
        variant="naked"
        class="mb-4"
      />

      <UPageCard
        variant="subtle"
        :ui="{ container: 'p-0 sm:p-0' }"
      >
        <div
          v-if="!subscription.invoices.length"
          class="text-muted p-4 text-sm sm:p-6"
        >
          No invoices yet.
        </div>
        <ul
          v-else
          class="divide-y divide-default"
        >
          <li
            v-for="invoice in subscription.invoices"
            :key="invoice.id"
            class="flex flex-wrap items-center justify-between gap-3 px-4 py-3 sm:px-6"
          >
            <div class="min-w-0 text-sm">
              <p class="text-highlighted font-medium">
                {{ invoice.number }}
              </p>
              <p class="text-muted">
                {{ formatDate(invoice.date) }} · {{ formatMoney(invoice.amount) }}
              </p>
            </div>
            <div class="flex items-center gap-2">
              <UBadge
                :label="invoice.status"
                :color="statusColor(invoice.status)"
                variant="subtle"
                class="capitalize"
              />
              <UButton
                label="PDF"
                color="neutral"
                variant="ghost"
                size="sm"
                icon="i-lucide-download"
                @click="onDownload(invoice)"
              />
            </div>
          </li>
        </ul>
      </UPageCard>
    </div>

    <LazyOrganizationChangePlanModal
      v-if="changeOpen"
      v-model:open="changeOpen"
      :current-plan-id="subscription.planId"
      :current-interval="subscription.interval"
      @confirm="onPlanConfirm"
    />

    <LazyOrganizationPaymentMethodModal
      v-if="paymentOpen"
      v-model:open="paymentOpen"
      :method="subscription.paymentMethod"
      @save="onPaymentSave"
    />

    <UModal
      v-if="cancelOpen"
      v-model:open="cancelOpen"
      title="Cancel plan?"
      description="You keep access until the end of the billing period. You can resume anytime before then."
    >
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton
            label="Keep plan"
            color="neutral"
            variant="ghost"
            @click="cancelOpen = false"
          />
          <UButton
            label="Cancel plan"
            color="error"
            @click="onConfirmCancel"
          />
        </div>
      </template>
    </UModal>
  </div>
</template>
