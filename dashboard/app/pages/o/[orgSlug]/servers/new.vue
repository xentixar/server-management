<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'
import type { SshAuthMethod } from '~/utils/servers'
import { ApiError } from '~/utils/api'

const { create, testConnection, atCapacity, serverLimit } = useServers()
const { orgHref } = useOrganizations()
const { currentPlan } = useBilling()
const router = useRouter()
const toast = useToast()
const submitting = ref(false)
const testing = ref(false)
const privateKey = ref('')

const schema = z.object({
  hostname: z.string().min(2, 'Too short'),
  ip: z.string().min(1, 'Required'),
  sshPort: z.coerce.number().int().min(1).max(65535),
  sshUser: z.string().min(1, 'Required'),
  authMethod: z.enum(['key', 'password']),
  passphrase: z.string().optional(),
  password: z.string().optional(),
  region: z.string().optional(),
  labels: z.string().optional(),
  notes: z.string().optional()
}).superRefine((data, ctx) => {
  if (data.authMethod === 'password' && !data.password?.trim()) {
    ctx.addIssue({
      code: 'custom',
      path: ['password'],
      message: 'Password is required for this auth method'
    })
  }
})

type Schema = z.output<typeof schema>

const state = reactive<Partial<Schema>>({
  hostname: '',
  ip: '',
  sshPort: 22,
  sshUser: 'root',
  authMethod: 'key',
  passphrase: '',
  password: '',
  region: '',
  labels: '',
  notes: ''
})

const keyFileName = ref('')
const keyUploaded = ref(false)
const keyInput = ref<HTMLInputElement>()

const authItems = [{
  label: 'SSH private key',
  value: 'key' as SshAuthMethod
}, {
  label: 'Password',
  value: 'password' as SshAuthMethod
}]

async function onKeyFile(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    keyFileName.value = ''
    keyUploaded.value = false
    privateKey.value = ''
    return
  }
  keyFileName.value = file.name
  privateKey.value = await file.text()
  keyUploaded.value = !!privateKey.value.trim()
}

async function onSubmit(event: FormSubmitEvent<Schema>) {
  if (atCapacity.value) {
    toast.add({
      title: 'Plan limit reached',
      description: `${currentPlan.value.name} allows ${serverLimit.value} servers.`,
      color: 'error'
    })
    return
  }

  if (event.data.authMethod === 'key' && !keyUploaded.value) {
    toast.add({
      title: 'Private key required',
      description: 'Choose a key file, or switch to password auth.',
      color: 'error'
    })
    return
  }

  submitting.value = true
  try {
    const server = await create(buildInput(event.data))

    if (server.bootstrap && !server.bootstrap.ok) {
      toast.add({
        title: 'Server saved, agent install failed',
        description: server.bootstrap.message,
        color: 'warning'
      })
    } else {
      toast.add({
        title: 'Server added',
        description: server.bootstrap?.message || `${server.hostname} agent install started.`,
        color: 'success'
      })
    }

    await router.push(orgHref(`/servers/${server.id}`))
  } catch (error) {
    const detail = error instanceof ApiError && error.data?.message
      ? String(error.data.message)
      : error instanceof Error ? error.message : 'Failed to create server'
    toast.add({
      title: error instanceof ApiError ? error.message : 'Failed to create server',
      description: detail !== (error instanceof Error ? error.message : '') ? detail : undefined,
      color: 'error'
    })
  } finally {
    submitting.value = false
  }
}

function buildInput(data: Schema): Parameters<typeof create>[0] {
  const labels = (data.labels || '')
    .split(',')
    .map(s => s.trim())
    .filter(Boolean)

  return {
    hostname: data.hostname,
    ip: data.ip,
    sshPort: data.sshPort,
    sshUser: data.sshUser,
    authMethod: data.authMethod,
    privateKey: data.authMethod === 'key' ? privateKey.value : undefined,
    passphrase: data.passphrase,
    password: data.password,
    region: data.region,
    labels,
    notes: data.notes
  }
}

async function onTestConnection() {
  if (!state.ip?.trim() || !state.sshUser?.trim()) {
    toast.add({ title: 'IP and SSH user are required', color: 'error' })
    return
  }
  if (state.authMethod === 'key' && !keyUploaded.value) {
    toast.add({ title: 'Private key required', color: 'error' })
    return
  }
  if (state.authMethod === 'password' && !state.password?.trim()) {
    toast.add({ title: 'Password required', color: 'error' })
    return
  }

  testing.value = true
  try {
    const result = await testConnection(buildInput({
      hostname: state.hostname || 'test',
      ip: state.ip!,
      sshPort: state.sshPort || 22,
      sshUser: state.sshUser!,
      authMethod: state.authMethod || 'key',
      passphrase: state.passphrase,
      password: state.password,
      region: state.region,
      labels: state.labels,
      notes: state.notes
    }))
    toast.add({
      title: 'SSH connection ok',
      description: result.message,
      color: 'success'
    })
  } catch (error) {
    const detail = error instanceof ApiError && error.data?.message
      ? String(error.data.message)
      : error instanceof Error ? error.message : 'SSH failed'
    toast.add({
      title: 'SSH connection failed',
      description: detail,
      color: 'error'
    })
  } finally {
    testing.value = false
  }
}

function onTerraformSoon() {
  toast.add({
    title: 'Coming later',
    description: 'Terraform provisioning lands in a later phase.',
    color: 'neutral'
  })
}
</script>

