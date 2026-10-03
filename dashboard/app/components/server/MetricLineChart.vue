<script setup lang="ts" generic="T extends Record<string, unknown>">
import { VisArea, VisAxis, VisCrosshair, VisLine, VisTooltip, VisXYContainer } from '@unovis/vue'
import type { MetricChartSeries } from '~/utils/metrics'

const props = defineProps<{
  title: string
  data: T[]
  x: (d: T) => number
  xDomain: [number, number]
  yDomain?: [number, number | undefined]
  yTickFormat?: (v: number) => string
  xTickFormat: (v: number | Date) => string
  series: MetricChartSeries<T>[]
  formatTooltip: (d: T) => string
}>()

const colorMode = useColorMode()

/** Inline vars beat Unovis :root injection for this chart subtree. */
const chartThemeStyle = computed(() => {
  if (colorMode.value !== 'dark') {
    return undefined
  }
  return {
    '--vis-axis-grid-color': 'rgb(255 255 255 / 0.06)',
    '--vis-axis-tick-color': 'rgb(255 255 255 / 0.08)',
    '--vis-axis-domain-color': 'rgb(255 255 255 / 0.08)',
    '--vis-axis-tick-label-color': 'var(--ui-text-muted)',
    '--vis-axis-label-color': 'var(--ui-text-muted)'
  } as Record<string, string>
})

const xAcc = (d: T) => props.x(d)

function yAcc(s: MetricChartSeries<T>) {
  return (d: T) => {
    const v = s.y(d)
    // undefined → Unovis line gap (first CPU sample often missing)
    return v == null || !Number.isFinite(v) ? undefined! : v
  }
}

const crosshairY = computed(() => props.series.map(s => yAcc(s)))
const crosshairColor = computed(() => {
  if (props.series.length === 1) {
    return props.series[0]!.color
  }
  return props.series.map(s => s.color)
})

function tooltipTemplate(d: T | undefined) {
  if (!d) {
    return ''
  }
  return props.formatTooltip(d)
}
</script>

<template>
  <section
    class="border-default flex flex-col gap-3 rounded-lg border p-4"
    :style="chartThemeStyle"
  >
    <div class="flex items-baseline justify-between gap-2">
      <h3 class="text-highlighted text-sm font-medium">
        {{ title }}
      </h3>
      <p
        v-if="series.length > 1"
        class="text-muted flex flex-wrap justify-end gap-x-3 gap-y-1 text-xs"
      >
        <span
          v-for="s in series"
          :key="s.key"
          class="inline-flex items-center gap-1.5"
        >
          <span
            class="size-2 shrink-0 rounded-full"
            :style="{ backgroundColor: s.color }"
          />
          {{ s.label }}
        </span>
      </p>
    </div>

    <div class="min-h-[220px] w-full">
      <VisXYContainer
        :data="data"
        :height="220"
        :x-domain="xDomain"
        :y-domain="yDomain"
        :prevent-empty-domain="true"
      >
        <template
          v-for="s in series"
          :key="s.key"
        >
          <VisArea
            v-if="s.area"
            :x="xAcc"
            :y="yAcc(s)"
            :color="s.color"
            :opacity="0.18"
          />
          <VisLine
            :x="xAcc"
            :y="yAcc(s)"
            :color="s.color"
            :line-width="2"
          />
        </template>
        <VisAxis
          type="x"
          :tick-format="xTickFormat"
          :num-ticks="5"
          :tick-text-hide-overlapping="true"
        />
        <VisAxis
          type="y"
          :num-ticks="4"
          :tick-format="yTickFormat"
        />
        <VisCrosshair
          :x="xAcc"
          :y="crosshairY"
          :color="crosshairColor"
          :template="tooltipTemplate"
        />
        <VisTooltip />
      </VisXYContainer>
    </div>
  </section>
</template>
