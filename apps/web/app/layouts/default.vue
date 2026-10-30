<script setup lang="ts">
import { useAppStore } from '~/stores/app'

const appStore = useAppStore()
const api = useApi()
const router = useRouter()
const route = useRoute()
const config = useRuntimeConfig()

const searchQuery = ref((route.query.q as string) || '')
const suggestions = ref<string[]>([])
const showSuggestions = ref(false)
const showUserMenu = ref(false)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

const fallbackSuggestions = [
  '4K HDR 서울 야경 드론 (Seoul 4K Drone)',
  'Go 1.24 Microservices Architecture',
  'Nuxt 4 SSR Performance Optimization',
  'K-Pop Live Concert 2026',
  'Lo-Fi Coding Beats 24/7 Live'
]

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  const q = searchQuery.value.trim()
  if (!q) {
    suggestions.value = fallbackSuggestions
    return
  }
  debounceTimer = setTimeout(async () => {
    try {
      const res = await api.getSearchSuggestions(q)
      suggestions.value = res.length
        ? res
        : fallbackSuggestions.filter(item => item.toLowerCase().includes(q.toLowerCase()))
    } catch {
      suggestions.value = fallbackSuggestions.filter(item =>
        item.toLowerCase().includes(q.toLowerCase())
      )
    }
  }, 180)
}

function onSearchBlur() {
  setTimeout(() => (showSuggestions.value = false), 200)
}

function submitSearch(term?: string) {
  const finalQuery = (term ?? searchQuery.value).trim()
  if (!finalQuery) return
  searchQuery.value = finalQuery
  showSuggestions.value = false
  router.push({ path: '/search', query: { q: finalQuery } })
}

const navItems = computed(() => [
  {
    path: '/',
    icon: 'M3 10.5L12 3l9 7.5V21a1 1 0 0 1-1 1h-5.5v-6.5h-5V22H4a1 1 0 0 1-1-1V10.5z',
    label: appStore.locale === 'ko' ? '홈' : 'Home'
  },
  {
    path: '/shorts',
    icon: 'M17.77 10.32l-1.2-.5L18 9.06a3.74 3.74 0 0 0-3.5-6.62l-6 2.5a3.74 3.74 0 0 0 .23 6.74l1.2.5L8.5 12.9a3.74 3.74 0 0 0 3.5 6.62l6-2.5a3.74 3.74 0 0 0-.23-6.7zM10 14.65v-5.3L15 12l-5 2.65z',
    label: appStore.locale === 'ko' ? '쇼츠 (Shorts)' : 'Shorts'
  },
  {
    path: '/live/live_main_01',
    icon: 'M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8zm-7.07-2.93a10 10 0 0 0 0 14.14l1.41-1.41a8 8 0 0 1 0-11.32L4.93 5.07zm14.14 0l-1.41 1.41a8 8 0 0 1 0 11.32l1.41 1.41a10 10 0 0 0 0-14.14z',
    label: appStore.locale === 'ko' ? '라이브 (Live)' : 'Live Now'
  },
  {
    path: '/channels/ch_tech_kr',
    icon: 'M4 6h16v2H4zm2-4h12v2H6zm14 8H4a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-8a2 2 0 0 0-2-2z',
    label: appStore.locale === 'ko' ? '구독 채널' : 'Subscriptions'
  },
  {
    path: '/premium',
    icon: 'M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z',
    label: 'Premium'
  }
])

