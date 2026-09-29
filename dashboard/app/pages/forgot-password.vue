<script setup lang="ts">
import { forgotPasswordSchema, type ForgotPasswordSchema } from '~/utils/auth.schema'

definePageMeta({ layout: 'auth' })

useSeoMeta({ title: 'Forgot password - Outpost' })

const sent = ref(false)

const state = reactive<Partial<ForgotPasswordSchema>>({
  email: ''
})

function onSubmit() {
  // ponytail: no API yet; flip local success UI only
  sent.value = true
}
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex flex-col gap-1">
        <h1 class="text-highlighted text-lg font-semibold">
          Forgot password
        </h1>
        <p class="text-muted text-sm">
          {{ sent ? 'Check your inbox for a reset link.' : 'Enter your email and we will send a reset link.' }}
        </p>
      </div>
    </template>

    <div
      v-if="sent"
      class="flex flex-col gap-4"
    >
      <UAlert
        color="success"
        variant="subtle"
        icon="i-lucide-mail-check"
        title="Email sent"
        :description="`If an account exists for ${state.email}, you will receive reset instructions shortly.`"
      />

      <UButton
        to="/login"
        label="Back to sign in"
        color="neutral"
        variant="outline"
        block
      />
    </div>

    <UForm
      v-else
      :schema="forgotPasswordSchema"
      :state="state"
      class="flex flex-col gap-4"
      @submit="onSubmit"
    >
      <UFormField
        label="Email"
        name="email"
      >
        <UInput
          v-model="state.email"
          type="email"
          autocomplete="email"
          placeholder="you@example.com"
          class="w-full"
        />
      </UFormField>

      <UButton
        type="submit"
        label="Send reset link"
        block
      />

      <UButton
        to="/login"
        label="Back to sign in"
        color="neutral"
        variant="ghost"
        block
      />
    </UForm>
  </UCard>
</template>
