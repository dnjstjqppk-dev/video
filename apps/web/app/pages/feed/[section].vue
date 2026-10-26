<script setup lang="ts">
import { useLibraryStore } from '~/stores/library'
import type { Video } from '@gvp/shared-types'
const library = useLibraryStore()
const route = useRoute()
const api = useApi()
const section = computed(() => String(route.params.section))
const titles: Record<string, string> = { history: '시청 기록', saved: '나중에 볼 동영상', liked: '좋아요 표시한 동영상', subscriptions: '구독' }
if (!titles[section.value]) throw createError({ statusCode: 404, statusMessage: 'Page not found' })
const query = ref('')
const confirmClear = ref(false)
const feed = ref<Video[]>([])
const loading = ref(false)
const failed = ref(false)
async function fetchSubscriptions() {
  loading.value = true
  failed.value = false
  try { feed.value = (await api.getHomeFeed(1, 100, 'all')).items || [] }
  catch { failed.value = true }
  finally { loading.value = false }
}
onMounted(() => { library.hydrate(); if (section.value === 'subscriptions') fetchSubscriptions() })
watch(section, value => { query.value = ''; confirmClear.value = false; if (value === 'subscriptions') fetchSubscriptions() })
const videos = computed(() => {
  const list = section.value === 'subscriptions' ? feed.value.filter(v => library.subscriptions.includes(v.channelId))
    : section.value === 'history' ? library.history : section.value === 'saved' ? library.saved : library.liked
  return list.filter(v => `${v.title} ${v.channel?.name || ''}`.toLowerCase().includes(query.value.toLowerCase()))
})
function remove(id: string) {
  if (section.value === 'history' || section.value === 'saved' || section.value === 'liked') library.remove(section.value, id)
}
useSeoMeta({ title: () => titles[section.value] || '보관함' })
</script>

<template>
  <section class="mx-auto max-w-6xl px-4 py-8 sm:px-8">
    <div class="flex flex-wrap items-start justify-between gap-5">
      <div><h1 class="text-3xl font-bold">{{ titles[section] }}</h1><p class="mt-3 text-sm text-neutral-400">{{ videos.length }}개의 동영상<span v-if="section !== 'subscriptions'"> · 이 브라우저에 저장됨</span></p></div>
      <input v-model="query" type="search" :aria-label="`${titles[section]} 검색`" placeholder="동영상 검색" class="rounded-full border border-neutral-500 bg-transparent px-5 py-2.5" />
    </div>
    <p v-if="library.storageError" role="alert" class="mt-4 text-amber-400">브라우저 저장소에 접근할 수 없습니다. 변경 사항은 이 화면에서만 유지됩니다.</p>
    <div v-if="section === 'history'" class="my-6 flex flex-wrap gap-3">
      <button class="rounded-full bg-neutral-800 px-4 py-2 text-sm text-white" @click="library.togglePaused()">{{ library.paused ? '시청 기록 다시 저장' : '시청 기록 일시중지' }}</button>
      <button v-if="library.history.length" class="rounded-full border border-neutral-500 px-4 py-2 text-sm" @click="confirmClear = true">시청 기록 모두 지우기</button>
      <div v-if="confirmClear" role="alert" class="flex w-full flex-wrap items-center gap-4 rounded-xl border border-neutral-500 p-4">
        <span>이 브라우저의 모든 시청 기록을 삭제할까요?</span>
        <button class="text-red-400" @click="library.clearHistory(); confirmClear = false">삭제</button><button @click="confirmClear = false">취소</button>
      </div>
    </div>
    <p v-if="loading" role="status" class="py-16 text-center">구독 동영상을 불러오는 중…</p>
    <div v-else-if="failed && section === 'subscriptions'" class="py-16 text-center"><p>구독 동영상을 불러오지 못했습니다.</p><button class="mt-4 text-blue-400" @click="fetchSubscriptions">다시 시도</button></div>
    <div v-else-if="!videos.length" class="py-24 text-center">
      <h2 class="text-xl font-semibold">{{ query ? '검색 결과가 없습니다' : '아직 동영상이 없습니다' }}</h2>
      <p class="mt-3 text-sm text-neutral-400">{{ section === 'saved' ? '동영상 메뉴에서 나중에 보기를 선택해 보세요.' : section === 'subscriptions' ? '좋아하는 채널을 구독하면 새로운 동영상을 모아 볼 수 있습니다.' : section === 'liked' ? '마음에 드는 동영상에 좋아요를 눌러 보세요.' : '시청한 동영상이 여기에 표시됩니다.' }}</p>
      <NuxtLink to="/" class="mt-6 inline-block rounded-full bg-blue-500 px-5 py-2.5 text-white">동영상 둘러보기</NuxtLink>
    </div>
    <div v-else class="mt-6 space-y-5">
      <article v-for="video in videos" :key="video.id" class="flex items-start gap-4">
        <NuxtLink :to="`/watch/${video.id}`" class="relative aspect-video w-40 shrink-0 overflow-hidden rounded-xl bg-neutral-800 sm:w-72">
          <img :src="video.thumbnailUrl" :alt="video.title" loading="lazy" class="h-full w-full object-cover" />
          <span class="absolute bottom-2 right-2 rounded bg-black/80 px-1 text-xs text-white">{{ Math.floor(video.durationSeconds / 60) }}:{{ String(video.durationSeconds % 60).padStart(2, '0') }}</span>
        </NuxtLink>
        <div class="min-w-0 flex-1"><NuxtLink :to="`/watch/${video.id}`" class="line-clamp-2 font-semibold sm:text-lg">{{ video.title }}</NuxtLink><NuxtLink :to="`/channels/${video.channelId}`" class="mt-2 block text-sm text-neutral-400">{{ video.channel?.name }}</NuxtLink><p class="mt-1 text-xs text-neutral-400">조회수 {{ video.viewCount.toLocaleString() }}회</p></div>
        <button v-if="section !== 'subscriptions'" :aria-label="`${video.title} 목록에서 삭제`" class="rounded-full px-3 py-1 hover:bg-neutral-700" @click="remove(video.id)">×</button>
      </article>
    </div>
  </section>
</template>
