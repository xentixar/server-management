<template>
  <div class="relative flex min-h-svh flex-col items-center justify-center px-4 py-12">
    <div
      class="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top,var(--color-green-500)_0%,transparent_50%)] opacity-[0.08] dark:opacity-[0.12]"
      aria-hidden="true"
    />

    <div class="relative z-10 flex w-full max-w-sm flex-col gap-8">
      <NuxtLink
        to="/login"
        class="text-highlighted self-center text-2xl font-semibold tracking-tight"
      >
        Outpost
      </NuxtLink>

      <slot />
    </div>
  </div>
</template>
