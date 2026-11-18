<script setup lang="ts">
const store = useAppStore()
const api = useApi()
onMounted(() => {
  if (!store.currentUser && !store.authLoading) void store.fetchSession(api)
})
</script>

<template>
  <header class="gvp-glass sticky top-0 z-40 flex h-14 items-center justify-between gap-2 bg-[var(--gvp-surface)] px-3 sm:px-4">
    <div class="flex shrink-0 items-center gap-2">
      <button type="button" aria-label="사이드바 전환" @click="store.toggleSidebar()"
        class="hidden h-9 w-9 items-center justify-center rounded-full text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)] md:inline-flex">
        <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" d="M4 6h16M4 12h16M4 18h16" /></svg>
      </button>
      <NuxtLink to="/" class="flex items-center gap-2" aria-label="StreamVerse 홈">
        <span class="flex h-7 w-7 items-center justify-center rounded-lg bg-[var(--gvp-brand)] text-white">
          <svg class="h-3.5 w-3.5 fill-current" viewBox="0 0 24 24"><path d="M8 5v14l11-7z" /></svg>
        </span>
        <span class="hidden text-base font-bold tracking-tight text-[var(--gvp-text)] min-[420px]:inline">StreamVerse</span>
        <span class="tabular-nums hidden rounded border border-[var(--gvp-border)] bg-[var(--gvp-elevated)] px-1.5 py-0.5 text-[10px] font-semibold uppercase text-[var(--gvp-text-secondary)] min-[420px]:inline">{{ store.locale === 'ko' ? 'KR' : 'GL' }}</span>
      </NuxtLink>
    </div>
    <AppSearchBar />
    <div class="flex shrink-0 items-center gap-1 sm:gap-1.5">
      <button type="button" :aria-label="store.locale === 'ko' ? 'Switch to English' : '한국어로 전환'"
        class="hidden h-8 items-center rounded-md px-2 text-xs font-medium text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] lg:inline-flex" @click="store.toggleLocale()">
        {{ store.locale === 'ko' ? '한국어' : 'EN' }}
      </button>
      <button type="button" aria-label="통화 전환" class="tabular-nums hidden h-8 items-center rounded-md px-2 text-xs font-medium text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] lg:inline-flex" @click="store.toggleCurrency()">
        {{ store.currency }}
      </button>
      <AppCreateMenu v-if="store.currentUser" />
      <AppNotificationCenter v-if="store.currentUser" />
      <AppProfileMenu />
    </div>
  </header>
</template>
