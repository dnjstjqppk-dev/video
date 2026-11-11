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
  <section class="mx-auto max-w-6xl bg-white px-4 py-8 text-slate-900 sm:px-8">
    <div class="flex flex-wrap items-start justify-between gap-5 border-b border-slate-200 pb-6">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 sm:text-3xl">{{ titles[section] }}</h1>
        <p class="mt-2 text-sm text-slate-500 tabular-nums">
          {{ videos.length }}개의 동영상<span v-if="section !== 'subscriptions'"> · 이 브라우저에 저장됨</span>
        </p>
      </div>
      <input
        v-model="query"
        type="search"
        :aria-label="`${titles[section]} 검색`"
        placeholder="동영상 검색"
        class="rounded-md border border-slate-200 bg-slate-50 px-4 py-2 text-sm text-slate-900 outline-none focus:border-blue-600 focus:bg-white"
      />
    </div>

    <p v-if="library.storageError" role="alert" class="mt-4 rounded-md border border-amber-200 bg-amber-50 px-4 py-2.5 text-xs font-medium text-amber-700">
      브라우저 저장소에 접근할 수 없습니다. 변경 사항은 이 화면에서만 유지됩니다.
    </p>

    <div v-if="section === 'history'" class="my-6 flex flex-wrap gap-3">
      <button
        type="button"
        class="rounded-md bg-slate-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-slate-800"
        @click="library.togglePaused()"
      >
        {{ library.paused ? '시청 기록 다시 저장' : '시청 기록 일시중지' }}
      </button>
      <button
        v-if="library.history.length"
        type="button"
        class="rounded-md border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700 transition hover:bg-slate-50"
        @click="confirmClear = true"
      >
        시청 기록 모두 지우기
      </button>
      <div
        v-if="confirmClear"
        role="alert"
        class="flex w-full flex-wrap items-center justify-between gap-4 rounded-lg border border-red-200 bg-red-50 p-4 text-xs text-red-700"
      >
        <span class="font-medium">이 브라우저의 모든 시청 기록을 삭제할까요?</span>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="rounded-md border border-red-200 bg-white px-3 py-1.5 font-semibold text-red-700 hover:bg-red-100"
            @click="library.clearHistory(); confirmClear = false"
          >
            삭제
          </button>
          <button
            type="button"
            class="rounded-md border border-slate-200 bg-white px-3 py-1.5 font-semibold text-slate-700 hover:bg-slate-50"
            @click="confirmClear = false"
          >
            취소
          </button>
        </div>
      </div>
    </div>

    <p v-if="loading" role="status" class="py-16 text-center text-sm text-slate-500">
      구독 동영상을 불러오는 중…
    </p>

    <div v-else-if="failed && section === 'subscriptions'" class="rounded-lg border border-red-200 bg-red-50 py-16 text-center">
      <p class="text-sm font-semibold text-red-700">구독 동영상을 불러오지 못했습니다.</p>
      <button
        type="button"
        class="mt-4 rounded-md bg-slate-900 px-4 py-2 text-xs font-semibold text-white hover:bg-slate-800"
        @click="fetchSubscriptions"
      >
        다시 시도
      </button>
    </div>

    <div v-else-if="!videos.length" class="mt-6 rounded-lg border border-slate-200 bg-slate-50 py-20 text-center">
      <h2 class="text-lg font-bold text-slate-900">{{ query ? '검색 결과가 없습니다' : '아직 동영상이 없습니다' }}</h2>
      <p class="mt-2 text-sm text-slate-500">
        {{ section === 'saved' ? '동영상 메뉴에서 나중에 보기를 선택해 보세요.' : section === 'subscriptions' ? '좋아하는 채널을 구독하면 새로운 동영상을 모아 볼 수 있습니다.' : section === 'liked' ? '마음에 드는 동영상에 좋아요를 눌러 보세요.' : '시청한 동영상이 여기에 표시됩니다.' }}
      </p>
      <NuxtLink to="/" class="mt-6 inline-block rounded-md bg-slate-900 px-5 py-2.5 text-xs font-semibold text-white transition hover:bg-slate-800">
        동영상 둘러보기
      </NuxtLink>
    </div>

    <div v-else class="mt-6 space-y-4">
      <article
        v-for="video in videos"
        :key="video.id"
        class="flex items-start gap-4 rounded-lg border border-slate-200 bg-white p-3 transition hover:bg-slate-50"
      >
        <NuxtLink :to="`/watch/${video.id}`" class="relative aspect-video w-40 shrink-0 overflow-hidden rounded-md border border-slate-200 bg-slate-100 sm:w-72">
          <img :src="video.thumbnailUrl" :alt="video.title" loading="lazy" class="h-full w-full object-cover" />
          <span class="absolute bottom-2 right-2 rounded-md border border-slate-200 bg-white/95 px-1.5 py-0.5 font-mono text-xs font-semibold text-slate-900 tabular-nums">
            {{ Math.floor(video.durationSeconds / 60) }}:{{ String(video.durationSeconds % 60).padStart(2, '0') }}
          </span>
        </NuxtLink>
        <div class="min-w-0 flex-1">
          <NuxtLink :to="`/watch/${video.id}`" class="line-clamp-2 font-bold text-slate-900 hover:text-blue-600 sm:text-base">
            {{ video.title }}
          </NuxtLink>
          <NuxtLink :to="`/channels/${video.channelId}`" class="mt-1.5 block text-xs font-medium text-slate-600 hover:text-slate-900">
            {{ video.channel?.name }}
          </NuxtLink>
          <p class="mt-1 text-xs text-slate-500 tabular-nums">
            조회수 {{ video.viewCount.toLocaleString() }}회
          </p>
        </div>
        <button
          v-if="section !== 'subscriptions'"
          type="button"
          :aria-label="`${video.title} 목록에서 삭제`"
          class="rounded-md border border-slate-200 bg-white px-2.5 py-1 text-xs font-semibold text-slate-500 transition hover:bg-slate-100 hover:text-slate-900"
          @click="remove(video.id)"
        >
          ×
        </button>
      </article>
    </div>
  </section>
</template>
