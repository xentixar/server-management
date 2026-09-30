<script setup lang="ts">
import { ApiError } from '~/utils/api'

definePageMeta({ layout: false })

const toast = useToast()
const { organizations, resolveLandingSlug, create, pending } = useOrganizations()

const slug = resolveLandingSlug()
if (slug) {
  await navigateTo(`/o/${slug}`, { replace: true })
}

const name = ref('')

async function onCreate() {
  const trimmed = name.value.trim()
  if (trimmed.length < 5) {
    toast.add({
      title: 'Name too short',
      description: 'Organization name must be at least 5 characters.',
      color: 'error'
    })
    return
  }

  try {
    const org = await create(trimmed)
    await navigateTo(`/o/${org.slug}`, { replace: true })
  } catch (error) {
    const err = error instanceof ApiError ? error : new ApiError('Create failed', 500)
    toast.add({
      title: 'Could not create organization',
      description: err.message,
      color: 'error'
    })
  }
}
</script>

<template>
  <div
    v-if="!organizations.length"
    class="flex min-h-svh items-center justify-center p-6"
  >
    <UCard class="w-full max-w-md">
      <template #header>
        <div class="flex flex-col gap-1">
          <h1 class="text-highlighted text-lg font-semibold">
            Create an organization
          </h1>
          <p class="text-muted text-sm">
            You need an organization to manage servers.
          </p>
        </div>
      </template>

      <form
        class="flex flex-col gap-4"
        @submit.prevent="onCreate"
      >
        <UFormField label="Name">
          <UInput
            v-model="name"
            placeholder="Acme Inc"
            autofocus
            class="w-full"
          />
        </UFormField>
        <UButton
          type="submit"
          label="Create organization"
          block
          :loading="pending"
        />
      </form>
    </UCard>
  </div>
</template>
