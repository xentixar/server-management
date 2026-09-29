<script setup lang="ts">
import { loginSchema, type LoginSchema } from '~/utils/auth.schema'

definePageMeta({ layout: 'auth' })

useSeoMeta({ title: 'Sign in - Outpost' })

const toast = useToast()

const state = reactive<Partial<LoginSchema>>({
  email: '',
  password: ''
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
          Sign in
        </h1>
        <p class="text-muted text-sm">
          Welcome back to Outpost.
        </p>
      </div>
    </template>

    <UForm
      :schema="loginSchema"
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

      <UFormField
        label="Password"
        name="password"
      >
        <UInput
          v-model="state.password"
          type="password"
          autocomplete="current-password"
          placeholder="••••••••"
          class="w-full"
        />
      </UFormField>

      <div class="flex justify-end">
        <UButton
          to="/forgot-password"
          label="Forgot password?"
          color="neutral"
          variant="link"
          class="px-0"
        />
      </div>

      <UButton
        type="submit"
        label="Sign in"
        block
      />
    </UForm>

    <template #footer>
      <p class="text-muted text-center text-sm">
        No account?
        <NuxtLink
          to="/register"
          class="text-primary font-medium"
        >
          Create one
        </NuxtLink>
      </p>
    </template>
  </UCard>
</template>
