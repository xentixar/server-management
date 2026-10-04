<script setup lang="ts">
import { ApiError } from '~/utils/api'
import { Ability } from '~/utils/auth/abilities'
import {
  GIT_PROVIDERS,
  connectionStatusColor,
  providerLabel,
  type GitConnection,
  type GitProvider
} from '~/utils/gitConnections'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { can } = useGate()
const {
  connections,
  pending,
  fetchAll,
  authorize,
  disconnect
} = useGitConnections()

const canView = computed(() => can(Ability.GitConnectionsView))
const canManage = computed(() => can(Ability.GitConnectionsManage))

const connecting = ref<GitProvider | null>(null)
const disconnectingId = ref<string | null>(null)
const disconnectOpen = ref(false)
const connectionToDisconnect = ref<GitConnection | null>(null)

await fetchAll()

if (import.meta.client) {
  const connected = typeof route.query.connected === 'string' ? route.query.connected : ''
  const error = typeof route.query.error === 'string' ? route.query.error : ''
  if (connected || error) {
    if (connected) {
      toast.add({
        title: 'Connected',
        description: `${providerLabel(connected as GitProvider)} account linked.`,
        color: 'success'
      })
      await fetchAll(true)
    } else {
      toast.add({
        title: 'Connection failed',
        description: error.replaceAll('_', ' '),
        color: 'error'
      })
    }
    const nextQuery = { ...route.query }
    delete nextQuery.connected
    delete nextQuery.error
    await router.replace({ query: nextQuery })
  }
}

async function onConnect(provider: GitProvider) {
  if (!canManage.value) {
    return
  }
  connecting.value = provider
  try {
    const url = await authorize(provider)
    window.location.href = url
  } catch (error) {
    const err = error instanceof ApiError ? error : new ApiError('Failed to start OAuth', 500)
    toast.add({ title: err.message, color: 'error' })
    connecting.value = null
  }
}

function askDisconnect(connection: GitConnection) {
  if (!canManage.value) {
    return
  }
  connectionToDisconnect.value = connection
  disconnectOpen.value = true
}

async function onConfirmDisconnect() {
  const connection = connectionToDisconnect.value
  if (!connection || !canManage.value) {
    return
  }
  disconnectingId.value = connection.id
  try {
    await disconnect(connection.id)
    toast.add({
      title: 'Disconnected',
      description: `${providerLabel(connection.provider)} (${connection.accountLogin})`,
      color: 'neutral'
    })
    disconnectOpen.value = false
    connectionToDisconnect.value = null
  } catch (error) {
    const err = error instanceof ApiError ? error : new ApiError('Disconnect failed', 500)
    toast.add({ title: err.message, color: 'error' })
  } finally {
    disconnectingId.value = null
  }
}

function connectionsFor(provider: GitProvider) {
  return connections.value.filter(c => c.provider === provider)
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-col gap-1">
      <h2 class="text-highlighted text-xl font-semibold">
        Git providers
      </h2>
      <p class="text-muted text-sm">
        Connect GitHub, GitLab, or Bitbucket so this organization can list repositories for sites.
      </p>
    </div>

    <UAlert
      v-if="!canView"
      color="warning"
      variant="subtle"
      title="You do not have permission to view git connections."
      icon="i-lucide-lock"
    />

    <div
      v-else
      class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3"
    >
      <UPageCard
        v-for="provider in GIT_PROVIDERS"
        :key="provider.value"
        :title="provider.label"
        :description="provider.description"
        :icon="provider.icon"
        variant="subtle"
        class="min-w-0 self-start overflow-hidden"
      >
        <div class="flex min-w-0 flex-col gap-3">
          <p
            v-if="!connectionsFor(provider.value).length"
            class="text-muted text-sm"
          >
            No account connected.
          </p>
          <ul
            v-else
            class="flex min-w-0 flex-col gap-2"
          >
            <li
              v-for="connection in connectionsFor(provider.value)"
              :key="connection.id"
              class="flex min-w-0 flex-col gap-2 rounded-lg border border-default px-3 py-2"
            >
              <div class="flex min-w-0 items-start gap-3">
                <UAvatar
                  :src="connection.avatarUrl || undefined"
                  :alt="connection.accountLogin"
                  size="sm"
                  class="shrink-0"
                />
                <div class="min-w-0 flex-1 overflow-hidden">
                  <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                    <span class="truncate text-sm font-medium text-highlighted">
                      {{ connection.accountLogin }}
                    </span>
                    <UBadge
                      :color="connectionStatusColor(connection.status)"
                      variant="subtle"
                      size="sm"
                      class="shrink-0"
                    >
                      {{ connection.status }}
                    </UBadge>
                  </div>
                  <div class="text-muted truncate text-xs">
                    {{ connection.accountName || connection.externalAccountId }}
                  </div>
                </div>
              </div>
              <UButton
                v-if="canManage"
                label="Disconnect"
                icon="i-lucide-unlink"
                color="error"
                variant="soft"
                size="sm"
                block
                @click="askDisconnect(connection)"
              />
            </li>
          </ul>

          <UButton
            v-if="canManage"
            :label="`Connect ${provider.label}`"
            icon="i-lucide-link"
            color="primary"
            variant="soft"
            class="w-fit"
            :loading="connecting === provider.value"
            :disabled="pending"
            @click="onConnect(provider.value)"
          />
        </div>
      </UPageCard>
    </div>

    <UModal
      v-model:open="disconnectOpen"
      title="Disconnect account?"
      :description="connectionToDisconnect
        ? `Remove ${providerLabel(connectionToDisconnect.provider)} account ${connectionToDisconnect.accountLogin} from this organization.`
        : undefined"
    >
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton
            label="Cancel"
            color="neutral"
            variant="ghost"
            :disabled="!!disconnectingId"
            @click="disconnectOpen = false"
          />
          <UButton
            label="Disconnect"
            icon="i-lucide-unlink"
            color="error"
            :loading="!!disconnectingId"
            @click="onConfirmDisconnect"
          />
        </div>
      </template>
    </UModal>
  </div>
</template>
