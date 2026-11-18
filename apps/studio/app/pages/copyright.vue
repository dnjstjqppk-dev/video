<script setup lang="ts">
import type { CopyrightClaim } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import GvpModal from '@gvp/ui/src/components/GvpModal.vue'
import { useStudioApi } from '~/composables/useStudioApi'

useSeoMeta({
  title: 'Channel Copyright & Claims — Creator Studio'
})

const api = useStudioApi()

const { data, pending, error, refresh } = await useAsyncData('studio-claims', async () => {
  return await api.getStudioClaims()
})

const claims = computed(() => data.value ?? [])
const activeCount = computed(() => claims.value.filter(c => c.status === 'active' || c.status === 'pending').length)
const disputedCount = computed(() => claims.value.filter(c => c.status === 'disputed').length)
const releasedCount = computed(() => claims.value.filter(c => c.status === 'released').length)

const disputeTarget = ref<CopyrightClaim | null>(null)
const disputeReason = ref('fair_use')
const disputeExplanation = ref('')
const disputePerjuryAgreed = ref(false)
const disputing = ref(false)
const actionError = ref<string | null>(null)
const toastMessage = ref('')

function openDisputeModal(c: CopyrightClaim) {
  disputeTarget.value = c
  disputeReason.value = 'fair_use'
  disputeExplanation.value = ''
  disputePerjuryAgreed.value = false
  actionError.value = null
}

function formatRange(c: CopyrightClaim): string {
  const fmt = (s: number) => {
    const m = Math.floor(s / 60)
    const sec = s % 60
    return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
  }
  return `${fmt(c.matchStartSeconds)} - ${fmt(c.matchEndSeconds)}`
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? '-' : d.toLocaleDateString('ko-KR')
}

function statusTone(status: string): string {
  switch (status) {
    case 'disputed': return 'border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] text-[var(--gvp-brand-hover)]'
    case 'released': return 'border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] text-[var(--gvp-success)]'
    case 'upheld':
    case 'takedown': return 'border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] text-[var(--gvp-live)]'
    default: return 'border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] text-[var(--gvp-warning)]'
  }
}

function statusLabel(c: CopyrightClaim): string {
  switch (c.status) {
    case 'disputed': return '이의 제기 심사 중'
    case 'released': return '클레임 해제됨'
    case 'upheld': return '클레임 유지됨'
    case 'takedown': return '삭제 조치됨'
    default: return `클레임 활성 (${String(c.claimAction).toUpperCase()})`
  }
}

