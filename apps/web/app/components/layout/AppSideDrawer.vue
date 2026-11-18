<script setup lang="ts">
const store = useAppStore()
const route = useRoute()
const subs = useSubscriptionChannels()
const showAllSubs = ref(false)
const ko = computed(() => store.locale === 'ko')
const mini = computed(() => store.sidebarCollapsed)
const main = computed(() => [
  { path: '/', label: ko.value ? '홈' : 'Home', icon: 'M3 10.5L12 3l9 7.5V21a1 1 0 0 1-1 1h-5.5v-6.5h-5V22H4a1 1 0 0 1-1-1V10.5z' },
  { path: '/shorts', label: ko.value ? '숏폼' : 'Shorts', icon: 'M10 14.65v-5.3L15 12l-5 2.65z' },
  { path: '/feed/subscriptions', label: ko.value ? '구독' : 'Subscriptions', icon: 'M4 6h16v2H4zm2-4h12v2H6zm14 8H4a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-8a2 2 0 0 0-2-2z' },
  { path: '/feed/history', label: ko.value ? '시청 기록' : 'History', icon: 'M12 8v4l3 3m6-3a9 9 0 1 1-9-9 9 9 0 0 1 9 9z' },
  { path: '/feed/liked', label: ko.value ? '좋아요' : 'Liked', icon: 'M12 21s-8-4.5-8-11a4.5 4.5 0 0 1 8-3 4.5 4.5 0 0 1 8 3c0 6.5-8 11-8 11z' },
])
const explore = computed(() => [
  { path: '/feed/trending', label: ko.value ? '인기 급상승' : 'Trending' },
  { path: '/feed/music', label: ko.value ? '음악' : 'Music' },
  { path: '/live/live_main_01', label: ko.value ? '라이브' : 'Live' },
  { path: '/feed/gaming', label: ko.value ? '게임' : 'Gaming' },
])
const shownSubs = computed(() => showAllSubs.value ? subs.channels.value : subs.channels.value.slice(0, 5))
</script>

<template>
  <aside :aria-label="ko ? '탐색 메뉴' : 'Navigation'"
    :class="mini ? 'w-[72px] px-2' : 'w-60 px-3'"
    class="sticky top-14 hidden h-[calc(100vh-3.5rem)] shrink-0 overflow-y-auto border-r border-[var(--gvp-border)] bg-[var(--gvp-surface)] transition-all md:block">
    <nav class="mt-3 flex flex-col gap-0.5">
      <NuxtLink v-for="it in main" :key="it.path" :to="it.path" :title="mini ? it.label : undefined"
        :class="[route.path === it.path ? 'bg-[var(--gvp-elevated)] font-semibold text-[var(--gvp-text)]' : 'text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]', mini ? 'flex-col gap-1 px-1 py-2.5 text-[10px]' : 'gap-3 px-3 py-2 text-sm']"
        class="flex items-center rounded-lg font-medium">
        <svg class="h-4 w-4 shrink-0" viewBox="0 0 24 24" fill="currentColor"><path :d="it.icon" /></svg>
        <span :class="mini ? '' : 'truncate'">{{ mini ? it.label : it.label }}</span>
      </NuxtLink>
    </nav>
    <template v-if="!mini">
      <div class="mt-5 border-t border-[var(--gvp-border)] pt-4">
        <h3 class="px-3 text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-muted)]">{{ ko ? '구독 채널' : 'Subscriptions' }}</h3>
        <p v-if="!subs.pending.value && !subs.channels.value.length" class="px-3 py-2 text-xs text-[var(--gvp-text-secondary)]">{{ ko ? '구독 중인 채널이 없습니다.' : 'No subscriptions yet.' }}</p>
        <NuxtLink v-for="s in shownSubs" :key="s.id" :to="`/channels/${s.id}`" class="flex items-center justify-between rounded-lg px-3 py-1.5 text-sm text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]">
          <span class="flex min-w-0 items-center gap-2.5">
            <img v-if="s.avatarUrl" :src="s.avatarUrl" :alt="s.name" class="h-6 w-6 shrink-0 rounded-full object-cover" />
            <span v-else class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-[var(--gvp-elevated)] text-xs font-semibold">{{ s.name.slice(0, 1) }}</span>
            <span class="truncate">{{ s.name }}</span>
          </span>
          <span v-if="s.isLive" class="tabular-nums rounded bg-[var(--gvp-live)] px-1 py-0.5 text-[10px] font-bold text-white">LIVE</span>
          <span v-else class="h-2 w-2 shrink-0 rounded-full bg-[var(--gvp-accent)]" title="새 영상" />
        </NuxtLink>
        <button v-if="subs.channels.value.length > 5" type="button" class="w-full rounded-lg px-3 py-1.5 text-left text-xs text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]" @click="showAllSubs = !showAllSubs">{{ showAllSubs ? (ko ? '간략히' : 'Show less') : (ko ? '더보기' : 'Show more') }}</button>
      </div>
      <div class="mb-6 mt-5 border-t border-[var(--gvp-border)] pt-4">
        <h3 class="px-3 text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-muted)]">{{ ko ? '탐색' : 'Explore' }}</h3>
        <NuxtLink v-for="it in explore" :key="it.path" :to="it.path" class="block rounded-lg px-3 py-1.5 text-sm text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]">{{ it.label }}</NuxtLink>
      </div>
    </template>
  </aside>
</template>
