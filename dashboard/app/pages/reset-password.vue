<script setup lang="ts">
import { resetPasswordSchema, type ResetPasswordSchema } from '~/utils/auth.schema'

definePageMeta({ layout: 'auth' })

useSeoMeta({ title: 'Reset password - Outpost' })

const route = useRoute()
const toast = useToast()

const token = computed(() => {
  const raw = route.query.token
  return typeof raw === 'string' ? raw : ''
})

const state = reactive<Partial<ResetPasswordSchema>>({
  password: '',
  confirmPassword: ''
})

function onSubmit() {
  toast.add({
    title: 'Not wired yet',
    description: token.value
      ? `Token present (${token.value.slice(0, 8)}…). Auth API will land later.`
      : 'Auth API will land in a later phase.',
    color: 'neutral'
  })
}
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex flex-col gap-1">
        <h1 class="text-highlighted text-lg font-semibold">
          Reset password
        </h1>
        <p class="text-muted text-sm">
          Choose a new password for your account.
        </p>
      </div>
    </template>

    <UAlert
      v-if="!token"
      class="mb-4"
      color="warning"
      variant="subtle"
      icon="i-lucide-triangle-alert"
      title="Missing reset token"
      description="Open the link from your email, or request a new reset."
    />

    <UForm
      :schema="resetPasswordSchema"
      :state="state"
      class="flex flex-col gap-4"
      @submit="onSubmit"
    >
      <UFormField
        label="New password"
        name="password"
      >
        <UInput
          v-model="state.password"
          type="password"
          autocomplete="new-password"
          placeholder="••••••••"
          class="w-full"
        />
      </UFormField>

      <UFormField
        label="Confirm password"
        name="confirmPassword"
      >
        <UInput
          v-model="state.confirmPassword"
          type="password"
          autocomplete="new-password"
          placeholder="••••••••"
          class="w-full"
        />
      </UFormField>

      <UButton
        type="submit"
        label="Update password"
        block
        :disabled="!token"
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
