<script setup lang="ts">
import VravraLogo from '@gvp/ui/src/components/VravraLogo.vue'
import { useAppStore } from '~/stores/app'

const appStore = useAppStore()
const api = useApi()

useSeoMeta({
  title: '로그인 — vravra 계정'
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
  <div class="mx-auto flex min-h-[60vh] max-w-md flex-col items-center justify-center bg-[var(--gvp-surface)] px-4 py-8 text-[var(--gvp-text)]">
    <div class="w-full rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-8">
      <VravraLogo :size="32" />
      <h1 class="mt-4 text-xl font-bold text-[var(--gvp-text)]">로그인</h1>
      <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
        OIDC 싱글사인온으로 안전하게 로그인하세요.
      </p>

      <a
        href="/auth/login"
        class="mt-6 flex w-full items-center justify-center gap-2 rounded-md bg-[var(--gvp-brand)] px-6 py-3 text-sm font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)]"
      >
        vravra 계정으로 계속하기
      </a>

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
    </div>
  </div>
</template>
