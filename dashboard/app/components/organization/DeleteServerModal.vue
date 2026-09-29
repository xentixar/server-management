<script setup lang="ts">
import type { Server } from '~/utils/servers'

const open = defineModel<boolean>('open', { default: false })

const props = defineProps<{
  server: Server | null
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

const expected = computed(() => props.server?.hostname ?? '')

const canConfirm = computed(() =>
  !!props.server && confirmText.value.trim().toLowerCase() === expected.value.toLowerCase()
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
    title="Delete server"
    :description="server ? `Remove ${server.hostname} from this organization.` : undefined"
  >
    <template #body>
      <div
        v-if="server"
        class="flex flex-col gap-4"
      >
        <UAlert
          color="error"
          variant="subtle"
          icon="i-lucide-triangle-alert"
          title="This cannot be undone"
          description="Sites on this server will be orphaned in the stub UI. Type the hostname to confirm."
        />

        <UFormField
          label="Hostname"
          :description="`Type ${server.hostname} to confirm`"
        >
          <UInput
            v-model="confirmText"
            autocomplete="off"
            :placeholder="server.hostname"
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
          label="Delete server"
          color="error"
          :disabled="!canConfirm"
          @click="onConfirm"
        />
      </div>
    </template>
  </UModal>
</template>