async function submitDispute() {
  if (!disputeTarget.value || !disputePerjuryAgreed.value || !disputeExplanation.value.trim()) return
  disputing.value = true
  actionError.value = null
  try {
    const updated = await api.disputeCopyrightClaim(
      disputeTarget.value.id,
      disputeReason.value,
      disputeExplanation.value.trim()
    )
    disputeTarget.value.status = updated.status
    disputeTarget.value.disputeReason = updated.disputeReason
    disputeTarget.value.disputedAt = updated.disputedAt
    disputeTarget.value = null
    toastMessage.value = '이의 신청이 접수되었습니다. 권리자 답변은 최대 30일이 소요될 수 있습니다.'
    setTimeout(() => { toastMessage.value = '' }, 6000)
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '이의 신청에 실패했습니다.'
  } finally {
    disputing.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-[var(--gvp-border)] pb-5">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">채널 저작권 (Copyright &amp; Rights)</h1>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          Content ID 클레임 상태 확인 및 이의신청(Dispute) 관리
        </p>
      </div>

      <!-- Claim Status Summary -->
      <div class="flex items-center gap-3 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-4 py-2 text-xs tabular-nums" role="status">
        <span class="text-[var(--gvp-text-secondary)]">활성 <strong class="font-semibold text-[var(--gvp-warning)]">{{ activeCount }}</strong></span>
        <span class="text-[var(--gvp-text-muted)]">·</span>
        <span class="text-[var(--gvp-text-secondary)]">심사 중 <strong class="font-semibold text-[var(--gvp-brand)]">{{ disputedCount }}</strong></span>
        <span class="text-[var(--gvp-text-muted)]">·</span>
        <span class="text-[var(--gvp-text-secondary)]">해제 <strong class="font-semibold text-[var(--gvp-success)]">{{ releasedCount }}</strong></span>
      </div>
    </div>

    <!-- Toast Notification -->
    <div
      v-if="toastMessage"
      class="rounded-md border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] p-4 text-xs font-medium text-[var(--gvp-success)]"
      role="status"
    >
      {{ toastMessage }}
    </div>

    <GvpBanner v-if="actionError" tone="danger" dismissible @dismiss="actionError = null">
      {{ actionError }}
    </GvpBanner>

    <!-- Active Claims Section -->
    <div class="overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
      <div class="border-b border-[var(--gvp-border)] px-6 py-4">
        <h2 class="text-sm font-semibold text-[var(--gvp-text)]">내 동영상에 제기된 Content ID 클레임</h2>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          클레임이 제기된 구간의 수익은 분쟁 해결 전까지 유보될 수 있습니다.
        </p>
      </div>

      <div v-if="pending" class="p-6 space-y-3" role="status" aria-label="불러오는 중">
        <div v-for="n in 3" :key="n" class="h-16 animate-pulse rounded-lg bg-[var(--gvp-elevated)]" />
      </div>

      <div v-else-if="error" class="p-6">
        <GvpBanner tone="danger">
          클레임을 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refresh()">다시 시도</button>
        </GvpBanner>
      </div>

      <div v-else-if="claims.length === 0" class="p-6">
        <GvpEmptyState
          title="제기된 클레임이 없습니다"
          description="Content ID와 일치하는 구간이 발견되면 이 목록에 표시됩니다."
        />
      </div>

      <table v-else class="w-full text-left">
        <thead class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider">
          <tr>
            <th scope="col" class="py-2.5 px-4">Video ID</th>
            <th scope="col" class="py-2.5 px-4">Claimant</th>
            <th scope="col" class="py-2.5 px-4">Segment</th>
            <th scope="col" class="py-2.5 px-4">Date</th>
            <th scope="col" class="py-2.5 px-4">Status</th>
            <th scope="col" class="py-2.5 px-4 text-right">Action</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--gvp-border)]">
          <tr
            v-for="c in claims"
            :key="c.id"
            class="hover:bg-[var(--gvp-canvas)] transition-colors"
          >
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums font-mono font-medium">
              {{ c.claimedVideoId }}
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)]">
              {{ c.ownerName }}
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text-secondary)] border-b border-[var(--gvp-border)] tabular-nums font-mono">
              {{ formatRange(c) }}
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text-secondary)] border-b border-[var(--gvp-border)] tabular-nums">
              {{ formatDate(c.createdAt) }}
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
              <span
                class="inline-flex items-center rounded-md border px-2.5 py-0.5 text-xs font-medium uppercase"
                :class="statusTone(c.status)"
              >
                {{ statusLabel(c) }}
              </span>
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums text-right">
              <button
                v-if="c.status === 'active' || c.status === 'pending'"
                type="button"
                class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
                @click="openDisputeModal(c)"
              >
                이의 제기 (Dispute)
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Dispute Modal -->
    <GvpModal :model-value="disputeTarget !== null" title="Content ID 클레임 공식 이의 제기" @update:model-value="(v: boolean) => { if (!v) disputeTarget = null }">
      <p class="text-xs text-[var(--gvp-text-secondary)]">
        허위 또는 악의적인 이의 제기는 계정 해지 및 법적 책임이 따를 수 있습니다.
      </p>

      <div class="mt-4 space-y-4">
        <div>
          <label for="dispute-reason" class="block text-xs font-medium text-[var(--gvp-text)]">이의 제기 사유</label>
          <select
            id="dispute-reason"
            v-model="disputeReason"
            class="mt-1 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] p-2.5 text-xs text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-focus)]"
          >
            <option value="fair_use">공정 이용 (비평, 교육, 패러디 등 저작권법상 인정된 이용)</option>
            <option value="public_domain">공유 저작물 / 퍼블릭 도메인 음원</option>
            <option value="license_held">원저작자로부터 적법한 라이선스 계약을 체결함</option>
            <option value="original_content">본인이 직접 작곡 및 제작한 순수 창작물임</option>
          </select>
        </div>

        <div>
          <label for="dispute-explain" class="block text-xs font-medium text-[var(--gvp-text)]">상세 소명 내용</label>
          <textarea
            id="dispute-explain"
            v-model="disputeExplanation"
            rows="4"
            class="mt-1 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] p-3 text-xs text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-focus)]"
            placeholder="해당 저작물을 적법하게 사용한 근거 및 라이선스 번호, 사용 목적을 구체적으로 기재해주세요."
          />
        </div>

        <label class="flex items-start gap-2 text-xs text-[var(--gvp-text)]">
          <input v-model="disputePerjuryAgreed" type="checkbox" class="mt-0.5 rounded-md border-[var(--gvp-border-strong)] accent-[var(--gvp-brand)]" />
          <span>위증 시 법적 처벌을 감수하며, 본 이의 신청 내용이 진실임을 서약합니다.</span>
        </label>
      </div>

      <template #footer>
        <button
          type="button"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
          @click="disputeTarget = null"
        >
          취소
        </button>
        <button
          type="button"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors disabled:opacity-50"
          :disabled="!disputePerjuryAgreed || !disputeExplanation.trim() || disputing"
          @click="submitDispute"
        >
          {{ disputing ? '제출 중…' : '이의 제기 제출' }}
        </button>
      </template>
    </GvpModal>
  </div>
</template>
