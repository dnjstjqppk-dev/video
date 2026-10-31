<script setup lang="ts">
import type { NuxtError } from '#app'

const props = defineProps<{
  error: NuxtError
}>()

const is404 = computed(() => props.error.statusCode === 404)

const handleError = () => clearError({ redirect: '/' })
</script>

<template>
  <div class="flex min-h-screen flex-col items-center justify-center bg-[#0F0F0F] px-4 text-center text-white">
    <div class="max-w-md rounded-2xl border border-white/10 bg-[#141414] p-8 shadow-2xl">
      <span class="rounded-full bg-red-500/15 px-3 py-1 font-mono text-xs font-bold text-red-400">
        HTTP {{ error.statusCode || 404 }}
      </span>
      <h1 class="mt-4 text-2xl font-black tracking-tight">
        {{ is404 ? '페이지를 찾을 수 없습니다' : '오류가 발생했습니다' }}
      </h1>
      <p class="mt-2 text-xs text-neutral-400">
        {{ error.statusMessage || (is404 ? '요청하신 주소는 존재하지 않거나 이전되었습니다.' : '일시적인 서버 문제가 발생했습니다.') }}
      </p>
      <button
        type="button"
        class="mt-6 rounded-full bg-[#FF0033] px-6 py-2.5 text-xs font-bold text-white transition hover:bg-[#D9002B]"
        @click="handleError"
      >
        홈으로 이동
      </button>
    </div>
  </div>
</template>
