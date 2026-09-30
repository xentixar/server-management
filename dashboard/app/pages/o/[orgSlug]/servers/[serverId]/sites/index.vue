<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import type { Site } from '~/utils/sites'
import { siteStatusColor, stackLabel } from '~/utils/sites'

const route = useRoute()
const { orgHref } = useOrganizations()
const { sites, remove } = useSites()
const toast = useToast()

const serverId = computed(() => String(route.params.serverId || ''))
const q = ref('')
const deleteOpen = ref(false)
const siteToDelete = ref<Site | null>(null)

const serverSites = computed(() =>
  sites.value.filter(s => s.serverId === serverId.value)
)

const filtered = computed(() => {
  const query = q.value.trim().toLowerCase()
  if (!query) {
    return serverSites.value
  }
  return serverSites.value.filter((site) => {
    return site.name.toLowerCase().includes(query)
      || site.domain.toLowerCase().includes(query)
      || site.stack.includes(query)
  })
})

function menuItems(site: Site): DropdownMenuItem[][] {
  return [[{
    label: 'Delete site',
    icon: 'i-lucide-trash-2',
    color: 'error',
    onSelect: () => {
      siteToDelete.value = site
      deleteOpen.value = true
    }
  }]]
}

function onConfirmDelete() {
  if (!siteToDelete.value) {
    return
  }
  const result = remove(siteToDelete.value.id)
  if (!result.ok) {
    toast.add({ title: result.error, color: 'error' })
    return
  }
  toast.add({
    title: 'Site deleted',
    description: result.site.domain,
    color: 'neutral'
  })
  siteToDelete.value = null
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h2 class="text-highlighted text-lg font-semibold">
          Sites
        </h2>
        <p class="text-muted text-sm">
          Applications hosted on this server.
        </p>
      </div>
      <UButton
        :to="orgHref(`/servers/${serverId}/sites/new`)"
        label="Add site"
        icon="i-lucide-plus"
      />
    </div>

    <template v-if="!serverSites.length">
      <section class="border-default from-elevated/40 to-default rounded-lg border bg-linear-to-br px-5 py-10 sm:px-8">
        <div class="flex max-w-xl flex-col gap-4">
          <div class="bg-primary/10 text-primary flex size-10 items-center justify-center rounded-lg">
            <UIcon
              name="i-lucide-globe"
              class="size-5"
            />
          </div>
          <div class="flex flex-col gap-2">
            <h3 class="text-highlighted text-lg font-semibold">
              No sites yet
            </h3>
            <p class="text-muted text-sm leading-relaxed">
              Create a site to configure Nginx, stack, and deploys for a domain on this server.
            </p>
          </div>
          <UButton
            :to="orgHref(`/servers/${serverId}/sites/new`)"
            label="Add site"
            icon="i-lucide-plus"
            class="w-fit"
          />
        </div>
      </section>
    </template>

    <template v-else>
      <UInput
        v-model="q"
        icon="i-lucide-search"
        placeholder="Search name or domain"
        class="w-full sm:max-w-xs"
      />

      <ul class="border-default divide-y divide-default overflow-hidden rounded-lg border">
        <li
          v-for="site in filtered"
          :key="site.id"
          class="flex flex-wrap items-center justify-between gap-3 px-4 py-3.5 sm:px-5"
        >
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <p class="text-highlighted truncate text-sm font-medium">
                {{ site.domain }}
              </p>
              <UBadge
                :label="site.status"
                :color="siteStatusColor(site.status)"
                variant="subtle"
                size="sm"
                class="capitalize"
              />
              <UBadge
                :label="stackLabel(site.stack)"
                color="neutral"
                variant="subtle"
                size="sm"
              />
            </div>
            <p class="text-muted mt-0.5 text-xs">
              {{ site.name }}
              · {{ site.webRoot }}
              <template v-if="site.gitRepo">
                · {{ site.gitBranch || 'main' }}
              </template>
            </p>
          </div>

          <UDropdownMenu
            :items="menuItems(site)"
            :content="{ align: 'end' }"
          >
            <UButton
              icon="i-lucide-ellipsis-vertical"
              color="neutral"
              variant="ghost"
              aria-label="Site actions"
            />
          </UDropdownMenu>
        </li>
      </ul>

      <p
        v-if="!filtered.length"
        class="text-muted text-sm"
      >
        No sites match your search.
      </p>
    </template>

    <OrganizationDeleteSiteModal
      v-model:open="deleteOpen"
      :site="siteToDelete"
      @confirm="onConfirmDelete"
    />
  </div>
</template>
