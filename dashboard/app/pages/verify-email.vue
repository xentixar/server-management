<script setup lang="ts">
definePageMeta({ layout: 'auth' })

useSeoMeta({ title: 'Verify email - Outpost' })

const route = useRoute()

const token = computed(() => {
  const raw = route.query.token
  return typeof raw === 'string' ? raw : ''
})

const confirmed = computed(() => !!token.value)
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex flex-col gap-1">
        <h1 class="text-highlighted text-lg font-semibold">
          {{ confirmed ? 'Email verified' : 'Verify your email' }}
        </h1>
        <p class="text-muted text-sm">
          {{ confirmed
            ? 'Your email address has been confirmed.'
            : 'We sent a verification link to your inbox.' }}
        </p>
      </div>
    </template>

    <UAlert
      v-if="confirmed"
      color="success"
      variant="subtle"
      icon="i-lucide-badge-check"
      title="You're all set"
      description="You can continue to Outpost. Auth API wiring comes later."
    />

    <UAlert
      v-else
      color="neutral"
      variant="subtle"
      icon="i-lucide-mail"
      title="Check your inbox"
      description="Open the verification link in the email we sent. If you already have a token, append ?token=… to this URL."
    />

    <template #footer>
      <UButton
        to="/login"
        :label="confirmed ? 'Continue to sign in' : 'Back to sign in'"
        block
      />
    </template>
  </UCard>
</template>
