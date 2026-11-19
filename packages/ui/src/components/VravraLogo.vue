<script setup lang="ts">
import { BRAND } from '@gvp/design-tokens/src/brand'

const props = withDefaults(defineProps<{
  size?: 'sm' | 'md' | 'lg'
  animate?: boolean
  showTagline?: boolean
}>(), { size: 'md', animate: false, showTagline: false })

const height = computed(() => ({ sm: 22, md: 30, lg: 44 })[props.size])
</script>

<template>
  <span class="vravra-lockup" :class="animate ? 'vravra-logo-animate' : ''">
    <svg
      class="vravra-logo vravra-logo-glow" viewBox="0 0 260 64" fill="none"
      xmlns="http://www.w3.org/2000/svg" role="img" :aria-label="BRAND.name"
      :height="height" :style="{ aspectRatio: '260 / 64' }"
    >
      <path class="logo-play" d="M8 23.5C8 21.8 9.9 20.8 11.3 21.7L20.5 27.7C21.8 28.5 21.8 30.5 20.5 31.3L11.3 37.3C9.9 38.2 8 37.2 8 35.5V23.5Z" fill="var(--color-brand-crimson)" />
      <g class="logo-wordmark" fill="var(--color-brand-crimson)">
        <path d="M24 18H33.5L42.5 42L51.5 18H61L47.2 52H37.8L24 18Z" />
        <path d="M66 18H74.5V23.2C76.4 19.6 79.8 17.5 84.5 17.5H87V26H83.5C77.8 26 75 29.2 75 35.2V52H66V18Z" />
        <path d="M105.5 17.3C115.6 17.3 122.5 24.4 122.5 35C122.5 45.6 115.6 52.7 105.5 52.7C100.3 52.7 96.2 50.5 94 47.2V52H85.5V35C85.5 24.4 92.4 17.3 105.5 17.3ZM104 44.8C109.7 44.8 113.6 40.7 113.6 35C113.6 29.3 109.7 25.2 104 25.2C98.3 25.2 94.4 29.3 94.4 35C94.4 40.7 98.3 44.8 104 44.8Z" />
        <path d="M127 18H136.5L145.5 42L154.5 18H164L150.2 52H140.8L127 18Z" />
        <path d="M169 18H177.5V23.2C179.4 19.6 182.8 17.5 187.5 17.5H190V26H186.5C180.8 26 178 29.2 178 35.2V52H169V18Z" />
        <path d="M208.5 17.3C218.6 17.3 225.5 24.4 225.5 35C225.5 45.6 218.6 52.7 208.5 52.7C203.3 52.7 199.2 50.5 197 47.2V52H188.5V35C188.5 24.4 195.4 17.3 208.5 17.3ZM207 44.8C212.7 44.8 216.6 40.7 216.6 35C216.6 29.3 212.7 25.2 207 25.2C201.3 25.2 197.4 29.3 197.4 35C197.4 40.7 201.3 44.8 207 44.8Z" />
      </g>
      <g class="logo-waves" fill="var(--bg-canvas)">
        <rect class="wave-bar" x="37.5" y="24" width="1.8" height="6" rx="0.9" /><rect class="wave-bar" x="40.1" y="21" width="1.8" height="12" rx="0.9" /><rect class="wave-bar" x="42.7" y="18" width="1.8" height="18" rx="0.9" /><rect class="wave-bar" x="45.3" y="21" width="1.8" height="12" rx="0.9" /><rect class="wave-bar" x="47.9" y="24" width="1.8" height="6" rx="0.9" />
        <rect class="wave-bar" x="140.5" y="24" width="1.8" height="6" rx="0.9" /><rect class="wave-bar" x="143.1" y="21" width="1.8" height="12" rx="0.9" /><rect class="wave-bar" x="145.7" y="18" width="1.8" height="18" rx="0.9" /><rect class="wave-bar" x="148.3" y="21" width="1.8" height="12" rx="0.9" /><rect class="wave-bar" x="150.9" y="24" width="1.8" height="6" rx="0.9" />
      </g>
    </svg>
    <span v-if="showTagline" class="vravra-tagline">{{ BRAND.sloganEn }}</span>
  </span>
</template>

<style scoped>
.vravra-lockup { display: inline-flex; flex-direction: column; align-items: flex-start; gap: 2px; padding: 0.4em; }
.vravra-tagline { font-size: 10px; font-weight: 600; letter-spacing: 0.14em; text-transform: uppercase; color: var(--text-secondary); white-space: nowrap; }
:global(.dark) .vravra-logo-glow { filter: drop-shadow(0 0 14px var(--color-brand-crimson-glow)); }
</style>
