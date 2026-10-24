<script setup lang="ts">
import { VideoProcessingStatus } from '@gvp/shared-types'

const api = useStudioApi()
const currency = useState<'KRW' | 'USD'>('studio-currency', () => 'KRW')

useSeoMeta({
  title: 'Creator Dashboard'
})

const { data: dashboard } = await useAsyncData('studio-dashboard', async () => {
  try {
    return await api.getStudioDashboard()
  } catch {
    return {
      channel: {
        id: 'ch_tech_kr',
        ownerId: 'usr_1',
        handle: '@gvptech',
        name: 'GVP Tech Engineering',
        subscriberCount: 284000,
        videoCount: 142,
        isVerified: true,
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2026-10-24T00:00:00Z'
      },
      analytics: {
        views: 1942800,
        watchTimeHours: 148520,
        subscribers: 284000,
        estimatedRevenue: 18450000
      },
      recentVideos: [
        {
          id: 'vid_arch_2026',
          channelId: 'ch_tech_kr',
          title: 'Building a 10M Concurrent Viewer Global Video Platform with Go 1.24, Nuxt 4 & ClickHouse',
          viewCount: 412900,
          likeCount: 32400,
          commentCount: 1842,
          status: VideoProcessingStatus.READY,
          publishedAt: '2026-10-20T09:00:00Z'
        },
        {
          id: 'vid_ai_agents',
          channelId: 'ch_tech_kr',
          title: 'Multimodal Video Indexing with Audio-Visual Fingerprints & Vector Embeddings',
          viewCount: 189400,
          likeCount: 16200,
          commentCount: 740,
          status: VideoProcessingStatus.READY,
          publishedAt: '2026-10-23T11:00:00Z'
        }
      ],
      recentComments: []
    }
  }
})

function formatRevenue(krw: number): string {
  if (currency.value === 'KRW') {
    return `₩${krw.toLocaleString('ko-KR')}`
  }
  return `$${(krw / 1350).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}
</script>

<template>
  <div>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">채널 대시보드 (Creator Dashboard)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          Real-time ClickHouse telemetry · Last 28 days performance & double-entry settlement accruals
        </p>
      </div>
      <div class="flex gap-2">
        <NuxtLink
          to="/upload"
          class="rounded-xl bg-[#FF0033] px-4 py-2 text-xs font-bold text-white hover:bg-[#D9002B]"
        >
          ⬆️ Upload Video
        </NuxtLink>
        <NuxtLink
          to="/live"
          class="rounded-xl border border-white/15 bg-[#181C28] px-4 py-2 text-xs font-bold text-white hover:bg-white/10"
        >
          📡 Go Live
        </NuxtLink>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <p class="text-xs font-semibold text-neutral-400">총 조회수 (28d Views)</p>
        <p class="mt-2 text-2xl font-extrabold text-white">
          {{ dashboard?.analytics.views.toLocaleString() }}
        </p>
        <p class="mt-1 text-xs font-semibold text-emerald-400">▲ +18.4% vs previous 28d</p>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <p class="text-xs font-semibold text-neutral-400">시청 시간 (Watch Time Hours)</p>
        <p class="mt-2 text-2xl font-extrabold text-white">
          {{ dashboard?.analytics.watchTimeHours.toLocaleString() }} hrs
        </p>
        <p class="mt-1 text-xs font-semibold text-emerald-400">▲ +12.1% avg retention 64.2%</p>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <p class="text-xs font-semibold text-neutral-400">구독자 수 (Subscribers)</p>
        <p class="mt-2 text-2xl font-extrabold text-white">
          {{ dashboard?.analytics.subscribers.toLocaleString() }}
        </p>
        <p class="mt-1 text-xs font-semibold text-[#3EA6FF]">+14,200 net new this month</p>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <p class="text-xs font-semibold text-neutral-400">예상 정산 수익 (Est. Net Revenue)</p>
        <p class="mt-2 text-2xl font-extrabold text-emerald-400">
          {{ formatRevenue(dashboard?.analytics.estimatedRevenue || 18450000) }}
        </p>
        <p class="mt-1 text-xs text-neutral-400">Ads + Premium + SuperChat (55%/70% split)</p>
      </div>
    </div>

    <!-- Recent Uploads & Copyright/Policy Alerts -->
    <div class="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Recent Uploads -->
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5 lg:col-span-2">
        <div class="flex items-center justify-between border-b border-white/10 pb-3">
          <h2 class="text-sm font-bold text-white">최근 업로드 동영상 (Recent Uploads)</h2>
          <NuxtLink to="/content" class="text-xs font-semibold text-[#3EA6FF] hover:underline">
            콘텐츠 전체 관리 →
          </NuxtLink>
        </div>

        <div class="mt-4 space-y-3">
          <div
            v-for="vid in dashboard?.recentVideos"
            :key="vid.id"
            class="flex flex-wrap items-center justify-between gap-4 rounded-xl bg-white/5 p-3.5"
          >
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-bold text-white">{{ vid.title }}</p>
              <p class="mt-1 text-xs text-neutral-400">
                Views: {{ vid.viewCount.toLocaleString() }} · Likes: {{ vid.likeCount.toLocaleString() }} · Comments: {{ vid.commentCount.toLocaleString() }}
              </p>
            </div>
            <div class="flex items-center gap-2">
              <span class="rounded-full bg-emerald-500/20 px-2.5 py-0.5 text-[11px] font-bold text-emerald-400">
                {{ vid.status }}
              </span>
              <NuxtLink
                :to="`/editor/${vid.id}`"
                class="rounded-lg bg-white/10 px-3 py-1.5 text-xs font-semibold text-white hover:bg-white/20"
              >
                ✂️ Edit
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>

      <!-- Copyright & Policy Status Card -->
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <h2 class="border-b border-white/10 pb-3 text-sm font-bold text-white">
          저작권 및 정책 알림 (Copyright & Policy Alerts)
        </h2>
        <div class="mt-4 space-y-3 text-xs">
          <div class="rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-3 text-emerald-300">
            <p class="font-bold">✓ 채널 상태 양호 (Channel in Good Standing)</p>
            <p class="mt-1 text-neutral-300">커뮤니티 가이드 위반 경고 0건 · 저작권 위반 경고 0건</p>
          </div>

          <div class="rounded-xl border border-amber-400/30 bg-amber-500/10 p-3 text-amber-200">
            <p class="font-bold">ℹ️ Content ID Audio Track Notice</p>
            <p class="mt-1 text-neutral-300">
              Video <code class="text-white">vid_seoul_4k</code> has a non-blocking revenue-share track claim from Synthwave Rights Group.
            </p>
            <NuxtLink to="/monetization" class="mt-2 inline-block font-bold text-[#3EA6FF] hover:underline">
              Review or File Counter-Dispute →
            </NuxtLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
