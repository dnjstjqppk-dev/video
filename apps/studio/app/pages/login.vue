<script setup lang="ts">
import type { UserProfile } from '@gvp/shared-types'

const api = useStudioApi()
const sessionUser = useState<UserProfile | null>('studio-session', () => null)

useSeoMeta({
  title: '로그인 — Creator Studio'
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
    <div class="w-full rounded-lg border border-slate-200 bg-white p-8">
      <div class="flex items-center gap-2.5">
        <span class="flex h-8 w-8 items-center justify-center rounded-md bg-slate-900 text-sm font-semibold text-white">
          S
        </span>
        <span class="text-base font-semibold tracking-tight text-slate-900">Creator Studio</span>
      </div>
      <h1 class="mt-5 text-xl font-semibold tracking-tight text-slate-900">크리에이터 로그인</h1>
      <p class="mt-1 text-xs text-slate-500">
        OIDC 싱글사인온으로 안전하게 로그인하세요.
      </p>

      <a
        href="/auth/login"
        class="mt-6 w-full inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors"
      >
        StreamVerse 계정으로 계속하기
      </a>

      <div v-if="isLocalhost" class="mt-6 border-t border-slate-200 pt-5">
        <p class="text-xs font-semibold text-slate-900">로컬 개발 전용 토큰 로그인</p>
        <p class="mt-1 text-xs text-slate-500">
          예: creator-token · admin-token
        </p>
        <div class="mt-3 flex gap-2">
          <input
            v-model="devToken"
            type="text"
            placeholder="dev token"
            class="h-9 flex-1 rounded-md border border-slate-300 bg-slate-50 px-3 font-mono text-xs text-slate-900 outline-none focus:border-slate-900 focus:bg-white"
            @keyup.enter="loginWithDevToken"
          />
          <button
            type="button"
            :disabled="devLoading"
            class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors disabled:opacity-50"
            @click="loginWithDevToken"
          >
            {{ devLoading ? '...' : '로그인' }}
          </button>
        </div>
        <p v-if="devError" class="mt-2 text-xs font-medium text-red-600">{{ devError }}</p>
      </div>
    </div>
  </div>
</template>
