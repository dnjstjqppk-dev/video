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

onMounted(() => {
  if (!appStore.currentUser && !appStore.authLoading) {
    void appStore.fetchSession(api)
  }
})

async function logout() {
  showUserMenu.value = false
  try {
    await $fetch('/auth/logout', { method: 'POST' })
  } catch {
    // Session cookies are cleared server-side; continue locally regardless.
  }
  const tokenCookie = useCookie('gvp_access_token')
  tokenCookie.value = null
  appStore.clearSession()
  await navigateTo('/')
}
</script>

<template>
  <div class="flex min-h-screen flex-col bg-white text-slate-900">
    <!-- Top Navbar -->
    <header
      class="sticky top-0 z-40 flex h-14 items-center justify-between border-b border-slate-200 bg-white px-4"
    >
      <!-- Left: Hamburger & Brand Logo -->
      <div class="flex items-center gap-3">
        <button
          type="button"
          class="inline-flex h-9 w-9 items-center justify-center rounded-md text-slate-600 transition-colors hover:bg-slate-100 hover:text-slate-900"
          aria-label="Toggle navigation sidebar"
          @click="appStore.toggleSidebar()"
        >
          <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>

        <NuxtLink to="/" class="flex items-center gap-2">
          <span class="flex h-7 w-7 items-center justify-center rounded-md bg-slate-900 text-white">
            <svg class="h-3.5 w-3.5 fill-current" viewBox="0 0 24 24">
              <path d="M8 5v14l11-7z" />
            </svg>
          </span>
          <span class="text-base font-bold tracking-tight text-slate-900">StreamVerse</span>
          <span class="rounded border border-slate-200 bg-slate-50 px-1.5 py-0.5 text-[10px] font-semibold uppercase text-slate-600">
            {{ appStore.locale === 'ko' ? 'KR' : 'GLOBAL' }}
          </span>
        </NuxtLink>
      </div>

      <!-- Center: Search Bar with Autocomplete -->
      <div class="relative mx-4 flex max-w-xl flex-1 items-center">
        <form class="flex w-full gap-1.5" @submit.prevent="submitSearch()">
          <div class="relative flex flex-1 items-center">
            <input
              v-model="searchQuery"
              type="search"
              :placeholder="appStore.locale === 'ko' ? '동영상, 쇼츠, 라이브, 채널 검색...' : 'Search videos, Shorts, live streams, channels...'"
              class="h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-900 placeholder:text-slate-400 focus:border-slate-900 focus:outline-none focus:ring-1 focus:ring-slate-900"
              @focus="showSuggestions = true; onSearchInput()"
              @input="onSearchInput"
              @blur="onSearchBlur"
            />
          </div>
          <button
            type="submit"
            class="inline-flex h-9 items-center justify-center rounded-md border border-slate-300 bg-white px-3.5 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-50"
            aria-label="Search"
          >
            <svg class="h-4 w-4 text-slate-600" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <circle cx="11" cy="11" r="8" />
              <path d="m21 21-4.35-4.35" />
            </svg>
          </button>
        </form>

        <!-- Search Autocomplete Dropdown -->
        <div
          v-if="showSuggestions && suggestions.length"
          class="absolute left-0 right-0 top-10 z-50 overflow-hidden rounded-lg border border-slate-200 bg-white py-1 shadow-md"
        >
          <button
            v-for="item in suggestions"
            :key="item"
            type="button"
            class="flex w-full items-center gap-2.5 px-3 py-2 text-left text-sm text-slate-700 transition-colors hover:bg-slate-50 hover:text-slate-900"
            @mousedown.prevent="submitSearch(item)"
          >
            <svg class="h-4 w-4 text-slate-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <circle cx="11" cy="11" r="8" />
              <path d="m21 21-4.35-4.35" />
            </svg>
            <span>{{ item }}</span>
          </button>
        </div>
      </div>

      <!-- Right: Controls (Locale, Currency, Console Links, User Menu) -->
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="inline-flex h-8 items-center rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 transition-colors hover:bg-slate-50"
          @click="appStore.toggleLocale()"
        >
          {{ appStore.locale === 'ko' ? '한국어' : 'EN' }}
        </button>

        <button
          type="button"
          class="inline-flex h-8 items-center rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 tabular-nums transition-colors hover:bg-slate-50"
          @click="appStore.toggleCurrency()"
        >
          {{ appStore.currency }}
        </button>

        <a
          :href="config.public.studioUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hidden h-8 items-center gap-1.5 rounded-md bg-slate-900 px-3 text-xs font-medium text-white transition-colors hover:bg-slate-800 sm:inline-flex"
        >
          <span>Studio</span>
          <svg class="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M7 17L17 7M7 7h10v10" /></svg>
        </a>

        <a
          :href="config.public.partnersUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hidden h-8 items-center gap-1 rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 transition-colors hover:bg-slate-50 lg:inline-flex"
        >
          <span>Partners</span>
          <svg class="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M7 17L17 7M7 7h10v10" /></svg>
        </a>

        <a
          :href="config.public.adminUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hidden h-8 items-center gap-1 rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 transition-colors hover:bg-slate-50 xl:inline-flex"
        >
          <span>Admin</span>
          <svg class="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M7 17L17 7M7 7h10v10" /></svg>
        </a>

        <div class="relative">
          <button
            v-if="appStore.currentUser"
            type="button"
            class="flex items-center gap-2 rounded-full border border-slate-200 p-0.5 transition hover:border-slate-400"
            @click="showUserMenu = !showUserMenu"
          >
            <img
              :src="appStore.currentUser?.avatarUrl"
              :alt="appStore.currentUser?.displayName"
              class="h-7 w-7 rounded-full object-cover"
            />
          </button>
          <NuxtLink
            v-else
            to="/login"
            class="inline-flex h-8 items-center rounded-md border border-slate-300 bg-white px-3 text-xs font-medium text-slate-900 transition-colors hover:bg-slate-50"
          >
            {{ appStore.locale === 'ko' ? '로그인' : 'Sign in' }}
          </NuxtLink>

          <div
            v-if="showUserMenu && appStore.currentUser"
            class="absolute right-0 top-10 z-50 w-60 rounded-lg border border-slate-200 bg-white p-2 text-slate-900 shadow-md"
          >
            <div class="border-b border-slate-100 px-2.5 py-2">
              <p class="text-sm font-semibold text-slate-900">{{ appStore.currentUser?.displayName }}</p>
              <p class="text-xs text-slate-500">@{{ appStore.currentUser?.username }} · {{ appStore.currentUser?.subscriptionTier?.toUpperCase() }}</p>
            </div>
            <div class="mt-1 flex flex-col gap-0.5 text-sm">
              <NuxtLink to="/channels/ch_tech_kr" class="rounded-md px-2.5 py-1.5 text-slate-700 hover:bg-slate-50 hover:text-slate-900" @click="showUserMenu = false">
                {{ appStore.locale === 'ko' ? '내 채널 보기' : 'Your Channel' }}
              </NuxtLink>
              <NuxtLink to="/premium" class="rounded-md px-2.5 py-1.5 font-medium text-blue-600 hover:bg-slate-50" @click="showUserMenu = false">
                {{ appStore.locale === 'ko' ? 'StreamVerse Premium 멤버십' : 'StreamVerse Premium' }}
              </NuxtLink>
              <button
                type="button"
                class="rounded-md px-2.5 py-1.5 text-left text-slate-600 hover:bg-slate-50 hover:text-slate-900"
                @click="logout"
              >
                {{ appStore.locale === 'ko' ? '로그아웃' : 'Sign out' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- Body Layout -->
    <div class="flex flex-1">
      <!-- Left Collapsible Sidebar -->
      <aside
        class="sticky top-14 hidden h-[calc(100vh-3.5rem)] shrink-0 overflow-y-auto border-r border-slate-200 bg-slate-50 transition-all md:block"
        :class="[appStore.sidebarCollapsed ? 'w-16 px-2' : 'w-56 px-3']"
      >
        <nav class="mt-3 flex flex-col gap-0.5">
          <NuxtLink
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors"
            :class="[
              route.path === item.path
                ? 'bg-slate-900 text-white font-semibold'
                : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'
            ]"
          >
            <svg class="h-4 w-4 shrink-0" viewBox="0 0 24 24" fill="currentColor">
              <path :d="item.icon" />
            </svg>
            <span v-if="!appStore.sidebarCollapsed" class="truncate">{{ item.label }}</span>
          </NuxtLink>
        </nav>

        <!-- Subscribed Channels -->
        <div v-if="!appStore.sidebarCollapsed" class="mt-5 border-t border-slate-200 pt-4">
          <h3 class="px-3 text-xs font-semibold uppercase tracking-wider text-slate-400">
            {{ appStore.locale === 'ko' ? '구독 채널' : 'Subscriptions' }}
          </h3>
          <div class="mt-2 flex flex-col gap-0.5">
            <NuxtLink
              v-for="sub in subscribedChannels"
              :key="sub.id"
              :to="`/channels/${sub.id}`"
              class="flex items-center justify-between rounded-md px-3 py-1.5 text-sm text-slate-700 transition-colors hover:bg-slate-100 hover:text-slate-900"
            >
              <div class="flex items-center gap-2.5 truncate">
                <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full border border-slate-200 bg-white text-xs font-semibold text-slate-700">
                  {{ sub.name.slice(0, 1) }}
                </span>
                <span class="truncate">{{ sub.name }}</span>
              </div>
              <span
                v-if="sub.isLive"
                class="rounded border border-red-200 bg-red-50 px-1 py-0.5 text-[10px] font-semibold text-red-700"
              >
                LIVE
              </span>
            </NuxtLink>
          </div>
        </div>
      </aside>

      <!-- Main Content -->
      <main class="min-w-0 flex-1 bg-white">
        <slot />
        <footer class="mt-12 border-t border-slate-200 bg-slate-50 px-4 py-6 sm:px-6">
          <div class="mx-auto max-w-7xl">
            <nav class="flex flex-wrap items-center gap-x-6 gap-y-2 text-xs text-slate-600" aria-label="서비스 정보">
              <span class="font-bold text-slate-900">StreamVerse</span>
              <NuxtLink to="/terms" class="hover:text-slate-900 hover:underline">이용약관</NuxtLink>
              <NuxtLink to="/privacy" class="hover:text-slate-900 hover:underline">개인정보처리방침</NuxtLink>
              <NuxtLink to="/premium" class="hover:text-slate-900 hover:underline">Premium</NuxtLink>
              <a :href="config.public.studioUrl" target="_blank" rel="noopener noreferrer" class="hover:text-slate-900 hover:underline">크리에이터 스튜디오</a>
            </nav>
            <p class="mt-2 text-xs text-slate-400">© 2026 StreamVerse Global Video Platform. 모든 영상·채널 데이터는 계정 삭제 요청 시 30일 유예 후 파기됩니다.</p>
          </div>
        </footer>
      </main>
    </div>
  </div>
</template>
