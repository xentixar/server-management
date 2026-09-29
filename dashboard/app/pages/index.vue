<script setup lang="ts">
definePageMeta({ layout: false })

const { resolveLandingSlug } = useOrganizations()

await navigateTo(`/o/${resolveLandingSlug()}`, { replace: true })
</script>

<template>
  <div />
</template>
