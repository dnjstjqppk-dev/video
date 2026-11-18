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
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-[var(--gvp-border)] pb-5">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">채널 대시보드 (Creator Dashboard)</h1>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          Real-time ClickHouse telemetry · Last 28 days performance &amp; double-entry settlement accruals
        </p>
      </div>
      <div class="flex items-center gap-2">
        <NuxtLink
          to="/upload"
          class="inline-flex items-center justify-center gap-2 h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors"
        >
          <svg class="h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 10.5V2.5M8 2.5L4.75 5.75M8 2.5L11.25 5.75M2.5 11v1.5a1 1 0 001 1h9a1 1 0 001-1V11" />
          </svg>
          <span>Upload Video</span>
        </NuxtLink>
        <NuxtLink
          to="/live"
          class="inline-flex items-center justify-center gap-2 h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
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
      <div v-for="i in 4" :key="i" class="h-28 animate-pulse rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-elevated)]" />
    </div>
    <div v-else-if="dashError || !dashboard" class="rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-[var(--gvp-live)]">대시보드를 불러오지 못했습니다.</p>
    </div>

    <!-- KPI Cards -->
    <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5">
        <p class="text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-secondary)]">총 조회수 (28d Views)</p>
        <p class="mt-2 text-2xl font-semibold text-[var(--gvp-text)] tabular-nums">
          {{ dashboard?.analytics.views.toLocaleString() }}
        </p>
        <p class="mt-1 text-xs font-medium text-[var(--gvp-text-secondary)] tabular-nums">최근 28일 집계</p>
      </div>

      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5">
        <p class="text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-secondary)]">시청 시간 (Watch Time Hours)</p>
        <p class="mt-2 text-2xl font-semibold text-[var(--gvp-text)] tabular-nums">
          {{ dashboard?.analytics.watchTimeHours.toLocaleString() }} hrs
        </p>
        <p class="mt-1 text-xs font-medium text-[var(--gvp-text-secondary)] tabular-nums">최근 28일 집계</p>
      </div>

      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5">
        <p class="text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-secondary)]">구독자 수 (Subscribers)</p>
        <p class="mt-2 text-2xl font-semibold text-[var(--gvp-text)] tabular-nums">
          {{ dashboard?.analytics.subscribers.toLocaleString() }}
        </p>
        <p class="mt-1 text-xs font-medium text-[var(--gvp-text-secondary)] tabular-nums">채널 누적 구독자</p>
      </div>

      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5">
        <p class="text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-secondary)]">예상 정산 수익 (Est. Net Revenue)</p>
        <p class="mt-2 text-2xl font-semibold text-[var(--gvp-text)] tabular-nums">
          {{ formatRevenue(dashboard?.analytics.estimatedRevenue ?? 0) }}
        </p>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)] tabular-nums">Ads + Premium + SuperChat (55%/70% split)</p>
      </div>
    </div>

    <!-- Recent Uploads & Copyright/Policy Alerts -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Recent Uploads -->
      <div v-if="(dashboard?.recentVideos || []).length === 0" class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-8 text-center lg:col-span-2">
        <p class="text-sm font-semibold text-[var(--gvp-text-secondary)]">아직 업로드한 동영상이 없습니다.</p>
        <NuxtLink to="/upload" class="mt-2 inline-block text-xs font-medium text-[var(--gvp-text)] underline">첫 동영상 업로드하기</NuxtLink>
      </div>
      <div v-else class="overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] lg:col-span-2">
        <div class="flex items-center justify-between border-b border-[var(--gvp-border)] px-5 py-4">
          <h2 class="text-sm font-semibold text-[var(--gvp-text)]">최근 업로드 동영상 (Recent Uploads)</h2>
          <NuxtLink to="/content" class="text-xs font-medium text-[var(--gvp-text-secondary)] hover:text-[var(--gvp-text)]">
            콘텐츠 전체 관리
          </NuxtLink>
        </div>

        <table class="w-full text-left">
          <thead class="bg-[var(--gvp-canvas)] border-b border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider">
            <tr>
              <th scope="col" class="py-2.5 px-4">Video</th>
              <th scope="col" class="py-2.5 px-4">Engagement</th>
              <th scope="col" class="py-2.5 px-4">Status</th>
              <th scope="col" class="py-2.5 px-4 text-right">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--gvp-border)]">
            <tr
              v-for="vid in dashboard?.recentVideos"
              :key="vid.id"
              class="hover:bg-[var(--gvp-canvas)] transition-colors"
            >
              <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)]">
                <p class="font-medium text-[var(--gvp-text)] line-clamp-1">{{ vid.title }}</p>
              </td>
              <td class="py-3 px-4 text-sm text-[var(--gvp-text-secondary)] border-b border-[var(--gvp-border)] tabular-nums whitespace-nowrap">
                {{ vid.viewCount.toLocaleString() }} views · {{ vid.likeCount.toLocaleString() }} likes · {{ vid.commentCount.toLocaleString() }} comments
              </td>
              <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
                <span class="inline-flex items-center rounded-md border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] px-2 py-0.5 text-xs font-medium text-[var(--gvp-success)]">
                  {{ vid.status }}
                </span>
              </td>
              <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums text-right">
                <NuxtLink
                  :to="`/editor/${vid.id}`"
                  class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
                >
                  Edit
                </NuxtLink>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Copyright & Policy Status Card -->
      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5">
        <h2 class="border-b border-[var(--gvp-border)] pb-3 text-sm font-semibold text-[var(--gvp-text)]">
          저작권 및 정책 알림 (Copyright &amp; Policy Alerts)
        </h2>
        <div class="mt-4 space-y-3 text-xs">
          <div v-if="activeClaims.length === 0" class="rounded-md border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] p-3.5 text-[var(--gvp-success)]">
            <p class="font-semibold">채널 상태 양호 (Channel in Good Standing)</p>
            <p class="mt-1 text-[var(--gvp-success)] tabular-nums">진행 중인 저작권 클레임이 없습니다.</p>
          </div>

          <div v-for="claim in activeClaims" :key="claim.id" class="rounded-md border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] p-3.5 text-[var(--gvp-warning)]">
            <p class="font-semibold">Content ID Claim · {{ claim.id }}</p>
            <p class="mt-1 text-[var(--gvp-warning)]">
              Video <code class="rounded-md bg-[var(--gvp-surface)] px-1.5 py-0.5 font-mono text-[var(--gvp-text)] border border-[var(--gvp-warning)]">{{ claim.claimedVideoId }}</code>
              has a {{ claim.claimAction }} claim from {{ claim.ownerName || claim.rightsOwnerId }}.
            </p>
            <NuxtLink to="/monetization" class="mt-2 inline-block font-semibold text-[var(--gvp-text)] underline hover:text-[var(--gvp-text)]">
              Review or File Counter-Dispute
            </NuxtLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
