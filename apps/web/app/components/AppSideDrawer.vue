<script setup lang="ts">
const store = useAppStore()
const route = useRoute()
const subs = useSubscriptionChannels()
const live = useDrawerLive()
const showAllSubs = ref(false)
const ko = computed(() => store.locale === 'ko')
const mini = computed(() => store.sidebarCollapsed)
const home = 'M3 10.5L12 3l9 7.5V21a1 1 0 0 1-1 1h-5.5v-6.5h-5V22H4a1 1 0 0 1-1-1V10.5z'
const shorts = 'M13 2L3 14h7l-1 8 10-12h-7l1-8z'
const subIc = 'M4 6h16v2H4zm2-4h12v2H6zm14 8H4a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-8a2 2 0 0 0-2-2z'
const aiIc = 'M12 2l2.2 6.3L20 10l-5.8 1.7L12 18l-2.2-6.3L4 10l5.8-1.7L12 2zm7 11l1.1 3.1L23 17l-2.9.9L19 21l-1.1-3.1L15 17l2.9-.9L19 13z'
const main = computed(() => [
  { path: '/', label: ko.value ? '홈' : 'Home', icon: home },
  { path: '/shorts', label: ko.value ? '숏폼' : 'Shorts', icon: shorts },
  { path: '/feed/subscriptions', label: ko.value ? '구독' : 'Subscriptions', icon: subIc },
  { path: '/ai', label: ko.value ? 'AI Originals' : 'AI Originals', icon: aiIc },
])
const you = computed(() => [
  { path: '/feed/history', label: ko.value ? '시청 기록' : 'History' },
  { path: '/feed/saved', label: ko.value ? '나중에 볼 동영상' : 'Watch Later' },
  { path: '/feed/liked', label: ko.value ? '좋아요 표시한 동영상' : 'Liked videos' },
])
const explore = computed(() => [
  { path: '/feed/trending', label: ko.value ? '인기 급상승' : 'Trending' },
  { path: '/feed/music', label: ko.value ? '음악' : 'Music' },
  { path: '/feed/gaming', label: ko.value ? '게임' : 'Gaming' },
  { path: '/feed/tech', label: ko.value ? '테크' : 'Tech' },
])
const shownSubs = computed(() => showAllSubs.value ? subs.channels.value : subs.channels.value.slice(0, 5))
onMounted(() => {
  const mq = window.matchMedia('(max-width: 1279px)')
  const sync = () => store.setSidebarCollapsed(mq.matches)
  sync()
  mq.addEventListener('change', sync)
})
</script>

<template>
  <aside :aria-label="ko ? '탐색 메뉴' : 'Navigation'"
    :class="mini ? 'w-[72px] px-2' : 'w-60 px-3'"
    class="sticky top-14 z-30 hidden h-[calc(100vh-3.5rem)] shrink-0 overflow-y-auto bg-[var(--gvp-canvas)] transition-all md:block">
    <nav class="mt-3 flex flex-col gap-0.5">
      <NuxtLink v-for="it in main" :key="it.path" :to="it.path" :title="mini ? it.label : undefined"
        :class="[route.path === it.path ? 'bg-[var(--gvp-elevated)] font-semibold text-[var(--gvp-text)]' : 'text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]', mini ? 'flex-col gap-1 px-1 py-2.5 text-[10px]' : 'gap-3 px-3 py-2 text-sm']"
        class="flex items-center rounded-lg font-medium">
        <svg class="h-4 w-4 shrink-0" viewBox="0 0 24 24" fill="currentColor"><path :d="it.icon" /></svg>
        <span class="truncate">{{ it.label }}</span>
      </NuxtLink>
    </nav>
    <template v-if="!mini">
      <div class="mt-5 border-t border-[var(--gvp-border)] pt-4">
        <h3 class="px-3 text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-muted)]">{{ ko ? '나' : 'You' }}</h3>
        <NuxtLink v-for="it in you" :key="it.path" :to="it.path" class="block rounded-lg px-3 py-1.5 text-sm text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]">{{ it.label }}</NuxtLink>
      </div>
      <div class="mt-5 border-t border-[var(--gvp-border)] pt-4">
        <h3 class="px-3 text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-muted)]">{{ ko ? '구독 채널' : 'Subscriptions' }}</h3>
        <p v-if="!subs.pending.value && !subs.channels.value.length" class="px-3 py-2 text-xs text-[var(--gvp-text-secondary)]">{{ ko ? '구독 중인 채널이 없습니다.' : 'No subscriptions yet.' }}</p>
        <DrawerSubItem v-for="s in shownSubs" :key="s.id" :sub="s" />
        <button v-if="subs.channels.value.length > 5" type="button" class="w-full rounded-lg px-3 py-1.5 text-left text-xs text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]" @click="showAllSubs = !showAllSubs">{{ showAllSubs ? (ko ? '간략히' : 'Show less') : (ko ? '더보기' : 'Show more') }}</button>
      </div>
      <div class="mb-6 mt-5 border-t border-[var(--gvp-border)] pt-4">
        <h3 class="px-3 text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-muted)]">{{ ko ? '탐색' : 'Explore' }}</h3>
        <NuxtLink v-for="it in explore" :key="it.path" :to="it.path" class="block rounded-lg px-3 py-1.5 text-sm text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]">{{ it.label }}</NuxtLink>
        <NuxtLink v-if="live.liveId.value" :to="`/live/${live.liveId.value}`" class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium text-[var(--gvp-live)] hover:bg-[var(--gvp-elevated)]">
          <span class="h-1.5 w-1.5 animate-pulse rounded-full bg-[var(--gvp-live)]" />{{ ko ? '라이브' : 'Live' }}
        </NuxtLink>
      </div>
    </template>
  </aside>
</template>
