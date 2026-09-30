<script setup lang="ts">
const route = useRoute()
const { findBySlug, resolveLandingSlug, rememberSlug, organizations } = useOrganizations()

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
  const landing = resolveLandingSlug()
  if (landing) {
    await navigateTo(`/o/${landing}`, { replace: true })
  } else if (!organizations.value.length) {
    await navigateTo('/', { replace: true })
  }
}
</script>

<template>
  <NuxtPage />
</template>
