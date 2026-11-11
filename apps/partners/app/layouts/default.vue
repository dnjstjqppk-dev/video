<script setup lang="ts">
import type { UserProfile } from '@gvp/shared-types'

const route = useRoute()
const config = useRuntimeConfig()
const api = usePartnerApi()

const sessionUser = useState<UserProfile | null>('partner-session', () => null)
const mobileNavOpen = ref(false)

watch(() => route.path, () => {
  mobileNavOpen.value = false
})
const activeOrg = useState<{ id: string; name: string; type: string }>('partner-org', () => ({
  id: '',
  name: '',
  type: ''
}))

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
  const tokenCookie = useCookie('gvp_partner_token')
  tokenCookie.value = null
  sessionUser.value = null
  await navigateTo('/login')
}

const navItems = [
  { path: '/', label: '파트너 종합 현황 (Overview)', icon: 'M3 3h7v7H3V3zm11 0h7v7h-7V3zm0 11h7v7h-7v-7zM3 14h7v7H3v-7z' },
  { path: '/advertiser', label: '독립 광고주 포털 (Advertiser)', icon: 'M12 2L2 7l10 5 10-5-10-5zm0 15l-10-5v5l10 5 10-5v-5l-10 5z' },
  { path: '/campaigns', label: '광고 캠페인 매니저 (Campaigns)', icon: 'M3 11l18-5v12L3 14v-3zm0 3l4 6h2l-3-6' },
  { path: '/mcn', label: 'MCN 소속 채널 관리 (MCN)', icon: 'M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2m8-10a4 4 0 100-8 4 4 0 000 8zm14 10v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75' },
  { path: '/rights', label: '저작권 & 핑거프린트 (Rights ID)', icon: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z' },
  { path: '/settlements', label: '복식부기 정산서 (Settlements)', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9h6m-6 4h6' }
]
</script>

<template>
  <div class="flex min-h-screen flex-col bg-white text-slate-900">
    <!-- Top Multi-Tenant Partner Header -->
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
        <span class="flex h-7 w-7 items-center justify-center rounded-md bg-slate-900 text-xs font-bold text-white">
          P
        </span>
        <div class="flex items-center gap-2">
          <span class="text-sm font-bold tracking-tight text-slate-900">StreamVerse Partner Console</span>
          <span
            v-if="activeOrg.name"
            class="rounded-md border border-blue-200 bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700"
          >
            {{ activeOrg.name }}
          </span>
        </div>
      </div>

      <div class="flex items-center gap-2.5 text-xs">
        <template v-if="sessionUser">
          <span class="hidden font-medium text-slate-600 md:inline">{{ sessionUser.displayName }}</span>
          <span class="rounded-md border border-slate-200 bg-slate-50 px-2.5 py-1 font-mono font-medium text-slate-700">
            {{ sessionUser.role }}
          </span>
          <button
            type="button"
            class="inline-flex h-8 items-center rounded-md border border-slate-300 bg-white px-3 font-medium text-slate-700 hover:bg-slate-50"
            @click="logout"
          >
            로그아웃
          </button>
        </template>
        <NuxtLink
          v-else
          to="/login"
          class="inline-flex h-8 items-center rounded-md bg-slate-900 px-3 font-medium text-white hover:bg-slate-800"
        >
          파트너 로그인
        </NuxtLink>
      </div>
    </header>

    <div
      v-if="mobileNavOpen"
      class="fixed inset-0 z-40 bg-slate-900/40 md:hidden"
      aria-hidden="true"
      @click="mobileNavOpen = false"
    />
    <div class="flex flex-1">
      <!-- Left Sidebar -->
      <aside
        class="fixed inset-y-0 left-0 z-50 w-64 shrink-0 overflow-y-auto border-r border-slate-200 bg-slate-50 p-3 transition-transform md:static md:z-auto md:w-60 md:translate-x-0"
        :class="mobileNavOpen ? 'translate-x-0' : '-translate-x-full'"
      >
        <div class="mb-4 rounded-lg border border-slate-200 bg-white p-3 text-xs">
          <p class="text-[10px] font-semibold uppercase tracking-wider text-blue-700">Active Organization</p>
          <p class="mt-1 font-semibold text-slate-900">{{ activeOrg.name || '조직 정보를 불러오는 중' }}</p>
          <p v-if="activeOrg.id" class="mt-0.5 font-mono text-[11px] text-slate-500">ID: {{ activeOrg.id }}</p>
        </div>

        <nav class="space-y-0.5">
          <NuxtLink
            v-for="item in navItems"
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
      </aside>

      <!-- Main Partner Content -->
      <main class="min-w-0 flex-1 bg-white p-4 sm:p-6 lg:p-8">
        <div class="mx-auto max-w-7xl">
          <div
            v-if="!sessionUser && route.path !== '/login'"
            class="mb-6 flex items-center justify-between rounded-md border border-amber-200 bg-amber-50 px-4 py-2.5 text-xs text-amber-800"
          >
            <div class="flex items-center gap-2">
              <span class="font-semibold">로그인이 필요합니다</span>
              <span class="text-amber-400">|</span>
              <span>실제 파트너 데이터를 보려면 로그인하세요.</span>
            </div>
            <NuxtLink to="/login" class="rounded border border-amber-300 bg-white px-2.5 py-1 font-mono text-[11px] font-medium text-amber-800 hover:bg-amber-100">
              LOGIN
            </NuxtLink>
          </div>
          <slot />
          <footer class="mt-10 border-t border-slate-200 pt-4 text-center text-xs text-slate-400">
            <span>StreamVerse Partner Console</span>
          </footer>
        </div>
      </main>
    </div>
  </div>
</template>
