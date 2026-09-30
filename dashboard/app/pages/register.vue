<script setup lang="ts">
import { registerSchema, type RegisterSchema } from '~/utils/auth.schema'
import { ApiError } from '~/utils/api'

definePageMeta({ layout: 'auth' })

useSeoMeta({ title: 'Create account - Outpost' })

const toast = useToast()
const { register, pending } = useSession()

const state = reactive<Partial<RegisterSchema>>({
  name: '',
  email: '',
  password: '',
  confirmPassword: ''
})

async function onSubmit() {
  try {
    await register(state.name!, state.email!, state.password!)
    toast.add({
      title: 'Account created',
      description: 'Sign in with your new credentials.',
      color: 'success'
    })
    await navigateTo('/login')
  } catch (error) {
    const err = error instanceof ApiError ? error : new ApiError('Registration failed', 500)
    toast.add({
      title: 'Registration failed',
      description: err.data.email || err.message,
      color: 'error'
    })
  }
}
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex flex-col gap-1">
        <h1 class="text-highlighted text-lg font-semibold">
          Create account
        </h1>
        <p class="text-muted text-sm">
          Start managing servers with Outpost.
        </p>
      </div>
    </template>

    <UForm
      :schema="registerSchema"
      :state="state"
      class="flex flex-col gap-4"
      @submit="onSubmit"
    >
      <UFormField
        label="Display name"
        name="name"
      >
        <UInput
          v-model="state.name"
          autocomplete="name"
          placeholder="Ada Lovelace"
          class="w-full"
        />
      </UFormField>

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

      <UFormField
        label="Password"
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
        label="Create account"
        block
        :loading="pending"
      />
    </UForm>

    <template #footer>
      <p class="text-muted text-center text-sm">
        Already have an account?
        <NuxtLink
          to="/login"
          class="text-primary font-medium"
        >
          Sign in
        </NuxtLink>
      </p>
    </template>
  </UCard>
</template>
