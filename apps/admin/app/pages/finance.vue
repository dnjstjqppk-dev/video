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
  <div class="space-y-6">
    <div class="border-b border-[var(--gvp-border)] pb-4">
      <h1 class="text-2xl font-bold tracking-tight text-[var(--gvp-text)]">
        재무 검토, 복식부기 대량 정산 승인 및 유닛 이코노믹스 시뮬레이터
      </h1>
      <p class="mt-1 text-sm text-[var(--gvp-text-secondary)]">
        Review live ledger balance · Step-up TOTP approval for monthly payouts · Simulate R2/CDN/Transcode costs
      </p>
    </div>

    <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <!-- Double-Entry Batch Payout Approval with Step-Up Re-Authentication -->
      <div class="overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
        <div class="border-b border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-5 py-3.5">
          <h2 class="text-sm font-semibold text-[var(--gvp-text)]">
            월간 복식부기 정산 배치 승인 (Step-Up MFA Required)
          </h2>
        </div>

        <div class="p-5">
          <div v-if="summaryPending" class="h-24 animate-pulse rounded-lg bg-[var(--gvp-elevated)]" role="status" aria-label="불러오는 중" />
          <GvpBanner v-else-if="summaryError" tone="danger">
            원장 요약을 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refreshSummary()">다시 시도</button>
          </GvpBanner>
          <div v-else-if="summary" class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-4 text-xs">
            <p class="font-mono font-semibold text-[var(--gvp-text)] tabular-nums">Batch ID: {{ summary.batchId }}</p>
            <p class="mt-1.5 text-[var(--gvp-text-secondary)] tabular-nums">
              차변 합계 {{ formatKrw(summary.debitTotalCents) }} · 대변 합계 {{ formatKrw(summary.creditTotalCents) }}
              <strong :class="summary.balanced ? 'text-[var(--gvp-success)]' : 'text-[var(--gvp-live)]'">
                {{ summary.balanced ? '· Balanced' : '· IMBALANCED — 승인 불가' }}
              </strong>
            </p>
            <p class="mt-1 text-[var(--gvp-text-secondary)] tabular-nums">
              원장 {{ summary.ledgerEntries }}건 · 창작자 미지급 {{ formatKrw(summary.creatorPayableCents) }} · 성공 거래 {{ formatKrw(summary.succeededCents) }}
            </p>
          </div>

          <div class="mt-4 flex flex-wrap items-center gap-2">
            <label for="finance-totp" class="sr-only">6자리 TOTP 코드</label>
            <input
              id="finance-totp"
              v-model="stepUpTotpCode"
              type="text"
              inputmode="numeric"
              maxlength="6"
              placeholder="본인 계정의 6자리 TOTP 코드"
              class="h-9 flex-1 rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3.5 font-mono text-sm text-[var(--gvp-text)] tabular-nums placeholder:text-[var(--gvp-text-muted)] focus:border-[var(--gvp-focus)] focus:bg-[var(--gvp-surface)] outline-none transition-colors"
            />
            <button
              type="button"
              :disabled="approving || !summary?.balanced"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors disabled:cursor-not-allowed disabled:opacity-40"
              @click="approveSettlementBatch"
            >
              {{ approving ? '검증 중…' : 'Verify & Approve Payouts' }}
            </button>
          </div>

          <GvpBanner v-if="approvalResult" tone="success" class="mt-3">{{ approvalResult }}</GvpBanner>
          <GvpBanner v-if="approvalError" tone="danger" class="mt-3">{{ approvalError }}</GvpBanner>
        </div>
      </div>

      <!-- Settlement Statements Review -->
      <div class="overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
        <div class="flex items-center justify-between border-b border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-5 py-3">
          <h2 class="text-sm font-semibold text-[var(--gvp-text)]">정산 명세 검토 (Settlement Statements)</h2>
          <button
            type="button"
            class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
            @click="refreshSettlements()"
          >
            새로고침
          </button>
        </div>

        <div class="p-5">
          <div v-if="settlementsPending" class="space-y-2" role="status" aria-label="불러오는 중">
            <div v-for="n in 3" :key="n" class="h-16 animate-pulse rounded-lg bg-[var(--gvp-elevated)]" />
          </div>
          <GvpBanner v-else-if="settlementsError" tone="danger">
            정산 명세를 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refreshSettlements()">다시 시도</button>
          </GvpBanner>
          <GvpEmptyState
            v-else-if="!settlements || settlements.length === 0"
            title="정산 명세가 없습니다"
          />
          <div v-else class="space-y-3 text-xs">
            <div
              v-for="s in settlements"
              :key="s.id"
              class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-3.5"
            >
              <div class="flex items-center justify-between">
                <p class="font-mono font-semibold text-[var(--gvp-text)] tabular-nums">{{ s.id }}</p>
                <span
                  class="rounded-md px-2 py-0.5 text-xs font-semibold uppercase"
                  :class="String(s.status).toLowerCase() === 'paid' ? 'border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] text-[var(--gvp-success)]' : 'border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] text-[var(--gvp-warning)]'"
                >
                  {{ s.status }}
                </span>
              </div>
              <p class="mt-1.5 text-[var(--gvp-text-secondary)] tabular-nums">
                {{ s.periodStart }} ~ {{ s.periodEnd }} · 총 {{ formatKrw(s.amount) }} · 세금 {{ formatKrw(s.taxWithheld) }} · 실지급 {{ formatKrw(s.netAmount) }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Platform Unit Economics & Infrastructure Cost Simulator -->
    <div class="overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
      <div class="border-b border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-5 py-3.5">
        <h2 class="text-sm font-semibold text-[var(--gvp-text)]">
          플랫폼 유닛 이코노믹스 시뮬레이터 (Storage / Transcoding / CDN Cost Model)
        </h2>
        <p class="mt-0.5 text-xs text-[var(--gvp-text-secondary)]">기획용 추정치 계산기입니다. 실제 청구는 클라우드 콘솔을 확인하세요.</p>
      </div>

      <div class="p-5">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
          <div>
            <label class="block text-xs font-medium text-[var(--gvp-text-secondary)]">Monthly R2 Master + HLS Storage (TB)</label>
            <input
              v-model.number="monthlyUploadsTB"
              type="number"
              class="mt-1.5 h-9 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3 text-sm text-[var(--gvp-text)] tabular-nums focus:border-[var(--gvp-focus)] focus:bg-[var(--gvp-surface)] outline-none transition-colors"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-[var(--gvp-text-secondary)]">Monthly Global Edge Delivery (TB)</label>
            <input
              v-model.number="monthlyEgressTB"
              type="number"
              class="mt-1.5 h-9 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3 text-sm text-[var(--gvp-text)] tabular-nums focus:border-[var(--gvp-focus)] focus:bg-[var(--gvp-surface)] outline-none transition-colors"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-[var(--gvp-text-secondary)]">Monthly FFmpeg ABR Transcode Minutes</label>
            <input
              v-model.number="transcodeMinutes"
              type="number"
              class="mt-1.5 h-9 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3 text-sm text-[var(--gvp-text)] tabular-nums focus:border-[var(--gvp-focus)] focus:bg-[var(--gvp-surface)] outline-none transition-colors"
            />
          </div>
        </div>

        <div class="mt-5 grid grid-cols-2 gap-4 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-4 text-xs sm:grid-cols-4">
          <div>
            <p class="font-medium text-[var(--gvp-text-secondary)]">R2 Storage Cost</p>
            <p class="mt-1 font-mono text-base font-bold text-[var(--gvp-text)] tabular-nums">${{ estimatedStorageCostUsd.toLocaleString() }}/mo</p>
          </div>
          <div>
            <p class="font-medium text-[var(--gvp-text-secondary)]">CDN Delivery (Zero R2 Egress)</p>
            <p class="mt-1 font-mono text-base font-bold text-[var(--gvp-text)] tabular-nums">${{ estimatedR2CdnCostUsd.toLocaleString() }}/mo</p>
          </div>
          <div>
            <p class="font-medium text-[var(--gvp-text-secondary)]">FFmpeg Worker Pool</p>
            <p class="mt-1 font-mono text-base font-bold text-[var(--gvp-warning)] tabular-nums">${{ estimatedTranscodeCostUsd.toLocaleString() }}/mo</p>
          </div>
          <div>
            <p class="font-medium text-[var(--gvp-text-secondary)]">Estimated Total Infra Spend</p>
            <p class="mt-1 font-mono text-base font-bold text-[var(--gvp-success)] tabular-nums">${{ totalInfraCostUsd.toLocaleString() }}/mo</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
