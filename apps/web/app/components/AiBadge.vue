<script setup lang="ts">
const props = withDefaults(defineProps<{
  tool?: string
  approved?: boolean
  size?: 'xs' | 'sm'
}>(), { size: 'xs' })
</script>

<template>
  <span
    class="inline-flex items-center gap-1 rounded bg-[var(--gvp-brand)] font-bold uppercase text-white"
    :class="size === 'sm' ? 'px-2 py-1 text-[11px]' : 'px-1.5 py-0.5 text-[10px]'"
    :title="tool ? `AI 제작${approved ? ' · Originals 선정작' : ''} — ${tool}` : (approved ? 'AI Originals 선정작' : 'AI 제작 영상')"
  >
    <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 3.5h4.5l-1 5.5h2.75l-5.5 9.5 1-5.5H8.75l1-9.5z" />
    </svg>
    AI
  </span>
</template>
