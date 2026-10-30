<script setup lang="ts">
useSeoMeta({
  title: 'Channel Copyright & Claims — Creator Studio'
})

interface CopyrightClaimItem {
  id: string
  videoTitle: string
  videoId: string
  claimant: string
  matchedAsset: string
  timeRange: string
  policy: 'MONETIZE' | 'TRACK' | 'BLOCK'
  status: 'CLAIMED' | 'DISPUTED' | 'RELEASED'
  claimDate: string
}

const strikes = ref(0)
const maxStrikes = 3

const claims = ref<CopyrightClaimItem[]>([
  {
    id: 'claim_cjenm_001',
    videoTitle: 'Building a 10M Concurrent Viewer Global Video Platform with Go 1.24',
    videoId: 'vid_arch_2026',
    claimant: 'CJ ENM Music Publishing',
    matchedAsset: 'Synthwave Skyline (Original BGM)',
    timeRange: '03:15 - 04:30',
    policy: 'MONETIZE',
    status: 'CLAIMED',
    claimDate: '2026-10-25'
  }
])

const disputeModalOpen = ref(false)
const selectedClaim = ref<CopyrightClaimItem | null>(null)
const disputeReason = ref('fair_use')
const disputeExplanation = ref('')
const disputePerjuryAgreed = ref(false)
const toastMessage = ref('')

function openDisputeModal(c: CopyrightClaimItem) {
  selectedClaim.value = c
  disputeReason.value = 'fair_use'
  disputeExplanation.value = ''
  disputePerjuryAgreed.value = false
  disputeModalOpen.value = true
}

function submitDispute() {
  if (!selectedClaim.value || !disputePerjuryAgreed.value) return
  selectedClaim.value.status = 'DISPUTED'
  disputeModalOpen.value = false
  toastMessage.value = '이의 신청이 권리자(CJ ENM)에게 공식 접수되었습니다. 30일 이내에 답변이 없으면 클레임이 자동 해제됩니다.'
  setTimeout(() => {
    toastMessage.value = ''
  }, 6000)
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-extrabold text-white">채널 저작권 (Copyright & Rights)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          Content ID 클레임 상태 확인, 저작권 침해 경고 및 30일 이의신청(Dispute) 관리
        </p>
      </div>

      <!-- Strike Status Pill -->
      <div class="flex items-center gap-2 rounded-2xl border border-white/10 bg-[#161922] px-4 py-2">
        <span class="text-xs text-neutral-400">저작권 침해 경고:</span>
        <span class="font-extrabold" :class="strikes === 0 ? 'text-emerald-400' : 'text-rose-500'">
          {{ strikes }} / {{ maxStrikes }}
        </span>
        <span class="text-[10px] text-neutral-500">(계정 상태 양호)</span>
      </div>
    </div>

    <!-- Toast Notification -->
    <div
      v-if="toastMessage"
      class="rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-4 text-xs font-semibold text-emerald-300"
    >
      {{ toastMessage }}
    </div>

    <!-- Active Claims Section -->
    <div class="rounded-2xl border border-white/10 bg-[#161922] p-6">
      <h2 class="text-base font-bold text-white">내 동영상에 제기된 Content ID 클레임</h2>
      <p class="mt-1 text-xs text-neutral-400">
        클레임이 제기된 구간의 수익은 분쟁 해결 전까지 에스크로 계좌에 안전하게 유보됩니다.
      </p>

      <div class="mt-6 divide-y divide-white/5">
        <div
          v-for="c in claims"
          :key="c.id"
          class="flex flex-col gap-4 py-4 md:flex-row md:items-center md:justify-between"
        >
          <div>
            <div class="text-sm font-semibold text-white">{{ c.videoTitle }}</div>
            <div class="mt-1 flex flex-wrap items-center gap-3 text-xs text-neutral-400">
              <span>권리자: <strong class="text-neutral-200">{{ c.claimant }}</strong></span>
              <span>·</span>
              <span>일치된 음원: <span class="italic text-neutral-200">{{ c.matchedAsset }}</span> (구간: {{ c.timeRange }})</span>
              <span>·</span>
              <span>접수일: {{ c.claimDate }}</span>
            </div>
          </div>

          <div class="flex items-center gap-3">
            <span
              class="rounded-lg px-2.5 py-1 text-xs font-bold uppercase"
              :class="{
                'bg-amber-500/20 text-amber-300': c.status === 'CLAIMED',
                'bg-sky-500/20 text-sky-300': c.status === 'DISPUTED',
                'bg-emerald-500/20 text-emerald-300': c.status === 'RELEASED'
              }"
            >
              {{ c.status === 'CLAIMED' ? '수익 분배 중 (MONETIZE)' : (c.status === 'DISPUTED' ? '이의 제기 심사 중' : '클레임 해제됨') }}
            </span>

            <button
              v-if="c.status === 'CLAIMED'"
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
    <div
      v-if="disputeModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-4 backdrop-blur-sm"
    >
      <div class="w-full max-w-lg rounded-2xl border border-white/15 bg-[#1B1F2B] p-6 text-white shadow-2xl">
        <h3 class="text-lg font-extrabold">Content ID 클레임 공식 이의 제기</h3>
        <p class="mt-1 text-xs text-neutral-400">
          허위 또는 악의적인 이의 제기는 계정 해지 및 법적 책임이 따를 수 있습니다.
        </p>

        <div class="mt-4 space-y-4">
          <div>
            <label class="block text-xs font-semibold text-neutral-300">이의 제기 사유</label>
            <select
              v-model="disputeReason"
              class="mt-1 w-full rounded-xl border border-white/15 bg-[#12151E] p-2.5 text-xs text-white outline-none"
            >
              <option value="fair_use">공정 이용 (비평, 교육, 패러디 등 저작권법상 인정된 이용)</option>
              <option value="public_domain">공유 저작물 / 퍼블릭 도메인 음원</option>
              <option value="license">원저작자로부터 적법한 라이선스 계약을 체결함</option>
              <option value="original">본인이 직접 작곡 및 제작한 순수 창작물임</option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-semibold text-neutral-300">상세 소명 내용</label>
            <textarea
              v-model="disputeExplanation"
              rows="4"
              class="mt-1 w-full rounded-xl border border-white/15 bg-[#12151E] p-3 text-xs text-white outline-none focus:border-[#FF0033]"
              placeholder="해당 저작물을 적법하게 사용한 근거 및 라이선스 번호, 사용 목적을 구체적으로 기재해주세요."
            ></textarea>
          </div>

          <label class="flex items-start gap-2 text-xs text-neutral-300">
            <input v-model="disputePerjuryAgreed" type="checkbox" class="mt-0.5 rounded border-white/20 bg-black/40 text-[#FF0033]" />
            <span>위증 시 법적 처벌(DMCA 512(f) 및 저작권법)을 감수하며, 본 이의 신청 내용이 진실임을 서약합니다.</span>
          </label>
        </div>

        <div class="mt-6 flex justify-end gap-3">
          <button
            type="button"
            class="rounded-xl border border-white/15 px-4 py-2 text-xs font-bold text-neutral-300 hover:bg-white/5"
            @click="disputeModalOpen = false"
          >
            취소
          </button>
          <button
            type="button"
            class="rounded-xl bg-[#FF0033] px-5 py-2 text-xs font-bold text-white transition hover:bg-[#D9002B] disabled:opacity-50"
            :disabled="!disputePerjuryAgreed || !disputeExplanation.trim()"
            @click="submitDispute"
          >
            이의 제기 제출
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

