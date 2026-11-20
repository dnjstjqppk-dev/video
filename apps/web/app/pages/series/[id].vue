<script setup lang="ts">
import { formatTimecode, formatViewCount } from '@gvp/ui'

const route = useRoute()
const seriesId = computed(() => String(route.params.id))
const appStore = useAppStore()
const api = useApi()
const ko = computed(() => appStore.locale === 'ko')

const { data: detail, pending, error } = await useAsyncData(`series-page-${seriesId.value}`, async () => {
  return await api.getSeries(seriesId.value)
})

useSeoMeta({
  title: () => detail.value ? `${detail.value.series.title} — ${ko.value ? '시리즈 정주행' : 'Binge the series'}` : 'vravra',
  description: () => (detail.value?.series.description || '').slice(0, 160),
  ogImage: () => detail.value?.series.coverUrl,
})

const episodeList = computed(() => detail.value?.episodes || [])
const seasons = computed(() => {
  const groups = new Map<number, typeof episodeList.value>()
  for (const ep of episodeList.value) {
    if (!groups.has(ep.seasonNo)) groups.set(ep.seasonNo, [])
    groups.get(ep.seasonNo)!.push(ep)
  }
  return [...groups.entries()].sort((a, b) => a[0] - b[0])
})
const firstPlayable = computed(() => episodeList.value.find(ep => ep.status === 'READY' && ep.visibility === 'public'))
const backdrop = computed(() =>
  detail.value?.series.coverUrl
  || episodeList.value[0]?.thumbnailUrl
  || 'https://images.unsplash.com/photo-1531297484001-80022131f5a1?w=1600'
)
</script>

<template>
  <!-- Cinema mode: matches the /ai hall, dark in both themes. -->
  <div class="min-h-screen bg-neutral-950 text-white">
    <div v-if="pending" class="mx-auto max-w-5xl animate-pulse px-4 py-6 sm:px-6" aria-label="로딩 중">
      <div class="aspect-video w-full rounded-xl bg-white/10" />
      <div class="mt-4 h-6 w-1/2 rounded bg-white/10" />
    </div>
    <div v-else-if="error || !detail" class="mx-auto max-w-5xl px-4 py-6 sm:px-6">
      <div class="rounded-xl border border-white/10 bg-white/5 p-10 text-center">
        <p class="text-sm text-white/60">{{ ko ? '시리즈를 찾을 수 없습니다.' : 'Series not found.' }}</p>
        <NuxtLink to="/ai" class="mt-4 inline-block rounded-full bg-[var(--gvp-brand)] px-5 py-2 text-xs font-semibold text-white">AI Originals</NuxtLink>
      </div>
    </div>
    <template v-else>
      <section class="relative h-[380px] w-full overflow-hidden sm:h-[480px]" aria-label="시리즈 소개">
        <img :src="backdrop" :alt="detail.series.title" class="absolute inset-0 h-full w-full object-cover" />
        <div class="absolute inset-0 bg-gradient-to-t from-neutral-950 via-neutral-950/40 to-transparent" />
        <div class="absolute inset-0 bg-gradient-to-r from-neutral-950/80 via-transparent to-transparent" />
        <div class="absolute inset-x-0 bottom-0 mx-auto max-w-5xl px-4 pb-8 sm:px-6">
          <p class="flex items-center gap-2 text-xs font-bold uppercase tracking-[0.2em] text-white/80">
            <AiBadge v-if="detail.series.aiOriginal" approved /> {{ ko ? '시리즈' : 'Series' }} · {{ detail.episodeCount }}{{ ko ? '화' : ' episodes' }}
          </p>
          <h1 class="mt-3 max-w-3xl text-3xl font-black drop-shadow-lg sm:text-5xl">{{ detail.series.title }}</h1>
          <p v-if="detail.series.description" class="mt-3 max-w-xl text-sm text-white/80">{{ detail.series.description }}</p>
          <div class="mt-5 flex flex-wrap items-center gap-3">
            <NuxtLink
              v-if="firstPlayable"
              :to="`/watch/${firstPlayable.videoId}`"
              class="inline-flex items-center gap-2 rounded-md bg-white px-6 py-2.5 text-sm font-bold text-black transition hover:bg-white/80"
            >
              <span aria-hidden="true">▶</span> {{ ko ? '1화부터 정주행' : 'Binge from E1' }}
            </NuxtLink>
            <span class="rounded-full bg-white/20 px-3 py-1 text-xs font-semibold text-white backdrop-blur">
              {{ detail.series.status === 'completed' ? (ko ? '완결' : 'Completed') : (ko ? '연재 중' : 'Ongoing') }}
            </span>
          </div>
        </div>
      </section>

      <div class="mx-auto max-w-5xl px-4 pb-10 sm:px-6">
        <section v-for="[season, eps] in seasons" :key="season" class="mt-8" :aria-label="`시즌 ${season}`">
          <h2 class="mb-3 text-base font-bold">{{ ko ? `시즌 ${season}` : `Season ${season}` }}</h2>
          <ol class="divide-y divide-white/10 overflow-hidden rounded-xl border border-white/10 bg-white/5">
            <li v-for="ep in eps" :key="ep.videoId">
              <NuxtLink :to="`/watch/${ep.videoId}`" class="group flex items-center gap-4 p-3 transition hover:bg-white/10 sm:p-4">
                <span class="w-8 shrink-0 text-center text-sm font-bold tabular-nums text-white/50">{{ ep.episodeNo }}</span>
                <span class="relative h-16 w-28 shrink-0 overflow-hidden rounded-lg bg-white/10 transition duration-300 group-hover:scale-[1.03] sm:h-[72px] sm:w-32">
                  <img :src="ep.thumbnailUrl" :alt="ep.title" class="h-full w-full object-cover" loading="lazy" />
                  <span class="absolute inset-0 flex items-center justify-center bg-black/30 text-lg text-white opacity-0 transition group-hover:opacity-100" aria-hidden="true">▶</span>
                  <span class="absolute bottom-1 right-1 rounded bg-black/80 px-1 py-0.5 text-[10px] font-medium text-white tabular-nums">{{ formatTimecode(ep.durationSeconds) }}</span>
                </span>
                <span class="min-w-0 flex-1">
                  <span class="line-clamp-2 block text-sm font-semibold">{{ ep.title }}</span>
                  <span class="mt-1 block text-xs tabular-nums text-white/50">{{ formatViewCount(ep.viewCount, appStore.locale) }}</span>
                </span>
                <AiBadge v-if="ep.aiGenerated" :tool="ep.aiTool" class="shrink-0" />
              </NuxtLink>
            </li>
          </ol>
        </section>
      </div>
    </template>
  </div>
</template>
