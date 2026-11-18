<script setup lang="ts">
const store = useAppStore()
const route = useRoute()
const ko = computed(() => store.locale === 'ko')
const tabs = computed(() => [
  { path: '/', label: ko.value ? '홈' : 'Home', icon: 'M3 10.5L12 3l9 7.5V21a1 1 0 0 1-1 1h-5.5v-6.5h-5V22H4a1 1 0 0 1-1-1V10.5z' },
  { path: '/shorts', label: '숏폼', icon: 'M10 14.65v-5.3L15 12l-5 2.65z' },
  { path: '/feed/subscriptions', label: ko.value ? '구독' : 'Subs', icon: 'M4 6h16v2H4zm2-4h12v2H6zm14 8H4a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-8a2 2 0 0 0-2-2z' },
  { path: '/feed/history', label: ko.value ? '나' : 'You', icon: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8zm-7 8a7 7 0 0 1 14 0' },
])
</script>

<template>
  <nav aria-label="모바일 탐색" class="gvp-glass fixed inset-x-0 bottom-0 z-40 flex border-t border-[var(--gvp-border)] bg-[var(--gvp-surface)] pb-[env(safe-area-inset-bottom)] md:hidden">
    <NuxtLink v-for="t in tabs" :key="t.path" :to="t.path"
      :aria-current="route.path === t.path ? 'page' : undefined"
      :class="route.path === t.path ? 'text-[var(--gvp-text)]' : 'text-[var(--gvp-text-secondary)]'"
      class="flex flex-1 flex-col items-center gap-0.5 py-2 text-[10px] font-medium">
      <svg class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor"><path :d="t.icon" /></svg>
      {{ t.label }}
    </NuxtLink>
  </nav>
</template>
