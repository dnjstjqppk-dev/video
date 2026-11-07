<script setup lang="ts">
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'

useSeoMeta({
  title: 'Finance — Ledger Review, Batch Settlement Approval & Cost Simulator'
})

const api = useAdminApi()

const stepUpTotpCode = ref('')
const approving = ref(false)
const approvalResult = ref<string | null>(null)
const approvalError = ref<string | null>(null)

// Unit Economics Simulator inputs (client-side planning calculator)
const monthlyUploadsTB = ref(120)
const monthlyEgressTB = ref(4500)
const transcodeMinutes = ref(850000)

const estimatedStorageCostUsd = computed(() => Math.round(monthlyUploadsTB.value * 15))
const estimatedR2CdnCostUsd = computed(() => Math.round(monthlyEgressTB.value * 2.5))
const estimatedTranscodeCostUsd = computed(() => Math.round(transcodeMinutes.value * 0.004))
const totalInfraCostUsd = computed(
  () => estimatedStorageCostUsd.value + estimatedR2CdnCostUsd.value + estimatedTranscodeCostUsd.value
)

const { data: summary, pending: summaryPending, error: summaryError, refresh: refreshSummary } = await useAsyncData(
  'admin-finance-summary',
  async () => await api.getAdminFinanceSummary()
)

const { data: settlements, pending: settlementsPending, error: settlementsError, refresh: refreshSettlements } = await useAsyncData(
  'admin-finance-settlements',
  async () => await api.getAdminFinanceSettlements()
)

function formatKrw(cents: number): string {
  return `₩${Math.round(cents / 100).toLocaleString('ko-KR')}`
}

