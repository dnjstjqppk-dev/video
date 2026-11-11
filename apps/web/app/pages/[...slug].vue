<script setup lang="ts">
const event = useRequestEvent()
if (event) {
  setResponseStatus(event, 404)
}

useSeoMeta({
  title: '404 — 페이지를 찾을 수 없습니다'
})
</script>

<template>
  <div class="flex min-h-[60vh] flex-col items-center justify-center px-4 text-center">
    <div class="w-full max-w-md rounded-lg border border-slate-200 bg-white p-6">
      <span class="inline-flex items-center rounded-md border border-red-200 bg-red-50 px-2.5 py-0.5 font-mono text-xs font-medium text-red-700 tabular-nums">
        HTTP 404 NOT FOUND
      </span>
      <h1 class="mt-4 text-xl font-bold tracking-tight text-slate-900">
        페이지를 찾을 수 없습니다
      </h1>
      <p class="mt-2 text-sm text-slate-600 leading-relaxed">
        요청하신 주소는 존재하지 않거나 이동되었습니다.
      </p>
      <div class="mt-6">
        <NuxtLink
          to="/"
          class="inline-flex h-9 items-center justify-center rounded-md bg-slate-900 px-4 text-sm font-medium text-white transition-colors hover:bg-slate-800"
        >
          홈으로 돌아가기
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
