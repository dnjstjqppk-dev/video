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
    case 'disputed': return 'bg-sky-500/20 text-sky-300'
    case 'released': return 'bg-emerald-500/20 text-emerald-300'
    case 'upheld':
    case 'takedown': return 'bg-red-500/20 text-red-300'
    default: return 'bg-amber-500/20 text-amber-300'
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
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">채널 저작권 (Copyright & Rights)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          Content ID 클레임 상태 확인 및 이의신청(Dispute) 관리
        </p>
      </div>

      <!-- Claim Status Summary -->
      <div class="flex items-center gap-2 rounded-2xl border border-white/10 bg-[#161922] px-4 py-2 text-xs" role="status">
        <span class="text-neutral-400">활성 <strong class="text-amber-300">{{ activeCount }}</strong></span>
        <span class="text-neutral-600">·</span>
        <span class="text-neutral-400">심사 중 <strong class="text-sky-300">{{ disputedCount }}</strong></span>
        <span class="text-neutral-600">·</span>
        <span class="text-neutral-400">해제 <strong class="text-emerald-300">{{ releasedCount }}</strong></span>
      </div>
    </div>

    <!-- Toast Notification -->
    <div
      v-if="toastMessage"
      class="rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-4 text-xs font-semibold text-emerald-300"
      role="status"
    >
      {{ toastMessage }}
    </div>

    <GvpBanner v-if="actionError" tone="danger" dismissible @dismiss="actionError = null">
      {{ actionError }}
    </GvpBanner>

    <!-- Active Claims Section -->
    <div class="rounded-2xl border border-white/10 bg-[#161922] p-6">
      <h2 class="text-base font-bold text-white">내 동영상에 제기된 Content ID 클레임</h2>
      <p class="mt-1 text-xs text-neutral-400">
        클레임이 제기된 구간의 수익은 분쟁 해결 전까지 유보될 수 있습니다.
      </p>

      <div v-if="pending" class="mt-6 space-y-3" role="status" aria-label="불러오는 중">
        <div v-for="n in 3" :key="n" class="h-20 animate-pulse rounded-xl bg-white/5" />
      </div>

      <GvpBanner v-else-if="error" tone="danger" class="mt-6">
        클레임을 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refresh()">다시 시도</button>
      </GvpBanner>

      <GvpEmptyState
        v-else-if="claims.length === 0"
        class="mt-6"
        title="제기된 클레임이 없습니다"
        description="Content ID와 일치하는 구간이 발견되면 이 목록에 표시됩니다."
      />

      <div v-else class="mt-6 divide-y divide-white/5">
        <div
          v-for="c in claims"
          :key="c.id"
          class="flex flex-col gap-4 py-4 md:flex-row md:items-center md:justify-between"
        >
          <div class="min-w-0">
            <div class="font-mono text-sm font-semibold text-white">{{ c.claimedVideoId }}</div>
            <div class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-neutral-400">
              <span>권리자: <strong class="text-neutral-200">{{ c.ownerName }}</strong></span>
              <span aria-hidden="true">·</span>
              <span>구간: {{ formatRange(c) }}</span>
              <span aria-hidden="true">·</span>
              <span>접수일: {{ formatDate(c.createdAt) }}</span>
            </div>
          </div>

          <div class="flex shrink-0 items-center gap-3">
            <span
              class="rounded-lg px-2.5 py-1 text-xs font-bold uppercase"
              :class="statusTone(c.status)"
            >
              {{ statusLabel(c) }}
            </span>

            <button
              v-if="c.status === 'active' || c.status === 'pending'"
              type="button"
              class="rounded-xl border border-white/15 bg-white/5 px-3 py-1.5 text-xs font-bold text-white transition hover:bg-white/15"
              @click="openDisputeModal(c)"
            >
              이의 제기 (Dispute)
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Dispute Modal -->
    <GvpModal :model-value="disputeTarget !== null" title="Content ID 클레임 공식 이의 제기" @update:model-value="(v: boolean) => { if (!v) disputeTarget = null }">
      <p class="text-xs text-neutral-400">
        허위 또는 악의적인 이의 제기는 계정 해지 및 법적 책임이 따를 수 있습니다.
      </p>

      <div class="mt-4 space-y-4">
        <div>
          <label for="dispute-reason" class="block text-xs font-semibold text-neutral-300">이의 제기 사유</label>
          <select
            id="dispute-reason"
            v-model="disputeReason"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#12151E] p-2.5 text-xs text-white outline-none"
          >
            <option value="fair_use">공정 이용 (비평, 교육, 패러디 등 저작권법상 인정된 이용)</option>
            <option value="public_domain">공유 저작물 / 퍼블릭 도메인 음원</option>
            <option value="license_held">원저작자로부터 적법한 라이선스 계약을 체결함</option>
            <option value="original_content">본인이 직접 작곡 및 제작한 순수 창작물임</option>
          </select>
        </div>

        <div>
          <label for="dispute-explain" class="block text-xs font-semibold text-neutral-300">상세 소명 내용</label>
          <textarea
            id="dispute-explain"
            v-model="disputeExplanation"
            rows="4"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#12151E] p-3 text-xs text-white outline-none focus:border-[#FF0033]"
            placeholder="해당 저작물을 적법하게 사용한 근거 및 라이선스 번호, 사용 목적을 구체적으로 기재해주세요."
          />
        </div>

        <label class="flex items-start gap-2 text-xs text-neutral-300">
          <input v-model="disputePerjuryAgreed" type="checkbox" class="mt-0.5 rounded border-white/20 bg-black/40 text-[#FF0033]" />
          <span>위증 시 법적 처벌을 감수하며, 본 이의 신청 내용이 진실임을 서약합니다.</span>
        </label>
      </div>

      <template #footer>
        <button
          type="button"
          class="rounded-xl border border-white/15 px-4 py-2 text-xs font-bold text-neutral-300 hover:bg-white/5"
          @click="disputeTarget = null"
        >
          취소
        </button>
        <button
          type="button"
          class="rounded-xl bg-[#FF0033] px-5 py-2 text-xs font-bold text-white transition hover:bg-[#D9002B] disabled:opacity-50"
          :disabled="!disputePerjuryAgreed || !disputeExplanation.trim() || disputing"
          @click="submitDispute"
        >
          {{ disputing ? '제출 중…' : '이의 제기 제출' }}
        </button>
      </template>
    </GvpModal>
  </div>
</template>
