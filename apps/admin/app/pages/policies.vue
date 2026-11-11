<script setup lang="ts">
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'

const api = useAdminApi()

useSeoMeta({
  title: 'Versioned Platform Policies, Quotas & Emergency Kill-Switches'
})

const maxUploadSizeGb = ref(0)
const maxVideoDurationHours = ref(0)
const creatorRevenueSharePct = ref(0)
const savedBanner = ref(false)
const saving = ref(false)
const actionError = ref<string | null>(null)
const togglingKeys = ref<Set<string>>(new Set())

interface PolicyFlag {
  key: string
  label: string
  description: string
  enabled: boolean
  updated_by: string
  updated_at: string
}

const flags = ref<PolicyFlag[]>([])

const { data, pending, error, refresh } = await useAsyncData('admin-policies', async () => {
  return await api.getAdminPolicies()
})

watch(data, (policies) => {
  if (!policies) return
  maxUploadSizeGb.value = policies.config.max_upload_size_gb
  maxVideoDurationHours.value = policies.config.max_video_duration_hours
  creatorRevenueSharePct.value = policies.config.creator_revenue_share_pct
  flags.value = policies.flags.map(f => ({ ...f }))
}, { immediate: true })

const policyVersion = computed(() => data.value?.config.version || '')

const { data: auditLogs, pending: auditPending } = await useAsyncData('admin-policy-audit', async () => {
  const res = await api.getAdminAuditLogs(1, 20)
  return res.items
})

async function publishPolicyVersion() {
  saving.value = true
  savedBanner.value = false
  actionError.value = null
  try {
    const updated = await api.updateAdminPolicyConfig({
      max_upload_size_gb: maxUploadSizeGb.value,
      max_video_duration_hours: maxVideoDurationHours.value,
      creator_revenue_share_pct: creatorRevenueSharePct.value,
    })
    maxUploadSizeGb.value = updated.max_upload_size_gb
    maxVideoDurationHours.value = updated.max_video_duration_hours
    creatorRevenueSharePct.value = updated.creator_revenue_share_pct
    savedBanner.value = true
    setTimeout(() => { savedBanner.value = false }, 3500)
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '정책 저장에 실패했습니다.'
  } finally {
    saving.value = false
  }
}

