<script setup lang="ts">
import { ApiError } from '~/utils/api'

definePageMeta({ layout: 'auth' })

useSeoMeta({ title: 'Accept invite - Outpost' })

const route = useRoute()
const toast = useToast()
const { user, isAuthenticated, signOutLocal } = useSession()
const { acceptInvite } = useMembers()
const { fetchOrganizations, organizations } = useOrganizations()

const raw = route.params.orgId
const orgId = Number(Array.isArray(raw) ? raw[0] : raw)
const inviteEmail = typeof route.query.email === 'string' ? route.query.email : ''

const invalid = !(Number.isFinite(orgId) && orgId > 0)
const accepting = ref(false)
const pageError = ref<string | null>(invalid ? 'This invite link is invalid.' : null)

const authQuery = computed(() => ({
  redirect: `/invites/${orgId}`,
  ...(inviteEmail ? { email: inviteEmail } : {})
}))

async function finishAccept() {
  if (invalid) {
    return
  }

  accepting.value = true
  pageError.value = null
  try {
    await acceptInvite(orgId)
    clearOrganizations()
    await fetchOrganizations()
    const org = organizations.value.find(o => o.id === orgId)
    toast.add({
      title: 'Invite accepted',
      description: org ? `Welcome to ${org.name}.` : undefined,
      color: 'success'
    })
    await navigateTo(org ? `/o/${org.slug}` : '/', { replace: true })
  } catch (error) {
    const err = error instanceof ApiError ? error : new ApiError('Could not accept invite', 500)
    pageError.value = err.message
  } finally {
    accepting.value = false
  }
}

if (!invalid && isAuthenticated.value) {
  await finishAccept()
}
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex flex-col gap-1">
        <h1 class="text-highlighted text-lg font-semibold">
          Organization invite
        </h1>
        <p class="text-muted text-sm">
          <template v-if="isAuthenticated">
            Signed in as {{ user?.email }}. Joining the organization…
          </template>
          <template v-else>
            Sign in or create an account with the invited email to join.
          </template>
        </p>
      </div>
    </template>

    <div class="flex flex-col gap-4">
      <UAlert
        v-if="pageError"
        color="error"
        variant="subtle"
        icon="i-lucide-circle-alert"
        :title="pageError"
      />

      <template v-if="!invalid && isAuthenticated">
        <UButton
          v-if="pageError"
          label="Try again"
          block
          :loading="accepting"
          @click="finishAccept"
        />
        <p
          v-else
          class="text-muted text-sm"
        >
          Accepting invite…
        </p>
        <p class="text-muted text-center text-sm">
          Wrong account?
          <button
            type="button"
            class="text-primary font-medium"
            @click="signOutLocal"
          >
            Sign out
          </button>
        </p>
      </template>

      <template v-else-if="!invalid">
        <UButton
          label="Create account"
          block
          :to="{ path: '/register', query: authQuery }"
        />
        <UButton
          label="Sign in"
          color="neutral"
          variant="outline"
          block
          :to="{ path: '/login', query: authQuery }"
        />
      </template>

      <UButton
        v-else
        label="Back home"
        to="/"
        color="neutral"
        variant="outline"
        block
      />
    </div>
  </UCard>
</template>
