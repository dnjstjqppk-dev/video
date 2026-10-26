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
  <div>
    <!-- Page Header -->
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-black text-white tracking-tight">광고 송출 진단 및 인스펙터 (Ad Delivery Inspector)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          Real-time Ad Engine Decision Telemetry · Fill Rate, Impression Verification & Pacing Diagnostics
        </p>
      </div>
      <button
        type="button"
        @click="refresh()"
        class="inline-flex items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-xs font-semibold text-white transition hover:bg-white/10"
      >
        <span>🔄 새로고침 (Refresh Telemetry)</span>
      </button>
    </div>

    <!-- KPI Metric Cards Grid -->
    <div class="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-4">
        <span class="text-[11px] font-medium text-neutral-400">Total Ad Requests</span>
        <p class="mt-1 text-2xl font-black text-white">{{ summary.total_requests.toLocaleString() }}</p>
        <span class="text-[10px] text-neutral-500">All Client Placements</span>
      </div>

      <div class="rounded-2xl border border-emerald-500/20 bg-[#141A2C] p-4">
        <span class="text-[11px] font-medium text-emerald-400">Fill Rate</span>
        <p class="mt-1 text-2xl font-black text-emerald-400">{{ summary.fill_rate_pct.toFixed(1) }}%</p>
        <span class="text-[10px] text-neutral-500">{{ summary.served_decisions.toLocaleString() }} served</span>
      </div>

      <div class="rounded-2xl border border-blue-500/20 bg-[#141A2C] p-4">
        <span class="text-[11px] font-medium text-blue-400">Verified Impressions</span>
        <p class="mt-1 text-2xl font-black text-blue-400">{{ summary.verified_impressions.toLocaleString() }}</p>
        <span class="text-[10px] text-neutral-500">Idempotent Deduplicated</span>
      </div>

      <div class="rounded-2xl border border-amber-500/20 bg-[#141A2C] p-4">
        <span class="text-[11px] font-medium text-amber-400">Clicks / CTR</span>
        <p class="mt-1 text-2xl font-black text-amber-400">{{ summary.total_clicks.toLocaleString() }} <span class="text-sm font-normal text-neutral-400">({{ summary.ctr_pct.toFixed(2) }}%)</span></p>
        <span class="text-[10px] text-neutral-500">Verified Direct Engagement</span>
      </div>

      <div class="rounded-2xl border border-purple-500/20 bg-[#141A2C] p-4">
        <span class="text-[11px] font-medium text-purple-400">Total Ad Spend</span>
        <p class="mt-1 text-2xl font-black text-purple-400">₩{{ (summary.total_spend_cents).toLocaleString() }}</p>
        <span class="text-[10px] text-neutral-500">{{ summary.active_campaigns_count }} Active Campaigns</span>
      </div>
    </div>

    <!-- Diagnostic Breakdown Sections -->
    <div class="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Unfilled Reason Breakdown -->
      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-6">
        <h2 class="text-sm font-bold text-white">미송출 사유 분석 (Unfilled Reasons)</h2>
        <p class="mt-1 text-xs text-neutral-400">Breakdown of decisions resulting in no-fill</p>

        <div class="mt-4 space-y-3">
          <div
            v-for="(count, reason) in unfilledReasons"
            :key="reason"
            class="flex items-center justify-between rounded-xl bg-[#0B0E17] p-3 text-xs"
          >
            <span class="font-mono text-neutral-300">{{ reason }}</span>
            <span class="font-bold text-amber-400">{{ count }} 건</span>
          </div>
          <div v-if="!Object.keys(unfilledReasons).length" class="py-6 text-center text-xs text-neutral-500">
            미송출 기록이 없습니다. (Zero No-Fills)
          </div>
        </div>
      </div>

      <!-- Recent Decisions Feed -->
      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-6 lg:col-span-2">
        <h2 class="text-sm font-bold text-white">최근 송출 의사결정 스트림 (Recent Ad Decisions)</h2>
        <p class="mt-1 text-xs text-neutral-400">Real-time audit log of CPM auction winners & flight allocations</p>

        <div class="mt-4 overflow-x-auto">
          <table class="w-full text-left text-xs text-neutral-300">
            <thead class="border-b border-white/10 text-[11px] uppercase text-neutral-400">
              <tr>
                <th class="py-2">Decision ID</th>
                <th class="py-2">Slot</th>
                <th class="py-2">Campaign</th>
                <th class="py-2">Status</th>
                <th class="py-2">CPM Bid</th>
                <th class="py-2">Reason</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-white/5 font-mono">
              <tr v-for="dec in recentDecisions" :key="dec.decision_id" class="hover:bg-white/5">
                <td class="py-2.5 truncate max-w-[120px] text-neutral-400">{{ dec.decision_id }}</td>
                <td class="py-2.5 font-sans font-semibold text-white">{{ dec.slot_id }}</td>
                <td class="py-2.5 truncate max-w-[140px]">{{ dec.campaign_id || '-' }}</td>
                <td class="py-2.5">
                  <span
                    class="rounded px-2 py-0.5 text-[10px] font-bold uppercase"
                    :class="dec.status === 'served' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-neutral-800 text-neutral-400'"
                  >
                    {{ dec.status }}
                  </span>
                </td>
                <td class="py-2.5 text-neutral-200">₩{{ dec.bid_cpm_cents?.toLocaleString() || 0 }}</td>
                <td class="py-2.5 font-sans text-neutral-400">{{ dec.reason }}</td>
              </tr>
              <tr v-if="!recentDecisions.length">
                <td colspan="6" class="py-8 text-center text-xs text-neutral-500">
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
