<script setup lang="ts">
import type { BillingInterval, PlanId } from '~/utils/billing'
import { PLAN_CATALOG, formatMoney, planById } from '~/utils/billing'

const open = defineModel<boolean>('open', { default: false })

const props = defineProps<{
  currentPlanId: PlanId
  currentInterval: BillingInterval
}>()

const emit = defineEmits<{
  confirm: [planId: PlanId, interval: BillingInterval]
}>()

const selectedPlanId = ref<PlanId>(props.currentPlanId)
const interval = ref<BillingInterval>(props.currentInterval)

watch(open, (isOpen) => {
  if (isOpen) {
    selectedPlanId.value = props.currentPlanId
    interval.value = props.currentInterval
  }
})

const selectedPlan = computed(() => planById(selectedPlanId.value))

const priceLabel = computed(() => {
  const plan = selectedPlan.value
  const amount = interval.value === 'yearly' ? plan.priceYearly : plan.priceMonthly
  return formatMoney(amount, interval.value)
})

const isCurrent = computed(() =>
  selectedPlanId.value === props.currentPlanId
  && interval.value === props.currentInterval
)

const actionLabel = computed(() => {
  if (isCurrent.value) {
    return 'Current plan'
  }
  const current = planById(props.currentPlanId)
  const next = selectedPlan.value
  const currentPrice = props.currentInterval === 'yearly' ? current.priceYearly : current.priceMonthly
  const nextPrice = interval.value === 'yearly' ? next.priceYearly : next.priceMonthly
  if (nextPrice > currentPrice) {
    return 'Upgrade'
  }
  if (nextPrice < currentPrice) {
    return 'Downgrade'
  }
  return 'Switch interval'
})

function onConfirm() {
  if (isCurrent.value) {
    return
  }
  emit('confirm', selectedPlanId.value, interval.value)
  open.value = false
}
</script>

<template>
  <UModal
    v-model:open="open"
    title="Change plan"
    description="Pick a plan for this organization. Charges are simulated until billing is wired."
    :ui="{ content: 'sm:max-w-2xl' }"
  >
    <template #body>
      <div class="flex flex-col gap-5">
        <div class="flex items-center justify-between gap-3">
          <p class="text-muted text-sm">
            Billing interval
          </p>
          <div class="flex gap-1">
            <UButton
              label="Monthly"
              size="sm"
              :color="interval === 'monthly' ? 'primary' : 'neutral'"
              :variant="interval === 'monthly' ? 'solid' : 'ghost'"
              @click="interval = 'monthly'"
            />
            <UButton
              label="Yearly"
              size="sm"
              :color="interval === 'yearly' ? 'primary' : 'neutral'"
              :variant="interval === 'yearly' ? 'solid' : 'ghost'"
              @click="interval = 'yearly'"
            />
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-3">
          <button
            v-for="plan in PLAN_CATALOG"
            :key="plan.id"
            type="button"
            class="border-default hover:border-primary/50 relative flex flex-col gap-3 rounded-lg border p-4 text-left transition-colors"
            :class="selectedPlanId === plan.id ? 'border-primary bg-primary/5 ring-primary/30 ring-1' : 'bg-default'"
            @click="selectedPlanId = plan.id"
          >
            <UBadge
              v-if="plan.highlighted"
              label="Popular"
              color="primary"
              variant="subtle"
              size="sm"
              class="absolute end-3 top-3"
            />
            <div>
              <p class="text-highlighted font-semibold">
                {{ plan.name }}
              </p>
              <p class="text-muted mt-1 text-xs">
                {{ plan.description }}
              </p>
            </div>
            <p class="text-highlighted text-lg font-semibold">
              {{ formatMoney(interval === 'yearly' ? plan.priceYearly : plan.priceMonthly, interval) }}
            </p>
            <ul class="text-muted flex flex-col gap-1 text-xs">
              <li
                v-for="limit in plan.limits"
                :key="limit.key"
              >
                {{ limit.max }} {{ limit.label.toLowerCase() }}
              </li>
            </ul>
          </button>
        </div>

        <ul class="text-muted flex flex-col gap-1.5 text-sm">
          <li
            v-for="feature in selectedPlan.features"
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
    </template>

    <template #footer>
      <div class="flex w-full items-center justify-between gap-3">
        <p class="text-muted text-sm">
          {{ selectedPlan.name }} · {{ priceLabel }}
        </p>
        <div class="flex gap-2">
          <UButton
            label="Cancel"
            color="neutral"
            variant="ghost"
            @click="open = false"
          />
          <UButton
            :label="actionLabel"
            :disabled="isCurrent"
            @click="onConfirm"
          />
        </div>
      </div>
    </template>
  </UModal>
</template>