async function approveSettlementBatch() {
  if (!summary.value) return
  if (stepUpTotpCode.value.trim().length !== 6) {
    approvalError.value = '6자리 TOTP 코드를 입력하세요.'
    return
  }
  approving.value = true
  approvalResult.value = null
  approvalError.value = null
  try {
    const res = await api.approveSettlementBatch(summary.value.batchId, stepUpTotpCode.value.trim())
    approvalResult.value = `승인 완료: ${res.approvalId} · ${res.batchId} · 차변=대변 ${formatKrw(res.debitTotal)}`
    stepUpTotpCode.value = ''
    await refreshSummary()
  } catch (e: unknown) {
    approvalError.value = e instanceof Error ? e.message : '배치 승인에 실패했습니다.'
  } finally {
    approving.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-extrabold text-white">
      재무 검토, 복식부기 대량 정산 승인 및 유닛 이코노믹스 시뮬레이터
    </h1>
    <p class="mt-1 text-xs text-neutral-400">
      Review live ledger balance · Step-up TOTP approval for monthly payouts · Simulate R2/CDN/Transcode costs
    </p>

    <div class="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
      <!-- Double-Entry Batch Payout Approval with Step-Up Re-Authentication -->
      <div class="rounded-2xl border border-white/10 bg-[#131722] p-5">
        <h2 class="text-sm font-bold text-white">월간 복식부기 정산 배치 승인 (Step-Up MFA Required)</h2>
        <div v-if="summaryPending" class="mt-3 h-24 animate-pulse rounded-xl bg-white/5" role="status" aria-label="불러오는 중" />
        <GvpBanner v-else-if="summaryError" tone="danger" class="mt-3">
          원장 요약을 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refreshSummary()">다시 시도</button>
        </GvpBanner>
        <div v-else-if="summary" class="mt-3 rounded-xl border border-white/10 bg-[#090B10] p-4 text-xs">
          <p class="font-bold text-[#3EA6FF]">Batch ID: {{ summary.batchId }}</p>
          <p class="mt-1 text-neutral-300">
            차변 합계 {{ formatKrw(summary.debitTotalCents) }} · 대변 합계 {{ formatKrw(summary.creditTotalCents) }}
            <strong :class="summary.balanced ? 'text-emerald-400' : 'text-red-400'">
              {{ summary.balanced ? '· Balanced ✓' : '· IMBALANCED — 승인 불가' }}
            </strong>
          </p>
          <p class="mt-0.5 text-neutral-400">
            원장 {{ summary.ledgerEntries }}건 · 창작자 미지급 {{ formatKrw(summary.creatorPayableCents) }} · 성공 거래 {{ formatKrw(summary.succeededCents) }}
          </p>
        </div>

        <div class="mt-4 flex items-center gap-2">
          <label for="finance-totp" class="sr-only">6자리 TOTP 코드</label>
          <input
            id="finance-totp"
            v-model="stepUpTotpCode"
            type="text"
            inputmode="numeric"
            maxlength="6"
            placeholder="본인 계정의 6자리 TOTP 코드"
            class="flex-1 rounded-xl border border-white/15 bg-[#090B10] px-3.5 py-2 font-mono text-xs text-white"
          />
          <button
            type="button"
            :disabled="approving || !summary?.balanced"
            class="rounded-xl bg-emerald-500 px-4 py-2 text-xs font-bold text-black hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-40"
            @click="approveSettlementBatch"
          >
            {{ approving ? '검증 중…' : 'Verify & Approve Payouts' }}
          </button>
        </div>

        <GvpBanner v-if="approvalResult" tone="success" class="mt-3">{{ approvalResult }}</GvpBanner>
        <GvpBanner v-if="approvalError" tone="danger" class="mt-3">{{ approvalError }}</GvpBanner>
      </div>

      <!-- Settlement Statements Review -->
      <div class="rounded-2xl border border-white/10 bg-[#131722] p-5">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-bold text-white">정산 명세 검토 (Settlement Statements)</h2>
          <button type="button" class="text-xs font-bold text-[#3EA6FF] hover:underline" @click="refreshSettlements()">새로고침</button>
        </div>
        <div v-if="settlementsPending" class="mt-4 space-y-2" role="status" aria-label="불러오는 중">
          <div v-for="n in 3" :key="n" class="h-16 animate-pulse rounded-xl bg-white/5" />
        </div>
        <GvpBanner v-else-if="settlementsError" tone="danger" class="mt-4">
          정산 명세를 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refreshSettlements()">다시 시도</button>
        </GvpBanner>
        <GvpEmptyState
          v-else-if="!settlements || settlements.length === 0"
          class="mt-4"
          title="정산 명세가 없습니다"
        />
        <div v-else class="mt-4 space-y-3 text-xs">
          <div
            v-for="s in settlements"
            :key="s.id"
            class="rounded-xl border border-white/10 bg-[#090B10] p-3.5"
          >
            <div class="flex items-center justify-between">
              <p class="font-mono font-bold text-[#3EA6FF]">{{ s.id }}</p>
              <span
                class="rounded px-2 py-0.5 text-[10px] font-bold uppercase"
                :class="String(s.status).toLowerCase() === 'paid' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-amber-500/20 text-amber-300'"
              >
                {{ s.status }}
              </span>
            </div>
            <p class="mt-1 text-neutral-300">{{ s.periodStart }} ~ {{ s.periodEnd }} · 총 {{ formatKrw(s.amount) }} · 세금 {{ formatKrw(s.taxWithheld) }} · 실지급 {{ formatKrw(s.netAmount) }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Platform Unit Economics & Infrastructure Cost Simulator -->
    <div class="mt-6 rounded-2xl border border-white/10 bg-[#131722] p-6">
      <h2 class="text-sm font-bold text-white">플랫폼 유닛 이코노믹스 시뮬레이터 (Storage / Transcoding / CDN Cost Model)</h2>
      <p class="mt-1 text-[11px] text-neutral-500">기획용 추정치 계산기입니다. 실제 청구는 클라우드 콘솔을 확인하세요.</p>
      <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3">
        <div>
          <label class="text-xs text-neutral-400">Monthly R2 Master + HLS Storage (TB)</label>
          <input
            v-model.number="monthlyUploadsTB"
            type="number"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3 py-2 text-xs text-white"
          />
        </div>
        <div>
          <label class="text-xs text-neutral-400">Monthly Global Edge Delivery (TB)</label>
          <input
            v-model.number="monthlyEgressTB"
            type="number"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3 py-2 text-xs text-white"
          />
        </div>
        <div>
          <label class="text-xs text-neutral-400">Monthly FFmpeg ABR Transcode Minutes</label>
          <input
            v-model.number="transcodeMinutes"
            type="number"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3 py-2 text-xs text-white"
          />
        </div>
      </div>

      <div class="mt-5 grid grid-cols-2 gap-4 rounded-xl bg-[#090B10] p-4 text-xs sm:grid-cols-4">
        <div>
          <p class="text-neutral-400">R2 Storage Cost</p>
          <p class="mt-1 font-mono text-base font-bold text-white">${{ estimatedStorageCostUsd.toLocaleString() }}/mo</p>
        </div>
        <div>
          <p class="text-neutral-400">CDN Delivery (Zero R2 Egress)</p>
          <p class="mt-1 font-mono text-base font-bold text-[#3EA6FF]">${{ estimatedR2CdnCostUsd.toLocaleString() }}/mo</p>
        </div>
        <div>
          <p class="text-neutral-400">FFmpeg Worker Pool</p>
          <p class="mt-1 font-mono text-base font-bold text-amber-300">${{ estimatedTranscodeCostUsd.toLocaleString() }}/mo</p>
        </div>
        <div>
          <p class="text-neutral-400">Estimated Total Infra Spend</p>
          <p class="mt-1 font-mono text-base font-extrabold text-emerald-400">${{ totalInfraCostUsd.toLocaleString() }}/mo</p>
        </div>
      </div>
    </div>
  </div>
</template>
