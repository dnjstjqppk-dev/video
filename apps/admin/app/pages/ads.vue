<script setup lang="ts">
const api = useAdminApi()

useSeoMeta({
  title: 'Ad Delivery Inspector & Telemetry Diagnostics — StreamVerse Admin'
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
    <div class="flex flex-col gap-4 border-b border-slate-200 pb-4 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-900">
          광고 송출 진단 및 인스펙터 (Ad Delivery Inspector)
        </h1>
        <p class="mt-1 text-sm text-slate-600">
          Real-time Ad Engine Decision Telemetry · Fill Rate, Impression Verification &amp; Pacing Diagnostics
        </p>
      </div>
      <button
        type="button"
        @click="refresh()"
        class="inline-flex items-center justify-center gap-2 h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
      >
        <svg class="h-4 w-4 text-slate-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        <span>새로고침 (Refresh Telemetry)</span>
      </button>
    </div>

    <!-- KPI Metric Cards Grid -->
    <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
      <div class="rounded-lg border border-slate-200 bg-white p-4">
        <span class="text-xs font-medium text-slate-500">Total Ad Requests</span>
        <p class="mt-1 text-2xl font-bold text-slate-900 tabular-nums">{{ summary.total_requests.toLocaleString() }}</p>
        <span class="text-xs text-slate-400">All Client Placements</span>
      </div>

      <div class="rounded-lg border border-slate-200 bg-white p-4">
        <span class="text-xs font-medium text-slate-500">Fill Rate</span>
        <p class="mt-1 text-2xl font-bold text-emerald-700 tabular-nums">{{ summary.fill_rate_pct.toFixed(1) }}%</p>
        <span class="text-xs text-slate-400 tabular-nums">{{ summary.served_decisions.toLocaleString() }} served</span>
      </div>

      <div class="rounded-lg border border-slate-200 bg-white p-4">
        <span class="text-xs font-medium text-slate-500">Verified Impressions</span>
        <p class="mt-1 text-2xl font-bold text-slate-900 tabular-nums">{{ summary.verified_impressions.toLocaleString() }}</p>
        <span class="text-xs text-slate-400">Idempotent Deduplicated</span>
      </div>

      <div class="rounded-lg border border-slate-200 bg-white p-4">
        <span class="text-xs font-medium text-slate-500">Clicks / CTR</span>
        <p class="mt-1 text-2xl font-bold text-slate-900 tabular-nums">
          {{ summary.total_clicks.toLocaleString() }}
          <span class="text-sm font-normal text-slate-500 tabular-nums">({{ summary.ctr_pct.toFixed(2) }}%)</span>
        </p>
        <span class="text-xs text-slate-400">Verified Direct Engagement</span>
      </div>

      <div class="rounded-lg border border-slate-200 bg-white p-4">
        <span class="text-xs font-medium text-slate-500">Total Ad Spend</span>
        <p class="mt-1 text-2xl font-bold text-slate-900 tabular-nums">₩{{ (summary.total_spend_cents).toLocaleString() }}</p>
        <span class="text-xs text-slate-400 tabular-nums">{{ summary.active_campaigns_count }} Active Campaigns</span>
      </div>
    </div>

    <!-- Diagnostic Breakdown Sections -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Unfilled Reason Breakdown -->
      <div class="overflow-hidden rounded-lg border border-slate-200 bg-white">
        <div class="border-b border-slate-200 bg-slate-50 px-5 py-3.5">
          <h2 class="text-sm font-semibold text-slate-900">미송출 사유 분석 (Unfilled Reasons)</h2>
          <p class="mt-0.5 text-xs text-slate-500">Breakdown of decisions resulting in no-fill</p>
        </div>

        <div class="divide-y divide-slate-100">
          <div
            v-for="(count, reason) in unfilledReasons"
            :key="reason"
            class="flex items-center justify-between px-5 py-3 text-sm hover:bg-slate-50"
          >
            <span class="font-mono text-xs text-slate-700">{{ reason }}</span>
            <span class="font-semibold text-amber-700 tabular-nums">{{ count }} 건</span>
          </div>
          <div v-if="!Object.keys(unfilledReasons).length" class="py-8 text-center text-xs text-slate-500">
            미송출 기록이 없습니다. (Zero No-Fills)
          </div>
        </div>
      </div>

      <!-- Recent Decisions Feed -->
      <div class="overflow-hidden rounded-lg border border-slate-200 bg-white lg:col-span-2">
        <div class="border-b border-slate-200 bg-slate-50 px-5 py-3.5">
          <h2 class="text-sm font-semibold text-slate-900">최근 송출 의사결정 스트림 (Recent Ad Decisions)</h2>
          <p class="mt-0.5 text-xs text-slate-500">Real-time audit log of CPM auction winners &amp; flight allocations</p>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr>
                <th class="bg-slate-50 border-y border-slate-200 text-xs font-semibold text-slate-600 uppercase tracking-wider py-2.5 px-4">Decision ID</th>
                <th class="bg-slate-50 border-y border-slate-200 text-xs font-semibold text-slate-600 uppercase tracking-wider py-2.5 px-4">Slot</th>
                <th class="bg-slate-50 border-y border-slate-200 text-xs font-semibold text-slate-600 uppercase tracking-wider py-2.5 px-4">Campaign</th>
                <th class="bg-slate-50 border-y border-slate-200 text-xs font-semibold text-slate-600 uppercase tracking-wider py-2.5 px-4">Status</th>
                <th class="bg-slate-50 border-y border-slate-200 text-xs font-semibold text-slate-600 uppercase tracking-wider py-2.5 px-4 text-right">CPM Bid</th>
                <th class="bg-slate-50 border-y border-slate-200 text-xs font-semibold text-slate-600 uppercase tracking-wider py-2.5 px-4">Reason</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="dec in recentDecisions" :key="dec.decision_id" class="hover:bg-slate-50">
                <td class="py-3 px-4 text-sm text-slate-600 border-b border-slate-100 tabular-nums font-mono truncate max-w-[120px]">
                  {{ dec.decision_id }}
                </td>
                <td class="py-3 px-4 text-sm font-semibold text-slate-900 border-b border-slate-100 tabular-nums">
                  {{ dec.slot_id }}
                </td>
                <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums font-mono truncate max-w-[140px]">
                  {{ dec.campaign_id || '-' }}
                </td>
                <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
                  <span
                    class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-semibold uppercase"
                    :class="dec.status === 'served' ? 'border border-emerald-200 bg-emerald-50 text-emerald-700' : 'border border-slate-200 bg-slate-100 text-slate-600'"
                  >
                    {{ dec.status }}
                  </span>
                </td>
                <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums font-mono text-right">
                  ₩{{ dec.bid_cpm_cents?.toLocaleString() || 0 }}
                </td>
                <td class="py-3 px-4 text-sm text-slate-600 border-b border-slate-100 tabular-nums">
                  {{ dec.reason }}
                </td>
              </tr>
              <tr v-if="!recentDecisions.length">
                <td colspan="6" class="py-8 px-4 text-center text-xs text-slate-500">
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
