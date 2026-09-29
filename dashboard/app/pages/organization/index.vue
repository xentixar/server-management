<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'

const { selected, updateSelected, slugify } = useOrganizations()
const toast = useToast()
const fileRef = ref<HTMLInputElement>()

const orgSchema = z.object({
  name: z.string().min(2, 'Too short'),
  slug: z.string().min(2, 'Too short').regex(/^[a-z0-9-]+$/, 'Lowercase letters, numbers, and hyphens only')
})

type OrgSchema = z.output<typeof orgSchema>

const state = reactive<Partial<OrgSchema>>({
  name: selected.value.name,
  slug: selected.value.slug
})

watch(selected, (org) => {
  state.name = org.name
  state.slug = org.slug
}, { deep: true })

watch(() => state.name, (name) => {
  if (name && state.slug === slugify(selected.value.name)) {
    state.slug = slugify(name)
  }
})

function onSubmit(event: FormSubmitEvent<OrgSchema>) {
  updateSelected({ name: event.data.name, slug: event.data.slug })
  toast.add({
    title: 'Saved locally',
    description: 'Organization settings are stubs until the API is wired.',
    color: 'neutral'
  })
}

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) {
    return
  }
  updateSelected({
    avatar: {
      src: URL.createObjectURL(input.files[0]!),
      alt: state.name || selected.value.name
    }
  })
}

function onDelete() {
  toast.add({
    title: 'Not wired yet',
    description: 'Org deletion will call the API later.',
    color: 'neutral'
  })
}
</script>

<template>
  <div class="flex flex-col gap-4 sm:gap-6">
    <UForm
      id="organization-general"
      :schema="orgSchema"
      :state="state"
      @submit="onSubmit"
    >
      <UPageCard
        title="General"
        description="Name and slug for the current organization."
        variant="naked"
        orientation="horizontal"
        class="mb-4"
      >
        <UButton
          form="organization-general"
          label="Save changes"
          color="neutral"
          type="submit"
          class="w-fit lg:ms-auto"
        />
      </UPageCard>

      <UPageCard variant="subtle">
        <UFormField
          name="name"
          label="Name"
          description="Shown in the sidebar switcher and billing."
          required
          class="flex max-sm:flex-col items-start justify-between gap-4"
        >
          <UInput
            v-model="state.name"
            autocomplete="off"
          />
        </UFormField>

        <USeparator />

        <UFormField
          name="slug"
          label="Slug"
          description="URL-safe identifier. Lowercase letters, numbers, hyphens."
          required
          class="flex max-sm:flex-col items-start justify-between gap-4"
        >
          <UInput
            v-model="state.slug"
            autocomplete="off"
          />
        </UFormField>

        <USeparator />

        <UFormField
          name="avatar"
          label="Avatar"
          description="JPG, GIF or PNG. 1MB max."
          class="flex max-sm:flex-col gap-4 sm:items-center sm:justify-between"
        >
          <div class="flex flex-wrap items-center gap-3">
            <UAvatar
              v-bind="selected.avatar"
              :alt="state.name || selected.name"
              size="lg"
            />
            <UButton
              label="Choose"
              color="neutral"
              @click="fileRef?.click()"
            />
            <input
              ref="fileRef"
              type="file"
              class="hidden"
              accept=".jpg, .jpeg, .png, .gif"
              @change="onFileChange"
            >
          </div>
        </UFormField>
      </UPageCard>
    </UForm>

    <UPageCard
      title="Danger zone"
      description="Delete this organization and all of its servers, sites, and members. This cannot be undone."
      class="bg-linear-to-tl from-error/10 from-5% to-default"
    >
      <template #footer>
        <UButton
          label="Delete organization"
          color="error"
          @click="onDelete"
        />
      </template>
    </UPageCard>
  </div>
</template>
