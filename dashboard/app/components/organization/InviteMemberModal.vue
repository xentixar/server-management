<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'
import type { MemberRole } from '~/utils/members'
import { MEMBER_ROLES, ROLE_OPTIONS } from '~/utils/members'

const open = defineModel<boolean>('open', { default: false })

const emit = defineEmits<{
  invite: [email: string, role: MemberRole]
}>()

const inviteRoles = ROLE_OPTIONS.filter(r => r.value !== 'owner')

const schema = z.object({
  email: z.string().email('Enter a valid email'),
  role: z.enum(['admin', 'member', 'viewer'])
})

type Schema = z.output<typeof schema>

const state = reactive<Partial<Schema>>({
  email: '',
  role: 'member'
})

watch(open, (isOpen) => {
  if (isOpen) {
    state.email = ''
    state.role = 'member'
  }
})

const roleHint = computed(() =>
  MEMBER_ROLES.find(r => r.value === state.role)?.description ?? ''
)

function onSubmit(event: FormSubmitEvent<Schema>) {
  emit('invite', event.data.email, event.data.role)
}
</script>

<template>
  <UModal
    v-model:open="open"
    title="Invite member"
    description="They get an email with a link to join this organization."
  >
    <template #body>
      <UForm
        :schema="schema"
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
            placeholder="teammate@example.com"
            autofocus
            class="w-full"
          />
        </UFormField>

        <UFormField
          label="Role"
          name="role"
          :description="roleHint"
        >
          <USelect
            v-model="state.role"
            :items="inviteRoles"
            value-key="value"
            class="w-full"
          />
        </UFormField>

        <div class="flex justify-end gap-2">
          <UButton
            label="Cancel"
            color="neutral"
            variant="ghost"
            @click="open = false"
          />
          <UButton
            type="submit"
            label="Send invite"
          />
        </div>
      </UForm>
    </template>
  </UModal>
</template>