async function toggleFlag(flag: PolicyFlag) {
  togglingKeys.value.add(flag.key)
  actionError.value = null
  const prev = flag.enabled
  try {
    const res = await api.updateAdminFlag(flag.key, !flag.enabled)
    flag.enabled = res.enabled
  } catch (e: unknown) {
    flag.enabled = prev
    actionError.value = e instanceof Error ? e.message : `플래그 ${flag.key} 변경에 실패했습니다.`
  } finally {
    togglingKeys.value.delete(flag.key)
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="border-b border-slate-200 pb-4">
      <h1 class="text-2xl font-bold tracking-tight text-slate-900">
        정책 버전 관리, 쿼터 제한 및 긴급 킬스위치 (Policies &amp; Kill-Switches)
      </h1>
      <p class="mt-1 text-sm text-slate-600">
        Versioned platform limits · Real-time feature flags &amp; emergency circuit breakers · Immutable audit log inspection
      </p>
    </div>

    <GvpBanner v-if="actionError" tone="danger" dismissible @dismiss="actionError = null">
      {{ actionError }}
    </GvpBanner>

    <div v-if="pending" class="space-y-6" role="status" aria-label="불러오는 중">
      <div class="h-56 animate-pulse rounded-lg bg-slate-100" />
      <div class="h-64 animate-pulse rounded-lg bg-slate-100" />
    </div>

    <GvpBanner v-else-if="error" tone="danger">
      정책을 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <template v-else>
      <!-- Versioned Policy Editor -->
      <form class="overflow-hidden rounded-lg border border-slate-200 bg-white" @submit.prevent="publishPolicyVersion">
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-slate-200 bg-slate-50 px-5 py-3.5">
          <h2 class="text-sm font-semibold text-slate-900 tabular-nums">
            플랫폼 업로드 및 정산 정책 버전 ({{ policyVersion }})
          </h2>
          <span v-if="savedBanner" class="rounded-md border border-emerald-200 bg-emerald-50 px-2.5 py-0.5 text-xs font-semibold text-emerald-700 tabular-nums" role="status">
            Policy {{ policyVersion }} committed
          </span>
        </div>

        <div class="p-5">
          <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
            <div>
              <label for="pol-upload" class="block text-xs font-medium text-slate-600">Max Multipart Upload Size (GB)</label>
              <input
                id="pol-upload"
                v-model.number="maxUploadSizeGb"
                type="number"
                min="1"
                class="mt-1.5 h-9 w-full rounded-md border border-slate-300 bg-slate-50 px-3 text-sm text-slate-900 tabular-nums focus:border-slate-900 focus:bg-white outline-none transition-colors"
              />
            </div>
            <div>
              <label for="pol-duration" class="block text-xs font-medium text-slate-600">Max Single Video Duration (Hours)</label>
              <input
                id="pol-duration"
                v-model.number="maxVideoDurationHours"
                type="number"
                min="1"
                class="mt-1.5 h-9 w-full rounded-md border border-slate-300 bg-slate-50 px-3 text-sm text-slate-900 tabular-nums focus:border-slate-900 focus:bg-white outline-none transition-colors"
              />
            </div>
            <div>
              <label for="pol-share" class="block text-xs font-medium text-slate-600">Standard Creator Ad Revenue Share (%)</label>
              <input
                id="pol-share"
                v-model.number="creatorRevenueSharePct"
                type="number"
                min="0"
                max="100"
                class="mt-1.5 h-9 w-full rounded-md border border-slate-300 bg-slate-50 px-3 text-sm text-slate-900 tabular-nums focus:border-slate-900 focus:bg-white outline-none transition-colors"
              />
            </div>
          </div>

          <div class="mt-4 flex justify-end">
            <button
              type="submit"
              :disabled="saving"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors disabled:opacity-50"
            >
              {{ saving ? '저장 중…' : 'Commit New Policy Version' }}
            </button>
          </div>
        </div>
      </form>

      <!-- Emergency Feature Flags & Kill-Switches -->
      <div class="overflow-hidden rounded-lg border border-slate-200 bg-white">
        <div class="border-b border-slate-200 bg-slate-50 px-5 py-3.5">
          <h2 class="text-sm font-semibold text-slate-900">
            긴급 킬스위치 및 피처 플래그 (Emergency Circuit Breakers)
          </h2>
        </div>

        <div class="p-5">
          <GvpEmptyState
            v-if="flags.length === 0"
            title="등록된 플래그가 없습니다"
          />
          <div v-else class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div
              v-for="flag in flags"
              :key="flag.key"
              class="flex items-center justify-between rounded-lg border border-slate-200 bg-slate-50 p-4 text-xs"
            >
              <div class="min-w-0">
                <p class="font-mono text-sm font-semibold text-slate-900">{{ flag.key }}</p>
                <p class="mt-0.5 text-xs text-slate-600">{{ flag.description || flag.label }}</p>
              </div>
              <button
                type="button"
                :disabled="togglingKeys.has(flag.key)"
                class="ml-3 shrink-0 inline-flex items-center justify-center h-9 px-4 text-xs font-semibold rounded-md transition-colors disabled:opacity-40"
                :class="flag.enabled ? 'border border-emerald-200 bg-emerald-50 text-emerald-700 hover:bg-emerald-100' : 'bg-slate-900 text-white hover:bg-slate-800'"
                @click="toggleFlag(flag)"
              >
                {{ flag.enabled ? 'ENABLED' : 'KILLED (OFF)' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Immutable Audit Log Viewer -->
    <div class="overflow-hidden rounded-lg border border-slate-200 bg-white">
      <div class="border-b border-slate-200 bg-slate-50 px-5 py-3.5">
        <h2 class="text-sm font-semibold text-slate-900">
          불변 감사 로그 원장 (<span class="font-mono text-xs">/v1/admin/audit-logs</span>)
        </h2>
      </div>
      <div v-if="auditPending" class="p-4 space-y-2" role="status" aria-label="감사 기록 불러오는 중">
        <div v-for="n in 3" :key="n" class="h-10 animate-pulse rounded-md bg-slate-100" />
      </div>
      <div v-else-if="!auditLogs || auditLogs.length === 0" class="p-5 text-xs text-slate-500">
        기록된 감사 로그가 없습니다.
      </div>
      <div v-else class="divide-y divide-slate-100 text-xs">
        <div
          v-for="item in auditLogs"
          :key="item.id"
          class="flex flex-wrap items-center justify-between gap-2 px-5 py-3 font-mono hover:bg-slate-50"
        >
          <span class="font-semibold text-slate-900">{{ item.id }} · {{ item.action }}</span>
          <span class="text-slate-600 tabular-nums">Actor: {{ item.actorId }} ({{ item.ip }})</span>
          <span class="text-slate-500 tabular-nums">{{ item.timestamp }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
