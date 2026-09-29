<script setup lang="ts">
import { registerSchema, type RegisterSchema } from '~/utils/auth.schema'

definePageMeta({ layout: 'auth' })

useSeoMeta({ title: 'Create account - Outpost' })

const toast = useToast()

const state = reactive<Partial<RegisterSchema>>({
  name: '',
  email: '',
  password: '',
  confirmPassword: ''
})

function onSubmit() {
  toast.add({
    title: 'Not wired yet',
    description: 'Auth API will land in a later phase.',
    color: 'neutral'
  })
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
