<script setup lang="ts">
import type { OrgMember } from '~/utils/members'

const open = defineModel<boolean>('open', { default: false })

const props = defineProps<{
  member: OrgMember | null
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

const expected = computed(() => props.member?.email ?? '')

const canConfirm = computed(() =>
  !!props.member && confirmText.value.trim().toLowerCase() === expected.value.toLowerCase()
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
    title="Remove member"
    :description="member ? `Remove ${member.name} from this organization. They lose access immediately.` : undefined"
  >
    <template #body>
      <div
        v-if="member"
        class="flex flex-col gap-4"
      >
        <UAlert
          color="error"
          variant="subtle"
          icon="i-lucide-triangle-alert"
          title="This cannot be undone"
          description="Type their email to confirm."
        />

        <UFormField
          label="Email"
          :description="`Type ${member.email} to confirm`"
        >
          <UInput
            v-model="confirmText"
            autocomplete="off"
            :placeholder="member.email"
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
          label="Remove member"
          color="error"
          :disabled="!canConfirm"
          @click="onConfirm"
        />
      </div>
    </template>
  </UModal>
</template>
