<script setup lang="ts">
import type { UserProfile } from '@gvp/shared-types'
import { UserRole } from '@gvp/shared-types'

const api = useAdminApi()
const sessionUser = useState<UserProfile | null>('admin-session', () => null)

useSeoMeta({
  title: '로그인 — Admin Governance Console'
})

const adminRoles: string[] = [
  UserRole.SUPER_ADMIN,
  UserRole.PLATFORM_ADMIN,
  UserRole.FINANCE_ADMIN,
  UserRole.MODERATOR,
  UserRole.SUPPORT_AGENT
]

const devToken = ref('')
const devError = ref<string | null>(null)
const devLoading = ref(false)
const isLocalhost = ref(false)

if (import.meta.client) {
  const host = window.location.hostname
  isLocalhost.value = host === 'localhost' || host === '127.0.0.1'
}

async function loginWithDevToken() {
  if (!devToken.value.trim()) {
    devError.value = '토큰을 입력하세요.'
    return
  }
  devError.value = null
  devLoading.value = true
  try {
    await $fetch('/api/auth/dev-token', {
      method: 'POST',
      body: { token: devToken.value.trim() }
    })
    const user = await api.getMe()
    if (!adminRoles.includes(user.role)) {
      devError.value = `관리자 권한이 없는 계정입니다 (role: ${user.role}).`
      await $fetch('/auth/logout', { method: 'POST' }).catch(() => {})
      sessionUser.value = null
      return
    }
    sessionUser.value = user
    await navigateTo('/')
  } catch (err: any) {
    devError.value = err?.data?.statusMessage || err?.message || '로그인에 실패했습니다.'
  } finally {
    devLoading.value = false
  }
}
</script>

<template>
  <div class="mx-auto flex min-h-[60vh] max-w-md flex-col items-center justify-center px-4">
    <div class="w-full rounded-2xl border border-red-500/25 bg-[#10131C] p-8 shadow-2xl">
      <div class="flex items-center gap-2">
        <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-red-600 font-black text-white">
          A
        </span>
        <span class="text-lg font-extrabold text-white">Admin Console</span>
      </div>
      <h1 class="mt-4 text-xl font-black text-white">관리자 로그인</h1>
      <p class="mt-1 text-xs text-neutral-400">
        OIDC + 관리망 IP 허용 + 2단계 인증으로 보호되는 거버넌스 콘솔입니다.
      </p>

      <a
        href="/auth/login"
        class="mt-6 flex w-full items-center justify-center gap-2 rounded-full bg-red-600 px-6 py-3 text-sm font-bold text-white transition hover:bg-red-500"
      >
        관리자 계정으로 계속하기
      </a>

      <div v-if="isLocalhost" class="mt-6 border-t border-white/10 pt-5">
        <p class="text-xs font-bold text-amber-300">로컬 개발 전용 토큰 로그인</p>
        <p class="mt-1 text-[11px] text-neutral-400">
          예: admin-token (SUPER_ADMIN)
        </p>
        <div class="mt-3 flex gap-2">
          <input
            v-model="devToken"
            type="text"
            placeholder="dev token"
            class="h-10 flex-1 rounded-xl border border-white/15 bg-[#090B10] px-3 font-mono text-xs text-white outline-none focus:border-red-500"
            @keyup.enter="loginWithDevToken"
          />
          <button
            type="button"
            :disabled="devLoading"
            class="rounded-xl bg-white/10 px-4 text-xs font-bold text-white hover:bg-white/20 disabled:opacity-50"
            @click="loginWithDevToken"
          >
            {{ devLoading ? '...' : '로그인' }}
          </button>
        </div>
        <p v-if="devError" class="mt-2 text-xs text-red-400">{{ devError }}</p>
      </div>
    </div>
  </div>
</template>
