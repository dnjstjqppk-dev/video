<script setup lang="ts">
const api = useStudioApi()
const currency = useState<'KRW' | 'USD'>('studio-currency', () => 'KRW')

useSeoMeta({
  title: 'Creator Dashboard'
})

const { data: dashboard, pending: dashPending, error: dashError } = await useAsyncData('studio-dashboard', async () => {
  return await api.getStudioDashboard()
})

const { data: claimsData } = await useAsyncData('studio-claims', async () => {
  try {
    return await api.getStudioClaims()
  } catch {
    return []
  }
})
const recentVideos = computed(() => dashboard.value?.recentVideos || [])
const recentComments = computed(() => dashboard.value?.recentComments || [])
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">채널 대시보드</h1>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          {{ dashboard?.channel?.name || '내 채널' }} · 최근 28일 실적 요약
        </p>
      </div>
      <div class="flex items-center gap-2">
        <NuxtLink to="/upload" class="inline-flex h-9 items-center gap-2 rounded-full bg-[var(--gvp-studio)] px-4 text-xs font-semibold text-white transition-colors hover:opacity-90">
          <span class="text-base leading-none" aria-hidden="true">+</span>
          <span>동영상 업로드</span>
        </NuxtLink>
        <NuxtLink to="/live" class="inline-flex h-9 items-center rounded-full border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-4 text-xs font-medium text-[var(--gvp-text)] transition-colors hover:bg-[var(--gvp-canvas)]">
          <span>라이브 시작</span>
        </NuxtLink>
      </div>
    </div>

    <div v-if="dashPending" class="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <div v-for="i in 3" :key="i" class="h-64 animate-pulse rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-elevated)]" />
    </div>
    <div v-else-if="dashError || !dashboard" class="rounded-xl border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-[var(--gvp-live)]">대시보드를 불러오지 못했습니다.</p>
    </div>
    <div v-else class="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <LatestVideoCard :videos="recentVideos" />
      <ChannelAnalyticsCard :analytics="dashboard.analytics" :videos="recentVideos" :currency="currency" />
      <LatestCommentsCard :comments="recentComments" />
    </div>

    <div v-if="!dashPending && !dashError && dashboard" class="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <div v-if="recentVideos.length === 0" class="rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-8 text-center lg:col-span-2">
        <p class="text-sm font-semibold text-[var(--gvp-text-secondary)]">아직 업로드한 동영상이 없습니다.</p>
        <NuxtLink to="/upload" class="mt-2 inline-block text-xs font-medium underline">첫 동영상 업로드하기</NuxtLink>
      </div>
      <div v-else class="overflow-x-auto rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] lg:col-span-2">
        <div class="flex items-center justify-between border-b border-[var(--gvp-border)] px-5 py-4">
          <h2 class="text-sm font-semibold text-[var(--gvp-text)]">최근 업로드</h2>
          <NuxtLink to="/content" class="text-xs font-medium text-[var(--gvp-brand-hover)] hover:underline">콘텐츠 전체 관리</NuxtLink>
        </div>
        <table class="w-full text-left">
          <thead class="border-b border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-secondary)]">
            <tr>
              <th scope="col" class="px-4 py-2.5">동영상</th>
              <th scope="col" class="px-4 py-2.5">반응</th>
              <th scope="col" class="px-4 py-2.5">상태</th>
              <th scope="col" class="px-4 py-2.5 text-right">관리</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--gvp-border)]">
            <tr v-for="vid in recentVideos" :key="vid.id" class="transition-colors hover:bg-[var(--gvp-canvas)]">
              <td class="px-4 py-3">
                <div class="flex items-center gap-3">
                  <img :src="vid.thumbnailUrl" :alt="vid.title" class="h-10 w-[72px] shrink-0 rounded-md object-cover" loading="lazy" />
                  <p class="line-clamp-1 text-sm font-medium text-[var(--gvp-text)]">{{ vid.title }}</p>
                </div>
              </td>
              <td class="whitespace-nowrap px-4 py-3 text-xs text-[var(--gvp-text-secondary)] tabular-nums">
                {{ vid.viewCount.toLocaleString() }}회 · {{ vid.likeCount.toLocaleString() }} · {{ vid.commentCount.toLocaleString() }}
              </td>
              <td class="px-4 py-3">
                <span class="inline-flex items-center rounded-md border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] px-2 py-0.5 text-xs font-medium text-[var(--gvp-success)]">{{ vid.status }}</span>
              </td>
              <td class="px-4 py-3 text-right">
                <NuxtLink :to="`/editor/${vid.id}`" class="inline-flex h-8 items-center rounded-md border border-[var(--gvp-border-strong)] px-3 text-xs font-medium hover:bg-[var(--gvp-canvas)]">편집</NuxtLink>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <ChannelAlertsCard :claims="claimsData || []" />
    </div>
  </div>
</template>
