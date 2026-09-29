<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'
import { STACK_OPTIONS, SITE_STACKS, defaultWebRoot, type SiteStack } from '~/utils/sites'

const route = useRoute()
const router = useRouter()
const { orgHref } = useOrganizations()
const { create } = useSites()
const toast = useToast()

const serverId = computed(() => String(route.params.serverId || ''))

const schema = z.object({
  name: z.string().min(2, 'Too short'),
  domain: z.string().min(3, 'Enter a domain'),
  stack: z.enum(['static', 'php', 'laravel', 'node', 'nextjs', 'nuxt', 'custom']),
  webRoot: z.string().min(1, 'Required'),
  gitRepo: z.string().optional(),
  gitBranch: z.string().optional()
})

type Schema = z.output<typeof schema>

const state = reactive<Partial<Schema>>({
  name: '',
  domain: '',
  stack: 'static',
  webRoot: 'public',
  gitRepo: '',
  gitBranch: 'main'
})

watch(() => state.stack, (stack) => {
  if (stack) {
    state.webRoot = defaultWebRoot(stack as SiteStack)
  }
})

const stackHint = computed(() =>
  SITE_STACKS.find(s => s.value === state.stack)?.description ?? ''
)

function onSubmit(event: FormSubmitEvent<Schema>) {
  const result = create({
    serverId: serverId.value,
    name: event.data.name,
    domain: event.data.domain,
    stack: event.data.stack,
    webRoot: event.data.webRoot,
    gitRepo: event.data.gitRepo,
    gitBranch: event.data.gitBranch
  })

  if (!result.ok) {
    toast.add({ title: result.error, color: 'error' })
    return
  }

  toast.add({
    title: 'Site created',
    description: `${result.site.domain} is ready for deploys (stub).`,
    color: 'success'
  })

  router.push(orgHref(`/servers/${serverId.value}/sites`))
}
</script>

<template>
  <div class="mx-auto flex w-full max-w-2xl flex-col gap-6">
    <div class="flex flex-col gap-1">
      <h2 class="text-highlighted text-xl font-semibold">
        Add site
      </h2>
      <p class="text-muted text-sm">
        Define the domain and stack. Nginx and deploys wire up when the API lands.
      </p>
    </div>

    <UForm
      :schema="schema"
      :state="state"
      class="flex flex-col gap-5"
      @submit="onSubmit"
    >
      <UPageCard
        title="Identity"
        description="How this site is addressed."
        variant="subtle"
      >
        <div class="flex flex-col gap-4">
          <UFormField
            label="Name"
            name="name"
            required
          >
            <UInput
              v-model="state.name"
              placeholder="Shop"
              class="w-full"
            />
          </UFormField>

          <UFormField
            label="Domain"
            name="domain"
            required
          >
            <UInput
              v-model="state.domain"
              placeholder="shop.example.com"
              class="w-full"
            />
          </UFormField>
        </div>
      </UPageCard>

      <UPageCard
        title="Stack"
        description="Runtime and document root."
        variant="subtle"
      >
        <div class="flex flex-col gap-4">
          <UFormField
            label="Type"
            name="stack"
            :description="stackHint"
            required
          >
            <USelect
              v-model="state.stack"
              :items="STACK_OPTIONS"
              value-key="value"
              class="w-full"
            />
          </UFormField>

          <UFormField
            label="Web root"
            name="webRoot"
            required
          >
            <UInput
              v-model="state.webRoot"
              placeholder="public"
              class="w-full"
            />
          </UFormField>
        </div>
      </UPageCard>

      <UPageCard
        title="Git"
        description="Optional. Used for deploys later."
        variant="subtle"
      >
        <div class="flex flex-col gap-4">
          <UFormField
            label="Repository"
            name="gitRepo"
          >
            <UInput
              v-model="state.gitRepo"
              placeholder="git@github.com:org/repo.git"
              class="w-full"
            />
          </UFormField>

          <UFormField
            label="Branch"
            name="gitBranch"
          >
            <UInput
              v-model="state.gitBranch"
              placeholder="main"
              class="w-full"
            />
          </UFormField>
        </div>
      </UPageCard>

      <div class="flex justify-end gap-2">
        <UButton
          :to="orgHref(`/servers/${serverId}/sites`)"
          label="Cancel"
          color="neutral"
          variant="ghost"
        />
        <UButton
          type="submit"
          label="Create site"
        />
      </div>
    </UForm>
  </div>
</template>
