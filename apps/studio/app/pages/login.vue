<script setup lang="ts">
import type { UserProfile } from '@gvp/shared-types'
import VravraLogo from '@gvp/ui/src/components/VravraLogo.vue'

const api = useStudioApi()
const sessionUser = useState<UserProfile | null>('studio-session', () => null)

useSeoMeta({
  title: '로그인 — vravra studio'
})

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
    sessionUser.value = await api.getMe()
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
      <VravraLogo :size="28" suffix="studio" suffix-tone="studio" />
      <h1 class="mt-5 text-xl font-semibold tracking-tight text-[var(--gvp-text)]">크리에이터 로그인</h1>
      <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
        OIDC 싱글사인온으로 안전하게 로그인하세요.
      </p>

      <a
        href="/auth/login"
        class="mt-6 w-full inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors"
      >
        vravra 계정으로 계속하기
      </a>

      <div v-if="isLocalhost" class="mt-6 border-t border-[var(--gvp-border)] pt-5">
        <p class="text-xs font-semibold text-[var(--gvp-text)]">로컬 개발 전용 토큰 로그인</p>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          예: creator-token · admin-token
        </p>
        <div class="mt-3 flex gap-2">
          <input
            v-model="devToken"
            type="text"
            placeholder="dev token"
            class="h-9 flex-1 rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3 font-mono text-xs text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-focus)] focus:bg-[var(--gvp-surface)]"
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
    </div>
  </div>
</template>
