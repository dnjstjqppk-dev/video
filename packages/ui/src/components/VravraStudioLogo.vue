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
    <span class="vravra-studio-row">
      <svg
        class="vravra-logo vravra-logo-glow" viewBox="0 0 64 40" fill="none"
        xmlns="http://www.w3.org/2000/svg" role="img" :aria-label="BRAND.studio.name"
        :height="height" :style="{ aspectRatio: '64 / 40' }"
      >
        <g class="logo-play" stroke="var(--color-brand-crimson)" stroke-width="6" stroke-linecap="round" fill="none">
          <path d="M9 7 L17 31" /><path d="M17 31 L22 21" />
        </g>
        <g class="logo-waves" fill="var(--bg-canvas)">
          <rect class="wave-bar" x="13.6" y="22" width="1.4" height="4" rx="0.7" /><rect class="wave-bar" x="15.8" y="19.5" width="1.4" height="9" rx="0.7" /><rect class="wave-bar" x="18" y="22" width="1.4" height="4" rx="0.7" />
        </g>
        <g class="logo-wordmark" fill="var(--color-studio-cobalt)">
          <rect x="33" y="23" width="5" height="9" rx="1" /><rect x="40" y="18" width="5" height="14" rx="1" /><rect x="47" y="13" width="5" height="19" rx="1" /><rect x="54" y="8" width="5" height="24" rx="1" />
        </g>
      </svg>
      <span class="vravra-studio-word" :class="size === 'sm' ? 'vravra-studio-word-sm' : ''" aria-hidden="true">
        <span class="vravra-studio-v">vravra</span>
        <span class="vravra-studio-s">studio</span>
      </span>
    </span>
    <span v-if="showTagline" class="vravra-tagline">{{ BRAND.studio.sloganEn }}</span>
  </span>
</template>

<style scoped>
.vravra-lockup { display: inline-flex; flex-direction: column; align-items: flex-start; gap: 2px; padding: 0.4em; }
.vravra-studio-row { display: inline-flex; align-items: center; gap: 8px; }
.vravra-studio-word { display: inline-flex; align-items: baseline; gap: 6px; font-weight: 800; letter-spacing: -0.02em; line-height: 1; }
@media (max-width: 479px) {
  .vravra-studio-word-sm { display: none; }
}
.vravra-studio-v { color: var(--color-brand-crimson); font-size: 19px; }
.vravra-studio-s { color: var(--color-studio-cobalt); font-size: 12px; text-transform: uppercase; letter-spacing: 0.12em; }
.vravra-tagline { font-size: 10px; font-weight: 600; letter-spacing: 0.14em; text-transform: uppercase; color: var(--text-secondary); white-space: nowrap; }
:global(.dark) .vravra-logo-glow { filter: drop-shadow(0 0 14px var(--color-studio-cobalt-glow)); }
</style>
