<script setup lang="ts">
const route = useRoute()
const { findBySlug, resolveLandingSlug, rememberSlug } = useOrganizations()

const slug = computed(() => {
  const raw = route.params.orgSlug
  return typeof raw === 'string' ? raw : ''
})

watchEffect(() => {
  const org = findBySlug(slug.value)
  if (!org) {
    return
  }
  rememberSlug(org.slug)
})

if (!findBySlug(slug.value)) {
  await navigateTo(`/o/${resolveLandingSlug()}`, { replace: true })
}
</script>

<template>
  <NuxtPage />
</template>
