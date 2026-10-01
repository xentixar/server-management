import type { Ability } from '~/utils/auth/abilities'

export function useGate() {
  const { selected } = useOrganizations()

  function can(ability: Ability | string) {
    return !!selected.value?.permissions?.includes(ability)
  }

  function cannot(ability: Ability | string) {
    return !can(ability)
  }

  return { can, cannot }
}
