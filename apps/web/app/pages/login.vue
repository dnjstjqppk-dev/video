<script setup lang="ts">
import { useAppStore } from '~/stores/app'

const appStore = useAppStore()
const api = useApi()

useSeoMeta({
  title: '로그인 — StreamVerse 계정'
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
  <div class="mx-auto flex min-h-[60vh] max-w-md flex-col items-center justify-center px-4">
    <div class="w-full rounded-2xl border border-white/10 bg-[#141414] p-8 shadow-2xl">
      <div class="flex items-center gap-2">
        <span class="flex h-8 w-10 items-center justify-center rounded-lg bg-[#FF0033] text-white">
          <svg class="h-4 w-4 fill-current" viewBox="0 0 24 24">
            <path d="M8 5v14l11-7z" />
          </svg>
        </span>
        <span class="text-lg font-extrabold text-white">StreamVerse</span>
      </div>
      <h1 class="mt-4 text-xl font-black text-white">로그인</h1>
      <p class="mt-1 text-xs text-neutral-400">
        OIDC 싱글사인온으로 안전하게 로그인하세요.
      </p>

      <a
        href="/auth/login"
        class="mt-6 flex w-full items-center justify-center gap-2 rounded-full bg-[#FF0033] px-6 py-3 text-sm font-bold text-white transition hover:bg-[#D9002B]"
      >
        StreamVerse 계정으로 계속하기
      </a>

      <div v-if="isLocalhost" class="mt-6 border-t border-white/10 pt-5">
        <p class="text-xs font-bold text-amber-300">로컬 개발 전용 토큰 로그인</p>
        <p class="mt-1 text-[11px] text-neutral-400">
          예: admin-token · creator-token · viewer-token · advertiser-token
        </p>
        <div class="mt-3 flex gap-2">
          <input
            v-model="devToken"
            type="text"
            placeholder="dev token"
            class="h-10 flex-1 rounded-xl border border-white/15 bg-[#0D0F14] px-3 font-mono text-xs text-white outline-none focus:border-[#3EA6FF]"
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
