import { createSharedComposable } from '@vueuse/core'

const _useDashboard = () => {
  const route = useRoute()
  const router = useRouter()
  const { orgHref } = useOrganizations()
  const isNotificationsSlideoverOpen = ref(false)

  defineShortcuts({
    'g-h': () => router.push(orgHref()),
    'g-o': () => router.push(orgHref('/organization')),
    'n': () => isNotificationsSlideoverOpen.value = !isNotificationsSlideoverOpen.value
  })

  watch(() => route.fullPath, () => {
    isNotificationsSlideoverOpen.value = false
  })

  return {
    isNotificationsSlideoverOpen
  }
}

export const useDashboard = createSharedComposable(_useDashboard)
