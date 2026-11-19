<script setup lang="ts">
const api = useAdminApi()

useSeoMeta({
  title: 'Ad Delivery Inspector & Telemetry Diagnostics — vravra Admin'
})

const { data: diagnosticsData, pending, error, refresh } = await useAsyncData('admin-ad-diagnostics', async () => {
  try {
    return await api.getAdDeliveryDiagnostics()
  } catch (err) {
    return null
  }
})

const summary = computed(() => diagnosticsData.value?.summary || {
  total_requests: 0,
  served_decisions: 0,
  unfilled_decisions: 0,
  fill_rate_pct: 0,
  verified_impressions: 0,
  total_clicks: 0,
  ctr_pct: 0,
  total_spend_cents: 0,
  active_campaigns_count: 0,
  total_registered_creatives: 0
})

const unfilledReasons = computed(() => diagnosticsData.value?.unfilled_reasons || {})
const recentDecisions = computed(() => diagnosticsData.value?.recent_decisions || [])
</script>

<template>
  <div class="space-y-6">
    <!-- Page Header -->
    <div class="flex flex-col gap-4 border-b border-[var(--gvp-border)] pb-4 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-[var(--gvp-text)]">
          광고 송출 진단 및 인스펙터 (Ad Delivery Inspector)
        </h1>
        <p class="mt-1 text-sm text-[var(--gvp-text-secondary)]">
          Real-time Ad Engine Decision Telemetry · Fill Rate, Impression Verification &amp; Pacing Diagnostics
        </p>
      </div>
      <button
        type="button"
        @click="refresh()"
        class="inline-flex items-center justify-center gap-2 h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
      >
        <svg class="h-4 w-4 text-[var(--gvp-text-secondary)]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        <span>새로고침 (Refresh Telemetry)</span>
      </button>
    </div>

    <!-- KPI Metric Cards Grid -->
    <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4">
        <span class="text-xs font-medium text-[var(--gvp-text-secondary)]">Total Ad Requests</span>
        <p class="mt-1 text-2xl font-bold text-[var(--gvp-text)] tabular-nums">{{ summary.total_requests.toLocaleString() }}</p>
        <span class="text-xs text-[var(--gvp-text-muted)]">All Client Placements</span>
      </div>

      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4">
        <span class="text-xs font-medium text-[var(--gvp-text-secondary)]">Fill Rate</span>
        <p class="mt-1 text-2xl font-bold text-[var(--gvp-success)] tabular-nums">{{ summary.fill_rate_pct.toFixed(1) }}%</p>
        <span class="text-xs text-[var(--gvp-text-muted)] tabular-nums">{{ summary.served_decisions.toLocaleString() }} served</span>
      </div>

      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4">
        <span class="text-xs font-medium text-[var(--gvp-text-secondary)]">Verified Impressions</span>
        <p class="mt-1 text-2xl font-bold text-[var(--gvp-text)] tabular-nums">{{ summary.verified_impressions.toLocaleString() }}</p>
        <span class="text-xs text-[var(--gvp-text-muted)]">Idempotent Deduplicated</span>
      </div>

      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4">
        <span class="text-xs font-medium text-[var(--gvp-text-secondary)]">Clicks / CTR</span>
        <p class="mt-1 text-2xl font-bold text-[var(--gvp-text)] tabular-nums">
          {{ summary.total_clicks.toLocaleString() }}
          <span class="text-sm font-normal text-[var(--gvp-text-secondary)] tabular-nums">({{ summary.ctr_pct.toFixed(2) }}%)</span>
        </p>
        <span class="text-xs text-[var(--gvp-text-muted)]">Verified Direct Engagement</span>
      </div>

      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4">
        <span class="text-xs font-medium text-[var(--gvp-text-secondary)]">Total Ad Spend</span>
        <p class="mt-1 text-2xl font-bold text-[var(--gvp-text)] tabular-nums">₩{{ (summary.total_spend_cents).toLocaleString() }}</p>
        <span class="text-xs text-[var(--gvp-text-muted)] tabular-nums">{{ summary.active_campaigns_count }} Active Campaigns</span>
      </div>
    </div>

    <!-- Diagnostic Breakdown Sections -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Unfilled Reason Breakdown -->
      <div class="overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
        <div class="border-b border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-5 py-3.5">
          <h2 class="text-sm font-semibold text-[var(--gvp-text)]">미송출 사유 분석 (Unfilled Reasons)</h2>
          <p class="mt-0.5 text-xs text-[var(--gvp-text-secondary)]">Breakdown of decisions resulting in no-fill</p>
        </div>

        <div class="divide-y divide-[var(--gvp-border)]">
          <div
            v-for="(count, reason) in unfilledReasons"
            :key="reason"
            class="flex items-center justify-between px-5 py-3 text-sm hover:bg-[var(--gvp-canvas)]"
          >
            <span class="font-mono text-xs text-[var(--gvp-text)]">{{ reason }}</span>
            <span class="font-semibold text-[var(--gvp-warning)] tabular-nums">{{ count }} 건</span>
          </div>
          <div v-if="!Object.keys(unfilledReasons).length" class="py-8 text-center text-xs text-[var(--gvp-text-secondary)]">
            미송출 기록이 없습니다. (Zero No-Fills)
          </div>
        </div>
      </div>

      <!-- Recent Decisions Feed -->
      <div class="overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] lg:col-span-2">
        <div class="border-b border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-5 py-3.5">
          <h2 class="text-sm font-semibold text-[var(--gvp-text)]">최근 송출 의사결정 스트림 (Recent Ad Decisions)</h2>
          <p class="mt-0.5 text-xs text-[var(--gvp-text-secondary)]">Real-time audit log of CPM auction winners &amp; flight allocations</p>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr>
                <th class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider py-2.5 px-4">Decision ID</th>
                <th class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider py-2.5 px-4">Slot</th>
                <th class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider py-2.5 px-4">Campaign</th>
                <th class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider py-2.5 px-4">Status</th>
                <th class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider py-2.5 px-4 text-right">CPM Bid</th>
                <th class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider py-2.5 px-4">Reason</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[var(--gvp-border)]">
              <tr v-for="dec in recentDecisions" :key="dec.decision_id" class="hover:bg-[var(--gvp-canvas)]">
                <td class="py-3 px-4 text-sm text-[var(--gvp-text-secondary)] border-b border-[var(--gvp-border)] tabular-nums font-mono truncate max-w-[120px]">
                  {{ dec.decision_id }}
                </td>
                <td class="py-3 px-4 text-sm font-semibold text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
                  {{ dec.slot_id }}
                </td>
                <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums font-mono truncate max-w-[140px]">
                  {{ dec.campaign_id || '-' }}
                </td>
                <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
                  <span
                    class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-semibold uppercase"
                    :class="dec.status === 'served' ? 'border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] text-[var(--gvp-success)]' : 'border border-[var(--gvp-border)] bg-[var(--gvp-elevated)] text-[var(--gvp-text-secondary)]'"
                  >
                    {{ dec.status }}
                  </span>
                </td>
                <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums font-mono text-right">
                  ₩{{ dec.bid_cpm_cents?.toLocaleString() || 0 }}
                </td>
                <td class="py-3 px-4 text-sm text-[var(--gvp-text-secondary)] border-b border-[var(--gvp-border)] tabular-nums">
                  {{ dec.reason }}
                </td>
              </tr>
              <tr v-if="!recentDecisions.length">
                <td colspan="6" class="py-8 px-4 text-center text-xs text-[var(--gvp-text-secondary)]">
                  최근 기록된 결정 로그가 없습니다. (Empty Decision Stream)
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
