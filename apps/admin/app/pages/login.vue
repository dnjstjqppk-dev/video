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

const route = useRoute()
const AUTH_ERROR_MESSAGES: Record<string, string> = {
  oidc_unconfigured: 'OIDC 로그인이 설정되지 않은 환경입니다. 로컬에서는 아래 개발 토큰 로그인을 이용하세요.',
  oidc_denied: 'OIDC 인증이 거부되었습니다. 다시 시도하세요.',
  oauth_state: '로그인 상태가 만료되었거나 유효하지 않습니다. 다시 시도하세요.',
  token_exchange: 'OIDC 토큰 교환에 실패했습니다. 잠시 후 다시 시도하세요.'
}
const authErrorMessage = computed(() => {
  const code = String(route.query.error || '')
  if (!code) return null
  return AUTH_ERROR_MESSAGES[code] || '로그인 처리 중 오류가 발생했습니다.'
})

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
    <div class="w-full rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-8">
      <div class="flex items-center gap-2.5">
        <span class="flex h-8 w-8 items-center justify-center rounded-md bg-[var(--gvp-brand)] text-sm font-bold text-white">
          A
        </span>
        <span class="text-base font-bold tracking-tight text-[var(--gvp-text)]">Admin Console</span>
      </div>
      <h1 class="mt-5 text-xl font-bold tracking-tight text-[var(--gvp-text)]">관리자 로그인</h1>
      <p class="mt-1.5 text-xs text-[var(--gvp-text-secondary)] leading-relaxed">
        OIDC + 관리망 IP 허용 + 2단계 인증으로 보호되는 거버넌스 콘솔입니다.
      </p>

      <div v-if="authErrorMessage" class="mt-4 rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] px-3 py-2 text-xs font-medium text-[var(--gvp-live)]" role="alert">
        {{ authErrorMessage }}
      </div>

      <a
        href="/auth/login"
        class="mt-6 inline-flex w-full items-center justify-center h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors"
      >
        관리자 계정으로 계속하기
      </a>

      <ClientOnly>
      <div v-if="isLocalhost" class="mt-6 border-t border-[var(--gvp-border)] pt-5">
        <p class="text-xs font-semibold text-[var(--gvp-text)]">로컬 개발 전용 토큰 로그인</p>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          예: <span class="font-mono text-[var(--gvp-text)]">admin-token</span> (SUPER_ADMIN)
        </p>
        <div class="mt-3 flex gap-2">
          <input
            v-model="devToken"
            type="text"
            placeholder="dev token"
            class="h-9 flex-1 rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3 font-mono text-xs text-[var(--gvp-text)] placeholder:text-[var(--gvp-text-muted)] focus:border-[var(--gvp-focus)] focus:bg-[var(--gvp-surface)] outline-none transition-colors"
            @keyup.enter="loginWithDevToken"
          />
          <button
            type="button"
            :disabled="devLoading"
            class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors disabled:opacity-50"
            @click="loginWithDevToken"
          >
            {{ devLoading ? '...' : '로그인' }}
          </button>
        </div>
        <p v-if="devError" class="mt-2 text-xs font-medium text-[var(--gvp-live)]">{{ devError }}</p>
      </div>
      </ClientOnly>
    </div>
  </div>
</template>
