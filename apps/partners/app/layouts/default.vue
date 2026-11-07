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
  { path: '/', label: '파트너 종합 현황 (Overview)', icon: '🏢' },
  { path: '/advertiser', label: '독립 광고주 포털 (Advertiser)', icon: '🎯' },
  { path: '/campaigns', label: '광고 캠페인 매니저 (Campaigns)', icon: '📣' },
  { path: '/mcn', label: 'MCN 소속 채널 관리 (MCN)', icon: '🤝' },
  { path: '/rights', label: '저작권 & 핑거프린트 (Rights ID)', icon: '🛡️' },
  { path: '/settlements', label: '복식부기 정산서 (Settlements)', icon: '🧾' }
]
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <!-- Top Multi-Tenant Partner Header -->
    <header class="sticky top-0 z-40 flex h-16 items-center justify-between border-b border-white/10 bg-[#111625]/95 px-4 backdrop-blur sm:px-6">
      <div class="flex items-center gap-3">
        <button
          type="button"
          class="rounded-lg p-2 text-neutral-300 hover:bg-white/10 md:hidden"
          aria-label="메뉴 열기"
          :aria-expanded="mobileNavOpen"
          @click="mobileNavOpen = !mobileNavOpen"
        >
          <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <path stroke-linecap="round" d="M4 7h16M4 12h16M4 17h16" />
          </svg>
        </button>
        <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-[#3EA6FF] font-black text-black">
          P
        </span>
        <div>
          <span class="text-base font-extrabold text-white">StreamVerse Partner Console</span>
          <span
            v-if="activeOrg.name"
            class="ml-2 rounded bg-[#3EA6FF]/20 px-2 py-0.5 text-[10px] font-bold text-[#3EA6FF]"
          >
            {{ activeOrg.name }}
          </span>
        </div>
      </div>

      <div class="flex items-center gap-3 text-xs">
        <template v-if="sessionUser">
          <span class="hidden text-neutral-300 md:inline">{{ sessionUser.displayName }}</span>
          <span class="rounded-xl border border-white/15 bg-[#0B0E17] px-3 py-1.5 font-mono font-bold text-[#3EA6FF]">
            {{ sessionUser.role }}
          </span>
          <button
            type="button"
            class="rounded-xl border border-white/15 px-3 py-1.5 font-bold text-neutral-300 hover:bg-white/10"
            @click="logout"
          >
            로그아웃
          </button>
        </template>
        <NuxtLink
          v-else
          to="/login"
          class="rounded-xl bg-[#3EA6FF] px-4 py-2 font-bold text-black hover:brightness-110"
        >
          파트너 로그인
        </NuxtLink>
      </div>
    </header>

    <div
      v-if="mobileNavOpen"
      class="fixed inset-0 z-40 bg-black/60 md:hidden"
      aria-hidden="true"
      @click="mobileNavOpen = false"
    />
    <div class="flex flex-1">
      <!-- Left Sidebar -->
      <aside
        class="fixed inset-y-0 left-0 z-50 w-72 shrink-0 overflow-y-auto border-r border-white/10 bg-[#111625] p-4 transition-transform md:static md:z-auto md:w-64 md:translate-x-0"
        :class="mobileNavOpen ? 'translate-x-0' : '-translate-x-full'"
      >
        <div class="mb-6 rounded-2xl border border-white/10 bg-[#171E31] p-3.5 text-xs">
          <p class="text-[10px] font-bold uppercase tracking-wider text-[#3EA6FF]">Active Organization</p>
          <p class="mt-1 font-bold text-white">{{ activeOrg.name || '조직 정보를 불러오는 중' }}</p>
          <p v-if="activeOrg.id" class="mt-0.5 font-mono text-[11px] text-neutral-400">ID: {{ activeOrg.id }}</p>
        </div>

        <nav class="space-y-1">
          <NuxtLink
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="flex items-center gap-3 rounded-xl px-3.5 py-2.5 text-xs font-semibold transition"
            :class="
              route.path === item.path
                ? 'bg-[#3EA6FF] text-black font-bold shadow'
                : 'text-neutral-300 hover:bg-white/10 hover:text-white'
            "
          >
            <span>{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </NuxtLink>
        </nav>
      </aside>

      <!-- Main Partner Content -->
      <main class="min-w-0 flex-1 p-6 lg:p-8">
        <div
          v-if="!sessionUser && route.path !== '/login'"
          class="mb-6 flex items-center justify-between rounded-xl border border-amber-500/30 bg-amber-500/10 px-4 py-2.5 text-xs text-amber-300"
        >
          <div class="flex items-center gap-2">
            <span class="font-bold">로그인이 필요합니다</span>
            <span class="text-neutral-400">|</span>
            <span class="text-neutral-300">실제 파트너 데이터를 보려면 로그인하세요.</span>
          </div>
          <NuxtLink to="/login" class="rounded bg-amber-500/20 px-2 py-0.5 font-mono text-[10px] text-amber-200 hover:bg-amber-500/30">
            LOGIN
          </NuxtLink>
        </div>
        <slot />
        <footer class="mt-12 border-t border-white/5 pt-4 text-center text-[11px] text-neutral-400">
          <span>StreamVerse Partner Console</span>
        </footer>
      </main>
    </div>
  </div>
</template>
