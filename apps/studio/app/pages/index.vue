<script setup lang="ts">
import { VideoProcessingStatus } from '@gvp/shared-types'

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
const activeClaims = computed(() => (claimsData.value || []).filter(c => c.status !== 'resolved'))

function formatRevenue(krw: number): string {
  if (currency.value === 'KRW') {
    return `₩${krw.toLocaleString('ko-KR')}`
  }
  return `$${(krw / 1350).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 pb-5">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-slate-900">채널 대시보드 (Creator Dashboard)</h1>
        <p class="mt-1 text-xs text-slate-500">
          Real-time ClickHouse telemetry · Last 28 days performance &amp; double-entry settlement accruals
        </p>
      </div>
      <div class="flex items-center gap-2">
        <NuxtLink
          to="/upload"
          class="inline-flex items-center justify-center gap-2 h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors"
        >
          <svg class="h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 10.5V2.5M8 2.5L4.75 5.75M8 2.5L11.25 5.75M2.5 11v1.5a1 1 0 001 1h9a1 1 0 001-1V11" />
          </svg>
          <span>Upload Video</span>
        </NuxtLink>
        <NuxtLink
          to="/live"
          class="inline-flex items-center justify-center gap-2 h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
        >
          <svg class="h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
            <circle cx="8" cy="8" r="2" />
            <path stroke-linecap="round" d="M4.93 4.93a4.34 4.34 0 000 6.14M11.07 4.93a4.34 4.34 0 010 6.14M2.8 2.8a7.35 7.35 0 000 10.4M13.2 2.8a7.35 7.35 0 010 10.4" />
          </svg>
          <span>Go Live</span>
        </NuxtLink>
      </div>
    </div>

    <div v-if="dashPending" class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <div v-for="i in 4" :key="i" class="h-28 animate-pulse rounded-lg border border-slate-200 bg-slate-100" />
    </div>
    <div v-else-if="dashError || !dashboard" class="rounded-lg border border-red-200 bg-red-50 p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-red-700">대시보드를 불러오지 못했습니다.</p>
    </div>

    <!-- KPI Cards -->
    <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">총 조회수 (28d Views)</p>
        <p class="mt-2 text-2xl font-semibold text-slate-900 tabular-nums">
          {{ dashboard?.analytics.views.toLocaleString() }}
        </p>
        <p class="mt-1 text-xs font-medium text-slate-500 tabular-nums">최근 28일 집계</p>
      </div>

      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">시청 시간 (Watch Time Hours)</p>
        <p class="mt-2 text-2xl font-semibold text-slate-900 tabular-nums">
          {{ dashboard?.analytics.watchTimeHours.toLocaleString() }} hrs
        </p>
        <p class="mt-1 text-xs font-medium text-slate-500 tabular-nums">최근 28일 집계</p>
      </div>

      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">구독자 수 (Subscribers)</p>
        <p class="mt-2 text-2xl font-semibold text-slate-900 tabular-nums">
          {{ dashboard?.analytics.subscribers.toLocaleString() }}
        </p>
        <p class="mt-1 text-xs font-medium text-slate-500 tabular-nums">채널 누적 구독자</p>
      </div>

      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">예상 정산 수익 (Est. Net Revenue)</p>
        <p class="mt-2 text-2xl font-semibold text-slate-900 tabular-nums">
          {{ formatRevenue(dashboard?.analytics.estimatedRevenue ?? 0) }}
        </p>
        <p class="mt-1 text-xs text-slate-500 tabular-nums">Ads + Premium + SuperChat (55%/70% split)</p>
      </div>
    </div>

    <!-- Recent Uploads & Copyright/Policy Alerts -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Recent Uploads -->
      <div v-if="(dashboard?.recentVideos || []).length === 0" class="rounded-lg border border-slate-200 bg-white p-8 text-center lg:col-span-2">
        <p class="text-sm font-semibold text-slate-600">아직 업로드한 동영상이 없습니다.</p>
        <NuxtLink to="/upload" class="mt-2 inline-block text-xs font-medium text-slate-900 underline">첫 동영상 업로드하기</NuxtLink>
      </div>
      <div v-else class="overflow-hidden rounded-lg border border-slate-200 bg-white lg:col-span-2">
        <div class="flex items-center justify-between border-b border-slate-200 px-5 py-4">
          <h2 class="text-sm font-semibold text-slate-900">최근 업로드 동영상 (Recent Uploads)</h2>
          <NuxtLink to="/content" class="text-xs font-medium text-slate-600 hover:text-slate-900">
            콘텐츠 전체 관리
          </NuxtLink>
        </div>

        <table class="w-full text-left">
          <thead class="bg-slate-50 border-b border-slate-200 text-xs font-semibold text-slate-600 uppercase tracking-wider">
            <tr>
              <th scope="col" class="py-2.5 px-4">Video</th>
              <th scope="col" class="py-2.5 px-4">Engagement</th>
              <th scope="col" class="py-2.5 px-4">Status</th>
              <th scope="col" class="py-2.5 px-4 text-right">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr
              v-for="vid in dashboard?.recentVideos"
              :key="vid.id"
              class="hover:bg-slate-50 transition-colors"
            >
              <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100">
                <p class="font-medium text-slate-900 line-clamp-1">{{ vid.title }}</p>
              </td>
              <td class="py-3 px-4 text-sm text-slate-600 border-b border-slate-100 tabular-nums whitespace-nowrap">
                {{ vid.viewCount.toLocaleString() }} views · {{ vid.likeCount.toLocaleString() }} likes · {{ vid.commentCount.toLocaleString() }} comments
              </td>
              <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
                <span class="inline-flex items-center rounded-md border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700">
                  {{ vid.status }}
                </span>
              </td>
              <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums text-right">
                <NuxtLink
                  :to="`/editor/${vid.id}`"
                  class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
                >
                  Edit
                </NuxtLink>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Copyright & Policy Status Card -->
      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <h2 class="border-b border-slate-200 pb-3 text-sm font-semibold text-slate-900">
          저작권 및 정책 알림 (Copyright &amp; Policy Alerts)
        </h2>
        <div class="mt-4 space-y-3 text-xs">
          <div v-if="activeClaims.length === 0" class="rounded-md border border-emerald-200 bg-emerald-50 p-3.5 text-emerald-900">
            <p class="font-semibold">채널 상태 양호 (Channel in Good Standing)</p>
            <p class="mt-1 text-emerald-700 tabular-nums">진행 중인 저작권 클레임이 없습니다.</p>
          </div>

          <div v-for="claim in activeClaims" :key="claim.id" class="rounded-md border border-amber-200 bg-amber-50 p-3.5 text-amber-900">
            <p class="font-semibold">Content ID Claim · {{ claim.id }}</p>
            <p class="mt-1 text-amber-800">
              Video <code class="rounded-md bg-white px-1.5 py-0.5 font-mono text-slate-900 border border-amber-200">{{ claim.claimedVideoId }}</code>
              has a {{ claim.claimAction }} claim from {{ claim.ownerName || claim.rightsOwnerId }}.
            </p>
            <NuxtLink to="/monetization" class="mt-2 inline-block font-semibold text-slate-900 underline hover:text-slate-700">
              Review or File Counter-Dispute
            </NuxtLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