const subscribedChannels = [
  { id: 'ch_tech_kr', name: 'GVP Tech Engineering', handle: '@gvptech', isLive: true },
  { id: 'ch_seoul_4k', name: 'Seoul 4K Cinema', handle: '@seoul4k', isLive: false },
  { id: 'ch_esports_global', name: 'Global Esports Arena', handle: '@esports', isLive: true },
  { id: 'ch_kfood_master', name: 'K-Food Studio', handle: '@kfood', isLive: false }
]
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <!-- Top Navbar -->
    <header
      class="sticky top-0 z-50 flex h-14 items-center justify-between border-b px-4 backdrop-blur-md transition-colors"
      :class="appStore.theme === 'dark' ? 'border-white/10 bg-[#0F0F0F]/95' : 'border-black/10 bg-white/95'"
    >
      <!-- Left: Hamburger & Brand Logo -->
      <div class="flex items-center gap-3">
        <button
          type="button"
          class="rounded-full p-2 transition hover:bg-white/10"
          aria-label="Toggle navigation sidebar"
          @click="appStore.toggleSidebar()"
        >
          <svg class="h-6 w-6" viewBox="0 0 24 24" fill="currentColor">
            <path d="M3 6h18v2H3V6zm0 5h18v2H3v-2zm0 5h18v2H3v-2z" />
          </svg>
        </button>

        <NuxtLink to="/" class="flex items-center gap-2 font-bold tracking-tight">
          <span class="flex h-7 w-9 items-center justify-center rounded-lg bg-[#FF0033] text-white shadow-sm">
            <svg class="h-4 w-4 fill-current" viewBox="0 0 24 24">
              <path d="M8 5v14l11-7z" />
            </svg>
          </span>
          <span class="text-lg font-extrabold tracking-tight">StreamVerse</span>
          <span class="rounded bg-white/10 px-1.5 py-0.5 text-[10px] font-semibold uppercase text-[#3EA6FF]">
            {{ appStore.locale === 'ko' ? 'KR' : 'GLOBAL' }}
          </span>
        </NuxtLink>
      </div>

      <!-- Center: Search Bar with Autocomplete -->
      <div class="relative mx-4 flex max-w-2xl flex-1 items-center">
        <form class="flex w-full" @submit.prevent="submitSearch()">
          <div class="relative flex flex-1 items-center">
            <input
              v-model="searchQuery"
              type="search"
              :placeholder="appStore.locale === 'ko' ? '검색어를 입력하세요 (동영상, 쇼츠, 라이브, 채널)...' : 'Search videos, Shorts, live streams, channels...'"
              class="h-10 w-full rounded-l-full border px-4 text-sm outline-none transition focus:border-[#3EA6FF]"
              :class="appStore.theme === 'dark' ? 'border-white/20 bg-[#121212] text-white placeholder-neutral-400' : 'border-neutral-300 bg-neutral-100 text-black placeholder-neutral-500'"
              @focus="showSuggestions = true; onSearchInput()"
              @input="onSearchInput"
              @blur="onSearchBlur"
            />
          </div>
          <button
            type="submit"
            class="flex h-10 w-16 items-center justify-center rounded-r-full border border-l-0 transition"
            :class="appStore.theme === 'dark' ? 'border-white/20 bg-[#272727] text-white hover:bg-[#3A3A3A]' : 'border-neutral-300 bg-neutral-200 text-black hover:bg-neutral-300'"
            aria-label="Search"
          >
            <svg class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <circle cx="11" cy="11" r="8" />
              <path d="m21 21-4.35-4.35" />
            </svg>
          </button>
        </form>

        <!-- Autocomplete Dropdown -->
        <div
          v-if="showSuggestions && suggestions.length > 0"
          class="absolute left-0 top-12 z-50 w-full overflow-hidden rounded-2xl border py-2 shadow-2xl"
          :class="appStore.theme === 'dark' ? 'border-white/10 bg-[#212121] text-white' : 'border-black/10 bg-white text-black'"
        >
          <button
            v-for="item in suggestions"
            :key="item"
            type="button"
            class="flex w-full items-center gap-3 px-4 py-2 text-left text-sm transition hover:bg-white/10"
            @mousedown.prevent="submitSearch(item)"
          >
            <svg class="h-4 w-4 text-neutral-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <circle cx="11" cy="11" r="8" />
              <path d="m21 21-4.35-4.35" />
            </svg>
            <span>{{ item }}</span>
          </button>
        </div>
      </div>

      <!-- Right: Controls (Locale, Currency, Theme, Studio Link, User Menu) -->
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="rounded-full border border-white/15 px-2.5 py-1 text-xs font-semibold transition hover:bg-white/10"
          @click="appStore.toggleLocale()"
        >
          {{ appStore.locale === 'ko' ? '한국어' : 'EN' }}
        </button>

        <button
          type="button"
          class="rounded-full border border-white/15 px-2.5 py-1 text-xs font-semibold transition hover:bg-white/10"
          @click="appStore.toggleCurrency()"
        >
          {{ appStore.currency }}
        </button>

        <button
          type="button"
          class="rounded-full p-2 transition hover:bg-white/10"
          :title="appStore.theme === 'dark' ? 'Switch to Light Mode' : 'Switch to Dark Mode'"
          @click="appStore.toggleTheme()"
        >
          <svg v-if="appStore.theme === 'dark'" class="h-5 w-5 text-amber-400" fill="currentColor" viewBox="0 0 20 20">
            <path d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0z" />
          </svg>
          <svg v-else class="h-5 w-5 text-slate-700" fill="currentColor" viewBox="0 0 20 20">
            <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
          </svg>
        </button>

        <a
          :href="config.public.studioUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hidden items-center gap-1.5 rounded-full bg-[#272727] px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-[#3A3A3A] sm:flex"
        >
          <svg class="h-4 w-4 text-[#FF0033]" fill="currentColor" viewBox="0 0 24 24">
            <path d="M17 10.5V7c0-.55-.45-1-1-1H4c-.55 0-1 .45-1 1v10c0 .55.45 1 1 1h12c.55 0 1-.45 1-1v-3.5l4 4v-11l-4 4z" />
          </svg>
          <span>Studio</span>
        </a>

        <div class="relative">
          <button
            type="button"
            class="flex items-center gap-2 rounded-full p-1 transition hover:ring-2 hover:ring-[#3EA6FF]"
            @click="showUserMenu = !showUserMenu"
          >
            <img
              :src="appStore.currentUser?.avatarUrl"
              :alt="appStore.currentUser?.displayName"
              class="h-8 w-8 rounded-full object-cover"
            />
          </button>

          <div
            v-if="showUserMenu"
            class="absolute right-0 top-11 z-50 w-64 rounded-2xl border p-3 shadow-2xl"
            :class="appStore.theme === 'dark' ? 'border-white/10 bg-[#212121] text-white' : 'border-black/10 bg-white text-black'"
          >
            <div class="border-b border-white/10 pb-3">
              <p class="font-semibold">{{ appStore.currentUser?.displayName }}</p>
              <p class="text-xs text-neutral-400">@{{ appStore.currentUser?.username }} · {{ appStore.currentUser?.subscriptionTier.toUpperCase() }}</p>
            </div>
            <div class="mt-2 flex flex-col gap-1 text-sm">
              <NuxtLink to="/channels/ch_tech_kr" class="rounded-lg px-3 py-2 hover:bg-white/10" @click="showUserMenu = false">
                {{ appStore.locale === 'ko' ? '내 채널 보기' : 'Your Channel' }}
              </NuxtLink>
              <NuxtLink to="/premium" class="rounded-lg px-3 py-2 text-[#3EA6FF] hover:bg-white/10" @click="showUserMenu = false">
                {{ appStore.locale === 'ko' ? 'StreamVerse Premium 멤버십' : 'StreamVerse Premium' }}
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- Body Layout -->
    <div class="flex flex-1">
      <!-- Left Collapsible Sidebar -->
      <aside
        class="sticky top-14 hidden h-[calc(100vh-3.5rem)] shrink-0 overflow-y-auto border-r transition-all md:block"
        :class="[
          appStore.sidebarCollapsed ? 'w-20 px-2' : 'w-60 px-3',
          appStore.theme === 'dark' ? 'border-white/10 bg-[#0F0F0F]' : 'border-black/10 bg-white'
        ]"
      >
        <nav class="mt-3 flex flex-col gap-1">
          <NuxtLink
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="flex items-center gap-4 rounded-xl px-3 py-2.5 text-sm font-medium transition"
            :class="[
              route.path === item.path
                ? (appStore.theme === 'dark' ? 'bg-[#272727] text-white font-semibold' : 'bg-neutral-200 text-black font-semibold')
                : 'hover:bg-white/10 text-neutral-300'
            ]"
          >
            <svg class="h-5 w-5 shrink-0" viewBox="0 0 24 24" fill="currentColor">
              <path :d="item.icon" />
            </svg>
            <span v-if="!appStore.sidebarCollapsed" class="truncate">{{ item.label }}</span>
          </NuxtLink>
        </nav>

        <!-- Subscribed Channels -->
        <div v-if="!appStore.sidebarCollapsed" class="mt-6 border-t border-white/10 pt-4">
          <h3 class="px-3 text-xs font-semibold uppercase tracking-wider text-neutral-400">
            {{ appStore.locale === 'ko' ? '구독 채널' : 'Subscriptions' }}
          </h3>
          <div class="mt-2 flex flex-col gap-1">
            <NuxtLink
              v-for="sub in subscribedChannels"
              :key="sub.id"
              :to="`/channels/${sub.id}`"
              class="flex items-center justify-between rounded-xl px-3 py-2 text-sm transition hover:bg-white/10"
            >
              <div class="flex items-center gap-2.5 truncate">
                <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-[#FF0033] to-purple-600 text-xs font-bold text-white">
                  {{ sub.name.slice(0, 1) }}
                </span>
                <span class="truncate">{{ sub.name }}</span>
              </div>
              <span
                v-if="sub.isLive"
                class="h-2 w-2 shrink-0 rounded-full bg-[#FF0033] ring-4 ring-[#FF0033]/20"
                title="LIVE"
              />
            </NuxtLink>
          </div>
        </div>
      </aside>

      <!-- Main Content -->
      <main class="min-w-0 flex-1">
        <slot />
      </main>
    </div>
  </div>
</template>
