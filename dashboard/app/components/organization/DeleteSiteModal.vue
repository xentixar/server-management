<script setup lang="ts">
import type { Site } from '~/utils/sites'

const open = defineModel<boolean>('open', { default: false })

const props = defineProps<{
  site: Site | null
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

const expected = computed(() => props.site?.domain ?? '')

const canConfirm = computed(() =>
  !!props.site && confirmText.value.trim().toLowerCase() === expected.value.toLowerCase()
)

function onConfirm() {
  if (!canConfirm.value) {
    return
  }
  emit('confirm')
  open.value = false
}
</script>

<template>
  <UModal
    v-model:open="open"
    title="Delete site"
    :description="site ? `Remove ${site.domain} from this server.` : undefined"
  >
    <template #body>
      <div
        v-if="site"
        class="flex flex-col gap-4"
      >
        <UAlert
          color="error"
          variant="subtle"
          icon="i-lucide-triangle-alert"
          title="This cannot be undone"
          description="Type the domain to confirm."
        />

        <UFormField
          label="Domain"
          :description="`Type ${site.domain} to confirm`"
        >
          <UInput
            v-model="confirmText"
            autocomplete="off"
            :placeholder="site.domain"
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
          @click="open = false"
        />
        <UButton
          label="Delete site"
          color="error"
          :disabled="!canConfirm"
          @click="onConfirm"
        />
      </div>
    </template>
  </UModal>
</template>
