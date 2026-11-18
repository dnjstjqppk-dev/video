<script setup lang="ts">
import type { NuxtError } from '#app'

const props = defineProps<{
  error: NuxtError
}>()

const is404 = computed(() => props.error.statusCode === 404)

const handleError = () => clearError({ redirect: '/' })
</script>

<template>
  <div class="flex min-h-screen flex-col items-center justify-center bg-[var(--gvp-canvas)] px-4 text-center text-[var(--gvp-text)]">
    <div class="w-full max-w-md rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-6">
      <span class="inline-flex items-center rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] px-2.5 py-0.5 font-mono text-xs font-medium text-[var(--gvp-live)] tabular-nums">
        HTTP {{ error.statusCode || 404 }}
      </span>
      <h1 class="mt-4 text-xl font-bold tracking-tight text-[var(--gvp-text)]">
        {{ is404 ? '어드민 페이지를 찾을 수 없습니다' : '요청을 처리할 수 없습니다' }}
      </h1>
      <p class="mt-2 text-sm text-[var(--gvp-text-secondary)] leading-relaxed">
        {{ error.statusMessage || (is404 ? '요청하신 주소는 존재하지 않거나 이동되었습니다.' : '일시적인 서버 오류가 발생했습니다.') }}
      </p>
      <div class="mt-6">
        <button
          type="button"
          class="inline-flex h-9 items-center justify-center rounded-md bg-[var(--gvp-brand)] px-4 text-sm font-medium text-white transition-colors hover:bg-[var(--gvp-brand-hover)]"
          @click="handleError"
        >
          관제 홈으로 이동
        </button>
      </div>
    </div>
  </div>
</template>
