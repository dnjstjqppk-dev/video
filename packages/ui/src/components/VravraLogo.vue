<script setup lang="ts">
withDefaults(defineProps<{
  size?: number
  suffix?: string
  suffixTone?: 'studio' | 'none'
}>(), { size: 28, suffix: '', suffixTone: 'none' })
</script>

<template>
  <span class="inline-flex items-center gap-2" aria-label="vravra">
    <svg
      :width="size" :height="size" viewBox="0 0 32 32" fill="none"
      role="img" aria-hidden="true"
    >
      <rect width="32" height="32" rx="8" fill="var(--gvp-brand)" />
      <rect x="7" y="12" width="2.6" height="8" rx="1.3" fill="#fff" opacity="0.85" />
      <rect x="11.2" y="9" width="2.6" height="14" rx="1.3" fill="#fff" opacity="0.85" />
      <path d="M16.5 11.2v9.6c0 .9.98 1.45 1.74.97l7.1-4.8a1.13 1.13 0 0 0 0-1.94l-7.1-4.8a1.13 1.13 0 0 0-1.74.97Z" fill="#fff" />
    </svg>
    <span class="text-[19px] font-extrabold leading-none tracking-tight text-[var(--gvp-text)]">
      vravra
      <span
        v-if="suffix"
        class="ml-1 align-middle text-[11px] font-bold uppercase tracking-widest"
        :class="suffixTone === 'studio' ? 'text-[var(--gvp-studio)]' : 'text-[var(--gvp-text-secondary)]'"
      >{{ suffix }}</span>
    </span>
  </span>
</template>
