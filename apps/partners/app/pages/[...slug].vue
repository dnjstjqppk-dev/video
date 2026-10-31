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
    <div class="max-w-md rounded-2xl border border-white/10 bg-[#141414] p-8 shadow-2xl">
      <span class="rounded-full bg-red-500/15 px-3 py-1 font-mono text-xs font-bold text-red-400">
        HTTP 404 NOT FOUND
      </span>
      <h1 class="mt-4 text-2xl font-black tracking-tight text-white">
        페이지를 찾을 수 없습니다
      </h1>
      <p class="mt-2 text-xs text-neutral-400">
        요청하신 주소는 존재하지 않거나 이전되었습니다.
      </p>
      <NuxtLink
        to="/"
        class="mt-6 inline-block rounded-full bg-[#FF0033] px-6 py-2.5 text-xs font-bold text-white transition hover:bg-[#D9002B]"
      >
        홈으로 이동
      </NuxtLink>
    </div>
  </div>
</template>
