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
  <div>
    <h1 class="text-2xl font-extrabold text-white">정책 버전 관리, 쿼터 제한 및 긴급 킬스위치 (Policies & Kill-Switches)</h1>
    <p class="mt-1 text-xs text-neutral-400">
      Versioned platform limits · Real-time feature flags & emergency circuit breakers · Immutable audit log inspection
    </p>

    <GvpBanner v-if="actionError" tone="danger" dismissible class="mt-4" @dismiss="actionError = null">
      {{ actionError }}
    </GvpBanner>

    <div v-if="pending" class="mt-6 space-y-6" role="status" aria-label="불러오는 중">
      <div class="h-56 animate-pulse rounded-2xl bg-white/5" />
      <div class="h-64 animate-pulse rounded-2xl bg-white/5" />
    </div>

    <GvpBanner v-else-if="error" tone="danger" class="mt-6">
      정책을 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <template v-else>
      <!-- Versioned Policy Editor -->
      <form class="mt-6 rounded-2xl border border-white/10 bg-[#131722] p-6" @submit.prevent="publishPolicyVersion">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <h2 class="text-sm font-bold text-white">플랫폼 업로드 및 정산 정책 버전 ({{ policyVersion }})</h2>
          <span v-if="savedBanner" class="text-xs font-bold text-emerald-400" role="status">
            ✓ Policy {{ policyVersion }} committed
          </span>
        </div>

        <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3">
          <div>
            <label for="pol-upload" class="text-xs text-neutral-400">Max Multipart Upload Size (GB)</label>
            <input
              id="pol-upload"
              v-model.number="maxUploadSizeGb"
              type="number"
              min="1"
              class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3 py-2 text-xs text-white"
            />
          </div>
          <div>
            <label for="pol-duration" class="text-xs text-neutral-400">Max Single Video Duration (Hours)</label>
            <input
              id="pol-duration"
              v-model.number="maxVideoDurationHours"
              type="number"
              min="1"
              class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3 py-2 text-xs text-white"
            />
          </div>
          <div>
            <label for="pol-share" class="text-xs text-neutral-400">Standard Creator Ad Revenue Share (%)</label>
            <input
              id="pol-share"
              v-model.number="creatorRevenueSharePct"
              type="number"
              min="0"
              max="100"
              class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3 py-2 text-xs text-white"
            />
          </div>
        </div>

        <div class="mt-4 flex justify-end">
          <button
            type="submit"
            :disabled="saving"
            class="rounded-xl bg-red-600 px-5 py-2 text-xs font-bold text-white hover:bg-red-500 disabled:opacity-50"
          >
            {{ saving ? '저장 중…' : 'Commit New Policy Version' }}
          </button>
        </div>
      </form>

      <!-- Emergency Feature Flags & Kill-Switches -->
      <div class="mt-6 rounded-2xl border border-red-500/30 bg-[#131722] p-6">
        <h2 class="text-sm font-bold text-white">긴급 킬스위치 및 피처 플래그 (Emergency Circuit Breakers)</h2>
        <GvpEmptyState
          v-if="flags.length === 0"
          class="mt-4"
          title="등록된 플래그가 없습니다"
        />
        <div v-else class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
          <div
            v-for="flag in flags"
            :key="flag.key"
            class="flex items-center justify-between rounded-xl border border-white/10 bg-[#090B10] p-4 text-xs"
          >
            <div class="min-w-0">
              <p class="font-mono font-bold text-white">{{ flag.key }}</p>
              <p class="mt-0.5 text-neutral-400">{{ flag.description || flag.label }}</p>
            </div>
            <button
              type="button"
              :disabled="togglingKeys.has(flag.key)"
              class="ml-3 shrink-0 rounded-xl px-3.5 py-1.5 font-extrabold transition disabled:opacity-40"
              :class="flag.enabled ? 'bg-emerald-500/20 text-emerald-400' : 'bg-red-600 text-white'"
              @click="toggleFlag(flag)"
            >
              {{ flag.enabled ? 'ENABLED' : 'KILLED (OFF)' }}
            </button>
          </div>
        </div>
      </div>
    </template>

    <!-- Immutable Audit Log Viewer -->
    <div class="mt-6 rounded-2xl border border-white/10 bg-[#131722] p-5">
      <h2 class="text-sm font-bold text-white">불변 감사 로그 원장 (`/v1/admin/audit-logs`)</h2>
      <div v-if="auditPending" class="mt-3 space-y-2" role="status" aria-label="감사 기록 불러오는 중">
        <div v-for="n in 3" :key="n" class="h-10 animate-pulse rounded-xl bg-white/5" />
      </div>
      <div v-else-if="!auditLogs || auditLogs.length === 0" class="mt-3 text-xs text-neutral-400">
        기록된 감사 로그가 없습니다.
      </div>
      <div v-else class="mt-3 space-y-2 text-xs">
        <div
          v-for="item in auditLogs"
          :key="item.id"
          class="flex flex-wrap items-center justify-between gap-2 rounded-xl bg-[#090B10] px-4 py-2.5 font-mono"
        >
          <span class="text-[#3EA6FF]">{{ item.id }} · {{ item.action }}</span>
          <span class="text-neutral-300">Actor: {{ item.actorId }} ({{ item.ip }})</span>
          <span class="text-neutral-500">{{ item.timestamp }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
