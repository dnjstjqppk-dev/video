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
  { path: '/', label: '대시보드 (Dashboard)', icon: '📊' },
  { path: '/content', label: '콘텐츠 관리 (Content)', icon: '🎬' },
  { path: '/upload', label: '멀티파트 업로드 (Upload)', icon: '⬆️' },
  { path: '/editor/vid_arch_2026', label: '타임라인 에디터 (Editor)', icon: '✂️' },
  { path: '/analytics', label: '채널 분석 (Analytics)', icon: '📈' },
  { path: '/monetization', label: '수익 창출 & 정산 (Earn)', icon: '💰' },
  { path: '/customization', label: '채널 맞춤설정 (Customization)', icon: '🎨' },
  { path: '/comments', label: '댓글 및 검토 (Comments)', icon: '💬' },
  { path: '/subtitles', label: '자막 관리 (Subtitles)', icon: '📝' },
  { path: '/copyright', label: '저작권 및 이의제기 (Copyright)', icon: '⚖️' },
  { path: '/live', label: '라이브 관제실 (Live Room)', icon: '📡' }
]
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <!-- Top Studio Header -->
    <header class="sticky top-0 z-40 flex h-16 items-center justify-between border-b border-white/10 bg-[#12151E]/95 px-4 backdrop-blur sm:px-6">
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
        <NuxtLink to="/" class="flex items-center gap-2.5">
          <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-[#FF0033] font-black text-white">
            S
          </span>
          <div>
            <span class="text-base font-extrabold tracking-tight text-white">Creator Studio</span>
            <span
              v-if="sessionUser"
              class="ml-2 rounded bg-emerald-500/20 px-1.5 py-0.5 text-[10px] font-bold text-emerald-400"
            >
              {{ sessionUser.role }}
            </span>
            <NuxtLink
              v-else
              to="/login"
              class="ml-2 rounded bg-amber-500/20 px-1.5 py-0.5 text-[10px] font-bold text-amber-300 hover:bg-amber-500/30"
            >
              LOGIN REQUIRED
            </NuxtLink>
          </div>
        </NuxtLink>
      </div>

      <div class="flex min-w-0 items-center gap-2 overflow-x-auto sm:gap-3">
        <span v-if="sessionUser" class="hidden text-xs text-neutral-300 md:inline">
          {{ sessionUser.displayName }}
        </span>
        <button
          v-if="sessionUser"
          type="button"
          class="rounded-xl border border-white/15 bg-[#1B1F2B] px-3 py-1.5 text-xs font-bold text-neutral-300 hover:bg-white/10"
          @click="logout"
        >
          로그아웃
        </button>
        <NuxtLink
          v-else
          to="/login"
          class="rounded-xl bg-[#FF0033] px-4 py-2 text-xs font-bold text-white hover:bg-[#D9002B]"
        >
          로그인
        </NuxtLink>

        <!-- Currency Switcher -->
        <button
          type="button"
          class="rounded-xl border border-white/15 bg-[#1B1F2B] px-3 py-1.5 text-xs font-bold text-[#3EA6FF] hover:bg-white/10"
          @click="currency = currency === 'KRW' ? 'USD' : 'KRW'"
        >
          {{ currency === 'KRW' ? '₩ KRW' : '$ USD' }}
        </button>

        <NuxtLink
          to="/upload"
          class="flex items-center gap-1.5 rounded-xl bg-[#FF0033] px-4 py-2 text-xs font-bold text-white shadow transition hover:bg-[#D9002B]"
        >
          <span>＋</span>
          <span class="hidden sm:inline">새 동영상 업로드 (Upload)</span>
        </NuxtLink>

        <a
          :href="config.public.webUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="rounded-xl border border-white/15 px-3 py-2 text-xs font-semibold text-neutral-300 hover:bg-white/10"
        >
          Public Channel ↗
        </a>

        <a
          :href="config.public.partnersUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hidden rounded-xl border border-[#3EA6FF]/30 px-3 py-2 text-xs font-semibold text-[#3EA6FF] hover:bg-white/10 md:inline-block"
        >
          Partners ↗
        </a>

        <a
          :href="config.public.adminUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hidden rounded-xl border border-red-500/30 px-3 py-2 text-xs font-semibold text-red-400 hover:bg-white/10 lg:inline-block"
        >
          Admin ↗
        </a>
      </div>
    </header>

    <div
      v-if="mobileNavOpen"
      class="fixed inset-0 z-40 bg-black/60 md:hidden"
      aria-hidden="true"
      @click="mobileNavOpen = false"
    />
    <div class="flex flex-1">
      <!-- Studio Left Navigation Sidebar -->
      <aside
        class="fixed inset-y-0 left-0 z-50 w-72 shrink-0 overflow-y-auto border-r border-white/10 bg-[#12151E] p-4 transition-transform md:static md:z-auto md:w-64 md:translate-x-0"
        :class="mobileNavOpen ? 'translate-x-0' : '-translate-x-full'"
      >
        <!-- Creator Channel Badge -->
        <div class="mb-6 flex items-center gap-3 rounded-2xl border border-white/10 bg-[#181C28] p-3">
          <img
            v-if="sessionUser?.avatarUrl"
            :src="sessionUser.avatarUrl"
            :alt="sessionUser.displayName"
            class="h-11 w-11 rounded-full object-cover ring-2 ring-[#FF0033]"
          />
          <span
            v-else
            class="flex h-11 w-11 items-center justify-center rounded-full bg-[#FF0033]/20 text-sm font-black text-[#FF0033] ring-2 ring-[#FF0033]"
          >
            ?
          </span>
          <div class="min-w-0">
            <p class="truncate text-xs font-bold text-white">{{ sessionUser?.displayName || '미로그인 크리에이터' }}</p>
            <p class="text-[11px] text-neutral-400">{{ sessionUser ? `@${sessionUser.username}` : '로그인 후 채널 표시' }}</p>
          </div>
        </div>

        <nav class="space-y-1">
          <NuxtLink
            v-for="item in navLinks"
            :key="item.path"
            :to="item.path"
            class="flex items-center gap-3 rounded-xl px-3.5 py-2.5 text-xs font-semibold transition"
            :class="
              route.path === item.path
                ? 'bg-[#FF0033] text-white shadow'
                : 'text-neutral-300 hover:bg-white/10 hover:text-white'
            "
          >
            <span class="text-base">{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </NuxtLink>
        </nav>

        <div class="mt-8 rounded-2xl border border-white/10 bg-[#181C28] p-3.5 text-xs">
          <p class="font-bold text-white">Subtitles & Copyright Status</p>
          <p class="mt-1 text-[11px] text-emerald-400">✓ 0 Active Copyright Strikes</p>
          <p class="mt-0.5 text-[11px] text-neutral-400">✓ Auto-captions (ko/en) enabled</p>
        </div>
      </aside>

            <!-- Main Studio Workspace -->
      <main class="min-w-0 flex-1 p-6 lg:p-8">
        <div
          v-if="!sessionUser && route.path !== '/login'"
          class="mb-6 flex items-center justify-between rounded-xl border border-amber-500/30 bg-amber-500/10 px-4 py-2.5 text-xs text-amber-300"
        >
          <div class="flex items-center gap-2">
            <span class="font-bold">로그인이 필요합니다</span>
            <span class="text-neutral-400">|</span>
            <span class="text-neutral-300">실제 채널 데이터를 보려면 로그인하세요. 미인증 상태에서는 샘플 화면이 표시될 수 있습니다.</span>
          </div>
          <NuxtLink to="/login" class="rounded bg-amber-500/20 px-2 py-0.5 font-mono text-[10px] text-amber-200 hover:bg-amber-500/30">
            LOGIN
          </NuxtLink>
        </div>
        <slot />
        <footer class="mt-12 border-t border-white/5 pt-4 text-center text-[11px] text-neutral-400">
          <span>StreamVerse Creator Studio</span>
        </footer>
      </main>
    </div>
  </div>
</template>
