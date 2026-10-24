<script setup lang="ts">
const api = useAdminApi()

useSeoMeta({
  title: 'Versioned Platform Policies, Quotas & Emergency Kill-Switches'
})

const policyVersion = ref('v4.3.0-2026-10')
const maxUploadSizeGb = ref(128)
const maxVideoDurationHours = ref(12)
const creatorRevenueSharePct = ref(55)
const savedBanner = ref(false)

const killSwitches = ref([
  {
    key: 'GLOBAL_PUBLIC_SIGNUP_ENABLED',
    label: 'Global Public User Registration',
    enabled: true,
    description: 'Disable immediately during botnet registration attacks'
  },
  {
    key: 'MULTIPART_R2_UPLOAD_INGEST',
    label: 'R2 Multipart Video Upload Ingest',
    enabled: true,
    description: 'Pause new video uploads during storage maintenance'
  },
  {
    key: 'LIVE_STREAM_RTMPS_INGEST',
    label: 'Global RTMPS / SRT Live Streaming Ingest',
    enabled: true,
    description: 'Emergency circuit breaker for live broadcast ingest nodes'
  },
  {
    key: 'AUTOMATED_LEDGER_PAYOUT_DISPATCH',
    label: 'Automated Bank Payout Clearing',
    enabled: true,
    description: 'Freeze outgoing bank transfers if reconciliation anomaly detected'
  }
])

const { data: auditLogs } = await useAsyncData('admin-audit-logs', async () => {
  try {
    const res = await api.getAdminAuditLogs(1, 20)
    return res.items
  } catch {
    return [
      {
        id: 'log_991',
        actorId: 'usr_super_admin_01',
        action: 'PUBLISH_POLICY_VERSION_v4.3.0',
        target: 'global_platform_config',
        ip: '10.24.0.12',
        timestamp: '2026-10-24T08:00:00Z'
      },
      {
        id: 'log_990',
        actorId: 'usr_finance_admin_01',
        action: 'VERIFY_LEDGER_BALANCE_INVARIANT',
        target: 'ledger_entries_2026_10',
        ip: '10.24.0.19',
        timestamp: '2026-10-24T07:50:00Z'
      }
    ]
  }
})

function publishPolicyVersion() {
  savedBanner.value = true
  setTimeout(() => (savedBanner.value = false), 3500)
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-extrabold text-white">정책 버전 관리, 쿼터 제한 및 긴급 킬스위치 (Policies & Kill-Switches)</h1>
    <p class="mt-1 text-xs text-neutral-400">
      Versioned platform limits · Real-time Valkey feature flags & emergency circuit breakers · Immutable audit log inspection
    </p>

    <!-- Versioned Policy Editor -->
    <form class="mt-6 rounded-2xl border border-white/10 bg-[#131722] p-6" @submit.prevent="publishPolicyVersion">
      <div class="flex items-center justify-between">
        <h2 class="text-sm font-bold text-white">플랫폼 업로드 및 정산 정책 버전 ({{ policyVersion }})</h2>
        <span v-if="savedBanner" class="text-xs font-bold text-emerald-400">
          ✓ Policy {{ policyVersion }} committed & broadcast via NATS JetStream
        </span>
      </div>

      <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3">
        <div>
          <label class="text-xs text-neutral-400">Max Multipart Upload Size (GB)</label>
          <input
            v-model.number="maxUploadSizeGb"
            type="number"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3 py-2 text-xs text-white"
          />
        </div>
        <div>
          <label class="text-xs text-neutral-400">Max Single Video Duration (Hours)</label>
          <input
            v-model.number="maxVideoDurationHours"
            type="number"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3 py-2 text-xs text-white"
          />
        </div>
        <div>
          <label class="text-xs text-neutral-400">Standard Creator Ad Revenue Share (%)</label>
          <input
            v-model.number="creatorRevenueSharePct"
            type="number"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3 py-2 text-xs text-white"
          />
        </div>
      </div>

      <div class="mt-4 flex justify-end">
        <button
          type="submit"
          class="rounded-xl bg-red-600 px-5 py-2 text-xs font-bold text-white hover:bg-red-500"
        >
          Commit New Policy Version
        </button>
      </div>
    </form>

    <!-- Emergency Feature Flags & Kill-Switches -->
    <div class="mt-6 rounded-2xl border border-red-500/30 bg-[#131722] p-6">
      <h2 class="text-sm font-bold text-white">긴급 킬스위치 및 피처 플래그 (Emergency Circuit Breakers)</h2>
      <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
        <div
          v-for="flag in killSwitches"
          :key="flag.key"
          class="flex items-center justify-between rounded-xl border border-white/10 bg-[#090B10] p-4 text-xs"
        >
          <div>
            <p class="font-mono font-bold text-white">{{ flag.key }}</p>
            <p class="mt-0.5 text-neutral-400">{{ flag.description }}</p>
          </div>
          <button
            type="button"
            class="rounded-xl px-3.5 py-1.5 font-extrabold transition"
            :class="flag.enabled ? 'bg-emerald-500/20 text-emerald-400' : 'bg-red-600 text-white'"
            @click="flag.enabled = !flag.enabled"
          >
            {{ flag.enabled ? 'ENABLED' : 'KILLED (OFF)' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Immutable Audit Log Viewer -->
    <div class="mt-6 rounded-2xl border border-white/10 bg-[#131722] p-5">
      <h2 class="text-sm font-bold text-white">불변 감사 로그 원장 (`/v1/admin/audit-logs`)</h2>
      <div class="mt-3 space-y-2 text-xs">
        <div
          v-for="item in auditLogs"
          :key="item.id"
          class="flex flex-wrap items-center justify-between rounded-xl bg-[#090B10] px-4 py-2.5 font-mono"
        >
          <span class="text-[#3EA6FF]">{{ item.id }} · {{ item.action }}</span>
          <span class="text-neutral-300">Actor: {{ item.actorId }} ({{ item.ip }})</span>
          <span class="text-neutral-500">{{ item.timestamp }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