<template>
  <UDashboardPanel id="servers-new">
    <template #header>
      <UDashboardNavbar title="Add server">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #right>
          <UButton
            :to="orgHref('/servers')"
            label="Cancel"
            color="neutral"
            variant="ghost"
          />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="mx-auto flex w-full max-w-2xl flex-col gap-6 p-4 sm:p-6 lg:p-8">
        <div class="flex flex-col gap-1">
          <h1 class="text-highlighted text-xl font-semibold">
            Connect an existing machine
          </h1>
          <p class="text-muted text-sm">
            We SSH into the host, verify access, install the agent binary, and start it with your API URL + server token.
          </p>
        </div>

        <UAlert
          v-if="atCapacity"
          color="error"
          variant="subtle"
          icon="i-lucide-circle-alert"
          title="Plan limit reached"
          :description="`${currentPlan.name} allows ${serverLimit} servers. Upgrade billing to add more.`"
        />

        <UForm
          :schema="schema"
          :state="state"
          class="flex flex-col gap-5"
          @submit="onSubmit"
        >
          <UPageCard
            title="Machine"
            description="How Outpost reaches this server."
            variant="subtle"
          >
            <div class="flex flex-col gap-4">
              <UFormField
                label="Hostname"
                name="hostname"
                required
              >
                <UInput
                  v-model="state.hostname"
                  placeholder="edge-01.example.com"
                  class="w-full"
                />
              </UFormField>

              <UFormField
                label="IP or host"
                name="ip"
                required
              >
                <UInput
                  v-model="state.ip"
                  placeholder="203.0.113.10"
                  class="w-full"
                />
              </UFormField>

              <div class="grid gap-4 sm:grid-cols-2">
                <UFormField
                  label="SSH port"
                  name="sshPort"
                  required
                >
                  <UInput
                    v-model="state.sshPort"
                    type="number"
                    class="w-full"
                  />
                </UFormField>

                <UFormField
                  label="SSH user"
                  name="sshUser"
                  required
                >
                  <UInput
                    v-model="state.sshUser"
                    placeholder="root"
                    class="w-full"
                  />
                </UFormField>
              </div>

              <UFormField
                label="Region"
                name="region"
                description="Optional label for your own sorting."
              >
                <UInput
                  v-model="state.region"
                  placeholder="nyc1"
                  class="w-full"
                />
              </UFormField>
            </div>
          </UPageCard>

          <UPageCard
            title="Authentication"
            description="Prefer SSH keys. Passwords are discouraged."
            variant="subtle"
          >
            <div class="flex flex-col gap-4">
              <UFormField
                label="Method"
                name="authMethod"
              >
                <USelect
                  v-model="state.authMethod"
                  :items="authItems"
                  value-key="value"
                  class="w-full"
                />
              </UFormField>

              <template v-if="state.authMethod === 'key'">
                <UFormField
                  label="Private key"
                  description="File is not uploaded anywhere yet; we only record that you chose one."
                >
                  <div class="flex flex-wrap items-center gap-3">
                    <UButton
                      label="Choose file"
                      color="neutral"
                      @click="keyInput?.click()"
                    />
                    <span class="text-muted text-sm">
                      {{ keyFileName || 'No file chosen' }}
                    </span>
                    <input
                      ref="keyInput"
                      type="file"
                      class="hidden"
                      @change="onKeyFile"
                    >
                  </div>
                </UFormField>

                <UFormField
                  label="Passphrase"
                  name="passphrase"
                  description="Optional."
                >
                  <UInput
                    v-model="state.passphrase"
                    type="password"
                    autocomplete="off"
                    class="w-full"
                  />
                </UFormField>
              </template>

              <template v-else>
                <UAlert
                  color="warning"
                  variant="subtle"
                  icon="i-lucide-triangle-alert"
                  title="Password auth is discouraged"
                  description="Prefer key-based SSH. Passwords are riskier to store and rotate."
                />
                <UFormField
                  label="Password"
                  name="password"
                  required
                >
                  <UInput
                    v-model="state.password"
                    type="password"
                    autocomplete="new-password"
                    class="w-full"
                  />
                </UFormField>
              </template>
            </div>
          </UPageCard>

          <UPageCard
            title="Extras"
            description="Optional metadata."
            variant="subtle"
          >
            <div class="flex flex-col gap-4">
              <UFormField
                label="Labels"
                name="labels"
                description="Comma-separated."
              >
                <UInput
                  v-model="state.labels"
                  placeholder="edge, production"
                  class="w-full"
                />
              </UFormField>

              <UFormField
                label="Notes"
                name="notes"
              >
                <UTextarea
                  v-model="state.notes"
                  :rows="3"
                  autoresize
                  class="w-full"
                />
              </UFormField>
            </div>
          </UPageCard>

          <div class="flex flex-wrap items-center justify-between gap-3">
            <UButton
              label="Test connection"
              color="neutral"
              variant="outline"
              :loading="testing"
              :disabled="submitting"
              @click="onTestConnection"
            />
            <div class="flex gap-2">
              <UButton
                :to="orgHref('/servers')"
                label="Cancel"
                color="neutral"
                variant="ghost"
              />
              <UButton
                type="submit"
                label="Add server"
                :disabled="atCapacity"
                :loading="submitting"
              />
            </div>
          </div>
        </UForm>

        <p class="text-muted text-xs">
          Prefer cloud provision?
          <button
            type="button"
            class="text-primary font-medium"
            @click="onTerraformSoon"
          >
            Terraform wizard
          </button>
          comes later.
        </p>
      </div>
    </template>
  </UDashboardPanel>
</template>
