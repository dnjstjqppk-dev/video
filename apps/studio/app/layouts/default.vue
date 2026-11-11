<script setup lang="ts">
import type { UserProfile } from '@gvp/shared-types'

const route = useRoute()
const config = useRuntimeConfig()
const api = useStudioApi()

const currency = useState<'KRW' | 'USD'>('studio-currency', () => 'KRW')
const sessionUser = useState<UserProfile | null>('studio-session', () => null)
const mobileNavOpen = ref(false)

watch(() => route.path, () => {
  mobileNavOpen.value = false
})

onMounted(async () => {
  if (sessionUser.value) return
  try {
    sessionUser.value = await api.getMe()
  } catch {
    sessionUser.value = null
  }
})

async function logout() {
  try {
    await $fetch('/auth/logout', { method: 'POST' })
  } catch {
    // Session cookies are cleared server-side; continue locally regardless.
  }
  const tokenCookie = useCookie('gvp_studio_token')
  tokenCookie.value = null
  sessionUser.value = null
  await navigateTo('/login')
}

const navLinks = [
  { path: '/', label: '대시보드 (Dashboard)', icon: 'M3 3h7v7H3V3zm11 0h7v7h-7V3zm0 11h7v7h-7v-7zM3 14h7v7H3v-7z' },
  { path: '/content', label: '콘텐츠 관리 (Content)', icon: 'M4 6h16v12H4V6zm2 2v8h12V8H6zm4 2l5 2-5 2v-4z' },
  { path: '/upload', label: '멀티파트 업로드 (Upload)', icon: 'M12 16V4m0 0L8 8m4-4l4 4M4 17v2a1 1 0 001 1h14a1 1 0 001-1v-2' },
  { path: '/editor/vid_arch_2026', label: '타임라인 에디터 (Editor)', icon: 'M4 7h16M4 12h16M4 17h16M8 4v16M16 4v16' },
  { path: '/analytics', label: '채널 분석 (Analytics)', icon: 'M3 20h18M6 16v-4m4 4V8m4 8v-6m4 6V4' },
  { path: '/monetization', label: '수익 창출 & 정산 (Earn)', icon: 'M12 2v20M17 5H9.5a3.5 3.5 0 000 7h5a3.5 3.5 0 010 7H6' },
  { path: '/customization', label: '채널 맞춤설정 (Customization)', icon: 'M12 20h9M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z' },
  { path: '/comments', label: '댓글 및 검토 (Comments)', icon: 'M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2v10z' },
  { path: '/subtitles', label: '자막 관리 (Subtitles)', icon: 'M4 6h16v12H4V6zm3 5h4m2 0h4m-10 3h6m2 0h2' },
  { path: '/copyright', label: '저작권 및 이의제기 (Copyright)', icon: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z' },
  { path: '/live', label: '라이브 관제실 (Live Room)', icon: 'M12 8a4 4 0 100 8 4 4 0 000-8zm-7 4a7 7 0 0114 0' }
]
</script>

<template>
  <div class="flex min-h-screen flex-col bg-white text-slate-900">
    <!-- Top Studio Header -->
    <header class="sticky top-0 z-40 flex h-14 items-center justify-between border-b border-slate-200 bg-white px-4 sm:px-6">
      <div class="flex items-center gap-3">
        <button
          type="button"
          class="inline-flex h-9 w-9 items-center justify-center rounded-md text-slate-600 hover:bg-slate-100 md:hidden"
          aria-label="메뉴 열기"
          :aria-expanded="mobileNavOpen"
          @click="mobileNavOpen = !mobileNavOpen"
        >
          <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <path stroke-linecap="round" d="M4 7h16M4 12h16M4 17h16" />
          </svg>
        </button>
        <NuxtLink to="/" class="flex items-center gap-2.5">
          <span class="flex h-7 w-7 items-center justify-center rounded-md bg-slate-900 text-xs font-bold text-white">
            S
          </span>
          <div class="flex items-center gap-2">
            <span class="text-sm font-bold tracking-tight text-slate-900">Creator Studio</span>
            <span
              v-if="sessionUser"
              class="rounded-md border border-green-200 bg-green-50 px-2 py-0.5 text-xs font-medium text-green-700"
            >
              {{ sessionUser.role }}
            </span>
            <NuxtLink
              v-else
              to="/login"
              class="rounded-md border border-amber-200 bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-700 hover:bg-amber-100"
            >
              LOGIN REQUIRED
            </NuxtLink>
          </div>
        </NuxtLink>
      </div>

      <div class="flex min-w-0 items-center gap-2 overflow-x-auto sm:gap-2.5">
        <span v-if="sessionUser" class="hidden text-xs font-medium text-slate-600 md:inline">
          {{ sessionUser.displayName }}
        </span>
        <button
          v-if="sessionUser"
          type="button"
          class="inline-flex h-8 items-center rounded-md border border-slate-300 bg-white px-3 text-xs font-medium text-slate-700 hover:bg-slate-50"
          @click="logout"
        >
          로그아웃
        </button>
        <NuxtLink
          v-else
          to="/login"
          class="inline-flex h-8 items-center rounded-md border border-slate-300 bg-white px-3 text-xs font-medium text-slate-700 hover:bg-slate-50"
        >
          로그인
        </NuxtLink>

        <!-- Currency Switcher -->
        <button
          type="button"
          class="inline-flex h-8 items-center rounded-md border border-slate-300 bg-white px-2.5 text-xs font-medium text-slate-700 tabular-nums hover:bg-slate-50"
          @click="currency = currency === 'KRW' ? 'USD' : 'KRW'"
        >
          {{ currency === 'KRW' ? '₩ KRW' : '$ USD' }}
        </button>

        <NuxtLink
          to="/upload"
          class="inline-flex h-8 items-center gap-1.5 rounded-md bg-slate-900 px-3 text-xs font-medium text-white transition-colors hover:bg-slate-800"
        >
          <span>+</span>
          <span class="hidden sm:inline">동영상 업로드</span>
        </NuxtLink>

        <a
          :href="config.public.webUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex h-8 items-center rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-600 hover:bg-slate-50 hover:text-slate-900"
        >
          Web Portal
        </a>

        <a
          :href="config.public.partnersUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hidden h-8 items-center rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-600 hover:bg-slate-50 hover:text-slate-900 md:inline-flex"
        >
          Partners
        </a>

        <a
          :href="config.public.adminUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hidden h-8 items-center rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-600 hover:bg-slate-50 hover:text-slate-900 lg:inline-flex"
        >
          Admin
        </a>
      </div>
    </header>

    <div
      v-if="mobileNavOpen"
      class="fixed inset-0 z-40 bg-slate-900/40 md:hidden"
      aria-hidden="true"
      @click="mobileNavOpen = false"
    />
    <div class="flex flex-1">
      <!-- Studio Left Navigation Sidebar -->
      <aside
        class="fixed inset-y-0 left-0 z-50 w-64 shrink-0 overflow-y-auto border-r border-slate-200 bg-slate-50 p-3 transition-transform md:static md:z-auto md:w-60 md:translate-x-0"
        :class="mobileNavOpen ? 'translate-x-0' : '-translate-x-full'"
      >
        <!-- Creator Channel Badge -->
        <div class="mb-4 flex items-center gap-3 rounded-lg border border-slate-200 bg-white p-3">
          <img
            v-if="sessionUser?.avatarUrl"
            :src="sessionUser.avatarUrl"
            :alt="sessionUser.displayName"
            class="h-9 w-9 rounded-full border border-slate-200 object-cover"
          />
          <span
            v-else
            class="flex h-9 w-9 items-center justify-center rounded-full border border-slate-200 bg-slate-100 text-xs font-bold text-slate-700"
          >
            CH
          </span>
          <div class="min-w-0">
            <p class="truncate text-xs font-semibold text-slate-900">{{ sessionUser?.displayName || '미로그인 크리에이터' }}</p>
            <p class="text-[11px] text-slate-500">{{ sessionUser ? `@${sessionUser.username}` : '로그인 후 채널 표시' }}</p>
          </div>
        </div>

        <nav class="space-y-0.5">
          <NuxtLink
            v-for="item in navLinks"
            :key="item.path"
            :to="item.path"
            class="flex items-center gap-2.5 rounded-md px-3 py-2 text-xs font-medium transition-colors"
            :class="
              route.path === item.path
                ? 'bg-slate-900 text-white font-semibold'
                : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'
            "
          >
            <svg class="h-4 w-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" :d="item.icon" />
            </svg>
            <span class="truncate">{{ item.label }}</span>
          </NuxtLink>
        </nav>

        <div class="mt-6 rounded-lg border border-slate-200 bg-white p-3 text-xs">
          <p class="font-semibold text-slate-900">Subtitles & Copyright Status</p>
          <p class="mt-1 text-green-700">0 Active Copyright Strikes</p>
          <p class="mt-0.5 text-slate-500">Auto-captions (ko/en) enabled</p>
        </div>
      </aside>

      <!-- Main Studio Workspace -->
      <main class="min-w-0 flex-1 bg-white p-4 sm:p-6 lg:p-8">
        <div class="mx-auto max-w-7xl">
          <div
            v-if="!sessionUser && route.path !== '/login'"
            class="mb-6 flex items-center justify-between rounded-md border border-amber-200 bg-amber-50 px-4 py-2.5 text-xs text-amber-800"
          >
            <div class="flex items-center gap-2">
              <span class="font-semibold">로그인이 필요합니다</span>
              <span class="text-amber-400">|</span>
              <span>실제 채널 데이터를 보려면 로그인하세요. 미인증 상태에서는 샘플 화면이 표시될 수 있습니다.</span>
            </div>
            <NuxtLink to="/login" class="rounded border border-amber-300 bg-white px-2.5 py-1 font-mono text-[11px] font-medium text-amber-800 hover:bg-amber-100">
              LOGIN
            </NuxtLink>
          </div>
          <slot />
          <footer class="mt-10 border-t border-slate-200 pt-4 text-center text-xs text-slate-400">
            <span>StreamVerse Creator Studio</span>
          </footer>
        </div>
      </main>
    </div>
  </div>
</template>
