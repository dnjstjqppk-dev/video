<script setup lang="ts">
import VravraLogo from '@gvp/ui/src/components/VravraLogo.vue'
import { BRAND } from '@gvp/design-tokens/src/brand'
import { useAppStore } from '~/stores/app'

const appStore = useAppStore()
const api = useApi()

useSeoMeta({
  title: `로그인 — ${BRAND.name} ${BRAND.nameKo} 계정`
})

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
    await appStore.fetchSession(api)
    if (appStore.currentUser) {
      await navigateTo('/')
    } else {
      devError.value = appStore.authError || '세션을 확인할 수 없습니다.'
    }
  } catch (err: any) {
    devError.value = err?.data?.statusMessage || err?.message || '로그인에 실패했습니다.'
  } finally {
    devLoading.value = false
  }
}
</script>

<template>
  <div class="mx-auto flex min-h-[60vh] max-w-md flex-col items-center justify-center bg-[var(--gvp-canvas)] px-4 py-8 text-[var(--gvp-text)]">
    <div class="w-full rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-8">
      <VravraLogo size="md" animate show-tagline />
      <h1 class="mt-4 text-xl font-bold text-[var(--gvp-text)]">로그인</h1>
      <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
        OIDC 싱글사인온으로 안전하게 로그인하세요.
      </p>

      <div v-if="authErrorMessage" class="mt-4 rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] px-3 py-2 text-xs font-medium text-[var(--gvp-live)]" role="alert">
        {{ authErrorMessage }}
      </div>

      <a
        href="/auth/login"
        class="mt-6 flex w-full items-center justify-center gap-2 rounded-md bg-[var(--gvp-brand)] px-6 py-3 text-sm font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)]"
      >
        {{ BRAND.name }} 계정으로 계속하기
      </a>

      <ClientOnly>
      <div v-if="isLocalhost" class="mt-6 border-t border-[var(--gvp-border)] pt-5">
        <p class="text-xs font-bold text-[var(--gvp-warning)]">로컬 개발 전용 토큰 로그인</p>
        <p class="mt-1 text-[11px] text-[var(--gvp-text-secondary)]">
          예: admin-token · creator-token · viewer-token · advertiser-token
        </p>
        <div class="mt-3 flex gap-2">
          <input
            v-model="devToken"
            type="text"
            placeholder="dev token"
            class="h-10 flex-1 rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-3 font-mono text-xs text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-brand)] focus:bg-[var(--gvp-surface)]"
            @keyup.enter="loginWithDevToken"
          />
          <button
            type="button"
            :disabled="devLoading"
            class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-elevated)] px-4 text-xs font-semibold text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)] disabled:opacity-50"
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
