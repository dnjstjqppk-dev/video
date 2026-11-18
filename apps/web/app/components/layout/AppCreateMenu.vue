<script setup lang="ts">
const config = useRuntimeConfig()
const open = ref(false)
const studio = computed(() => config.public.studioUrl as string)
const items = computed(() => [
  { label: '동영상 업로드', hint: 'U', href: `${studio.value}/upload`, icon: 'M12 16V4m0 0L8 8m4-4l4 4M4 17v2a1 1 0 001 1h14a1 1 0 001-1v-2' },
  { label: '숏폼(Shorts) 등록', hint: 'S', href: `${studio.value}/upload?type=shorts`, icon: 'M10 14.65v-5.3L15 12l-5 2.65z' },
  { label: '실시간 스트리밍 시작', hint: 'L', href: `${studio.value}/live`, icon: 'M12 8a4 4 0 100 8 4 4 0 000-8zm-7 4a7 7 0 0114 0' },
  { label: '커뮤니티 게시글 작성', hint: 'P', href: `${studio.value}/content?tab=posts`, icon: 'M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2v10z' },
])
</script>

<template>
  <div v-click-outside="() => (open = false)" class="relative">
    <button type="button" :aria-expanded="open" aria-haspopup="menu" aria-label="만들기"
      class="inline-flex h-9 items-center gap-1.5 rounded-full bg-[var(--gvp-elevated)] px-3.5 text-sm font-medium text-[var(--gvp-text)] hover:bg-[var(--gvp-brand)] hover:text-white" @click="open = !open">
      <span class="text-lg leading-none">＋</span><span class="hidden sm:inline">만들기</span>
    </button>
    <div v-if="open" role="menu" class="gvp-dropdown gvp-elevated absolute right-0 top-full z-50 mt-2 w-64 rounded-xl p-1.5">
      <a v-for="it in items" :key="it.label" :href="it.href" target="_blank" rel="noopener noreferrer" role="menuitem"
        class="flex items-center gap-3 rounded-lg px-2.5 py-2 text-sm text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]" @click="open = false">
        <svg class="h-4 w-4 shrink-0 text-[var(--gvp-text-secondary)]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75"><path stroke-linecap="round" stroke-linejoin="round" :d="it.icon" /></svg>
        <span class="flex-1">{{ it.label }}</span>
        <kbd class="font-mono text-[11px] text-[var(--gvp-text-muted)]">{{ it.hint }}</kbd>
      </a>
    </div>
  </div>
</template>
