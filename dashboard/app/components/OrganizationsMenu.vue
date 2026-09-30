<script setup lang="ts">
import * as z from 'zod'
import type { DropdownMenuItem, FormSubmitEvent } from '@nuxt/ui'
import { ApiError } from '~/utils/api'

defineProps<{
  collapsed?: boolean
}>()

const router = useRouter()
const toast = useToast()
const { organizations, selected, switchTo, create, orgHref, pending } = useOrganizations()

const createOpen = ref(false)
const createSchema = z.object({
  name: z.string().min(5, 'Must be at least 5 characters')
})
type CreateSchema = z.output<typeof createSchema>
const createState = reactive<Partial<CreateSchema>>({ name: '' })

async function onCreate(event: FormSubmitEvent<CreateSchema>) {
  try {
    const org = await create(event.data.name)
    createState.name = ''
    createOpen.value = false
    toast.add({
      title: 'Organization created',
      color: 'success'
    })
    await router.push(`/o/${org.slug}`)
  } catch (error) {
    const err = error instanceof ApiError ? error : new ApiError('Create failed', 500)
    toast.add({
      title: 'Could not create organization',
      description: err.message,
      color: 'error'
    })
  }
}

const items = computed<DropdownMenuItem[][]>(() => {
  return [organizations.value.map(org => ({
    label: org.name,
    avatar: { alt: org.name },
    type: 'checkbox' as const,
    checked: org.id === selected.value?.id,
    onSelect(e: Event) {
      e.preventDefault()
      switchTo(org.slug)
    }
  })), [{
    label: 'Create organization',
    icon: 'i-lucide-circle-plus',
    onSelect(e: Event) {
      e.preventDefault()
      createOpen.value = true
    }
  }, {
    label: 'Organization settings',
    icon: 'i-lucide-cog',
    onSelect() {
      router.push(orgHref('/organization'))
    }
  }]]
})
</script>

<template>
  <UDropdownMenu
    :items="items"
    :content="{ align: 'center', collisionPadding: 12 }"
    :ui="{ content: collapsed ? 'w-48' : 'w-(--reka-dropdown-menu-trigger-width)' }"
  >
    <UButton
      v-bind="{
        label: collapsed ? undefined : selected?.name,
        avatar: selected ? { alt: selected.name } : undefined,
        trailingIcon: collapsed ? undefined : 'i-lucide-chevrons-up-down'
      }"
      color="neutral"
      variant="ghost"
      block
      :square="collapsed"
      class="data-[state=open]:bg-elevated"
      :class="[!collapsed && 'py-2']"
      :ui="{
        trailingIcon: 'text-dimmed'
      }"
    />
  </UDropdownMenu>

  <UModal
    v-model:open="createOpen"
    title="Create organization"
  >
    <template #body>
      <UForm
        :schema="createSchema"
        :state="createState"
        class="flex flex-col gap-4"
        @submit="onCreate"
      >
        <UFormField
          label="Name"
          name="name"
        >
          <UInput
            v-model="createState.name"
            placeholder="Acme Inc"
            autofocus
            class="w-full"
          />
        </UFormField>

        <div class="flex justify-end gap-2">
          <UButton
            label="Cancel"
            color="neutral"
            variant="ghost"
            @click="createOpen = false"
          />
          <UButton
            type="submit"
            label="Create"
            :loading="pending"
          />
        </div>
      </UForm>
    </template>
  </UModal>
</template>
