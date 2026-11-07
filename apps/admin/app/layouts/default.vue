<script setup lang="ts">
import type { UserProfile } from '@gvp/shared-types'
import { UserRole } from '@gvp/shared-types'

const route = useRoute()
const config = useRuntimeConfig()
const api = useAdminApi()

const sessionUser = useState<UserProfile | null>('admin-session', () => null)
const sessionLoaded = useState<boolean>('admin-session-loaded', () => false)

const adminRoles: string[] = [
  UserRole.SUPER_ADMIN,
  UserRole.PLATFORM_ADMIN,
  UserRole.FINANCE_ADMIN,
  UserRole.MODERATOR,
  UserRole.SUPPORT_AGENT
]

const isAdmin = computed(() => !!sessionUser.value && adminRoles.includes(sessionUser.value.role))
const isLoginRoute = computed(() => route.path === '/login')
const mobileNavOpen = ref(false)

watch(() => route.path, () => {
  mobileNavOpen.value = false
})

onMounted(async () => {
  if (sessionLoaded.value) return
  try {
    sessionUser.value = await api.getMe()
  } catch {
    sessionUser.value = null
  } finally {
    sessionLoaded.value = true
  }
})

async function logout() {
  try {
    await $fetch('/auth/logout', { method: 'POST' })
  } catch {
    // Session cookies are cleared server-side; continue locally regardless.
  }
  const tokenCookie = useCookie('gvp_admin_session')
  tokenCookie.value = null
  sessionUser.value = null
  await navigateTo('/login')
}

const navLinks = [
  { path: '/', label: '플랫폼 통합 관제 (Operations)', icon: '🛰️' },
  { path: '/ads', label: '광고 송출 진단 (Ad Delivery)', icon: '🎯' },
  { path: '/users', label: '사용자 & RBAC 거버넌스 (Users)', icon: '👥' },
  { path: '/moderation', label: '신뢰 및 안전 큐 (Moderation)', icon: '🛡️' },
  { path: '/finance', label: '정산 승인 & 유닛 이코노믹스 (Finance)', icon: '🏦' },
  { path: '/policies', label: '정책 버전 & 킬스위치 (Policies)', icon: '⚙️' },
  { path: '/infrastructure', label: '멀티 리전 K8s & DR 관제 (Infra)', icon: '🌐' }
]
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <!-- High-Security Super Admin Top Bar -->
    <header class="sticky top-0 z-40 flex h-16 items-center justify-between border-b border-red-500/25 bg-[#10131C]/95 px-4 backdrop-blur sm:px-6">
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
        <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-red-600 font-black text-white shadow">
          A
        </span>
        <div>
          <span class="text-base font-extrabold text-white">Super Admin Governance Console</span>
          <span class="ml-2 rounded bg-red-500/20 px-2 py-0.5 font-mono text-[10px] font-bold text-red-300">
            ZERO-TRUST
          </span>
        </div>
      </div>

      <div class="flex items-center gap-3 text-xs">
        <template v-if="sessionUser">
          <span class="text-neutral-400">{{ sessionUser.displayName }}</span>
          <span class="rounded-xl border border-white/15 bg-[#090B10] px-3 py-1.5 font-mono text-xs font-bold text-emerald-400">
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
          class="rounded-xl bg-red-600 px-4 py-1.5 font-bold text-white hover:bg-red-500"
        >
          관리자 로그인
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
      <aside
        class="fixed inset-y-0 left-0 z-50 w-72 shrink-0 overflow-y-auto border-r border-white/10 bg-[#10131C] p-4 transition-transform md:static md:z-auto md:w-64 md:translate-x-0"
        :class="mobileNavOpen ? 'translate-x-0' : '-translate-x-full'"
      >
        <nav class="space-y-1">
          <NuxtLink
            v-for="item in navLinks"
            :key="item.path"
            :to="item.path"
            class="flex items-center gap-3 rounded-xl px-3.5 py-2.5 text-xs font-semibold transition"
            :class="
              route.path === item.path
                ? 'bg-red-600 text-white font-bold shadow'
                : 'text-neutral-300 hover:bg-white/10 hover:text-white'
            "
          >
            <span>{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </NuxtLink>
        </nav>

        <div class="mt-8 rounded-2xl border border-red-500/25 bg-red-950/20 p-3.5 text-[11px] text-red-200">
          <p class="font-bold">🔒 Immutable Audit Trail Active</p>
          <p class="mt-1 text-neutral-400">
            Every administrative mutation is signed and appended to the audit log.
          </p>
        </div>
      </aside>

      <main class="min-w-0 flex-1 p-6 lg:p-8">
        <template v-if="isLoginRoute">
          <slot />
        </template>
        <div v-else-if="!sessionLoaded" class="rounded-xl border border-white/10 bg-white/5 px-4 py-8 text-center text-xs text-neutral-400">
          세션을 확인하는 중입니다...
        </div>
        <div v-else-if="!sessionUser" class="mx-auto mt-10 max-w-md rounded-2xl border border-white/10 bg-[#131722] p-8 text-center">
          <p class="text-sm font-bold text-white">관리자 로그인이 필요합니다</p>
          <p class="mt-2 text-xs text-neutral-400">거버넌스 콘솔은 인증된 관리자만 접근할 수 있습니다.</p>
          <NuxtLink to="/login" class="mt-6 inline-block rounded-full bg-red-600 px-6 py-2.5 text-xs font-bold text-white hover:bg-red-500">
            로그인하기
          </NuxtLink>
        </div>
        <div v-else-if="!isAdmin" class="mx-auto mt-10 max-w-md rounded-2xl border border-red-500/30 bg-red-950/20 p-8 text-center">
          <p class="text-sm font-bold text-white">접근 권한이 없습니다 (403)</p>
          <p class="mt-2 text-xs text-neutral-400">현재 계정({{ sessionUser.role }})에는 관리자 권한이 없습니다.</p>
          <button type="button" class="mt-6 rounded-full bg-white/10 px-6 py-2.5 text-xs font-bold text-white hover:bg-white/20" @click="logout">
            로그아웃
          </button>
        </div>
        <template v-else>
          <slot />
        </template>
        <footer class="mt-12 border-t border-white/5 pt-4 text-center text-[11px] text-neutral-400">
          <span>StreamVerse Admin Console · Zero-Trust Governance</span>
        </footer>
      </main>
    </div>
  </div>
</template>
