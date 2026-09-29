<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'
import type { PaymentMethod } from '~/utils/billing'

const open = defineModel<boolean>('open', { default: false })

defineProps<{
  method: PaymentMethod | null
}>()

const emit = defineEmits<{
  save: [method: PaymentMethod]
}>()

const schema = z.object({
  name: z.string().min(2, 'Too short'),
  number: z.string().min(12, 'Enter a card number').max(19),
  exp: z.string().regex(/^\d{2}\/\d{2}$/, 'Use MM/YY'),
  cvc: z.string().min(3, 'Invalid CVC').max(4)
})

type Schema = z.output<typeof schema>

const state = reactive<Partial<Schema>>({
  name: '',
  number: '',
  exp: '',
  cvc: ''
})

watch(open, (isOpen) => {
  if (isOpen) {
    state.name = ''
    state.number = ''
    state.exp = ''
    state.cvc = ''
  }
})

function onSubmit(event: FormSubmitEvent<Schema>) {
  const digits = event.data.number.replace(/\s+/g, '')
  const [mm, yy] = event.data.exp.split('/')
  emit('save', {
    brand: digits.startsWith('4') ? 'Visa' : 'Card',
    last4: digits.slice(-4),
    expMonth: Number(mm),
    expYear: 2000 + Number(yy)
  })
  open.value = false
}
</script>

<template>
  <UModal
    v-model:open="open"
    title="Payment method"
    description="Card details stay in the browser for now. Stripe will replace this later."
  >
    <template #body>
      <UForm
        :schema="schema"
        :state="state"
        class="flex flex-col gap-4"
        @submit="onSubmit"
      >
        <UFormField
          label="Name on card"
          name="name"
        >
          <UInput
            v-model="state.name"
            autocomplete="cc-name"
            class="w-full"
          />
        </UFormField>

        <UFormField
          label="Card number"
          name="number"
        >
          <UInput
            v-model="state.number"
            inputmode="numeric"
            autocomplete="cc-number"
            placeholder="4242 4242 4242 4242"
            class="w-full"
          />
        </UFormField>

        <div class="grid grid-cols-2 gap-3">
          <UFormField
            label="Expiry"
            name="exp"
          >
            <UInput
              v-model="state.exp"
              autocomplete="cc-exp"
              placeholder="MM/YY"
              class="w-full"
            />
          </UFormField>

          <UFormField
            label="CVC"
            name="cvc"
          >
            <UInput
              v-model="state.cvc"
              inputmode="numeric"
              autocomplete="cc-csc"
              placeholder="123"
              class="w-full"
            />
          </UFormField>
        </div>

        <div class="flex justify-end gap-2">
          <UButton
            label="Cancel"
            color="neutral"
            variant="ghost"
            @click="open = false"
          />
          <UButton
            type="submit"
            label="Save card"
          />
        </div>
      </UForm>
    </template>
  </UModal>
</template>
