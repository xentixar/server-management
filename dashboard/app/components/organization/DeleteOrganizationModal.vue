<script setup lang="ts">
import type { Organization } from '~/types'

const open = defineModel<boolean>('open', { default: false })

const props = defineProps<{
  organization: Organization | null
  loading?: boolean
}>()

const emit = defineEmits<{
  confirm: []
}>()

const confirmText = ref('')

watch(open, (isOpen) => {
  if (isOpen) {
    confirmText.value = ''
  }
})

const expected = computed(() => props.organization?.name ?? '')

const canConfirm = computed(() =>
  !!props.organization && confirmText.value.trim().toLowerCase() === expected.value.toLowerCase()
)

function onConfirm() {
  if (!canConfirm.value || props.loading) {
    return
  }
  emit('confirm')
}
</script>

<template>
  <UModal
    v-model:open="open"
    title="Delete organization"
    :description="organization ? `Permanently delete ${organization.name}.` : undefined"
  >
    <template #body>
      <div
        v-if="organization"
        class="flex flex-col gap-4"
      >
        <UAlert
          color="error"
          variant="subtle"
          icon="i-lucide-triangle-alert"
          title="This cannot be undone"
          description="Servers, sites, and members for this organization will be removed. Type the organization name to confirm."
        />

        <UFormField
          label="Organization name"
          :description="`Type ${organization.name} to confirm`"
        >
          <UInput
            v-model="confirmText"
            autocomplete="off"
            :placeholder="organization.name"
            autofocus
            class="w-full"
          />
        </UFormField>
      </div>
    </template>

    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton
          label="Cancel"
          color="neutral"
          variant="ghost"
          :disabled="loading"
          @click="open = false"
        />
        <UButton
          label="Delete organization"
          color="error"
          :disabled="!canConfirm"
          :loading="loading"
          @click="onConfirm"
        />
      </div>
    </template>
  </UModal>
</template>
