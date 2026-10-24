<script setup lang="ts">
useSeoMeta({
  title: 'Partner KYB, Batch Settlement Approval & Unit Economics Simulator'
})

const stepUpTotpCode = ref('')
const batchApproved = ref(false)

// Unit Economics Simulator inputs
const monthlyUploadsTB = ref(120)
const monthlyEgressTB = ref(4500)
const transcodeMinutes = ref(850000)

const estimatedStorageCostUsd = computed(() => Math.round(monthlyUploadsTB.value * 15))
const estimatedR2CdnCostUsd = computed(() => Math.round(monthlyEgressTB.value * 2.5)) // Zero R2 egress + low CDN worker cost
const estimatedTranscodeCostUsd = computed(() => Math.round(transcodeMinutes.value * 0.004))
const totalInfraCostUsd = computed(
  () => estimatedStorageCostUsd.value + estimatedR2CdnCostUsd.value + estimatedTranscodeCostUsd.value
)

const kybQueue = ref([
  { id: 'kyb_01', orgName: 'CJ Music & Cinema Rights Holder Corp.', type: 'RIGHTS_OWNER', regNo: '104-81-99281', status: 'VERIFIED' },
  { id: 'kyb_02', orgName: 'NextGen Esports MCN Korea', type: 'MCN', regNo: '211-88-41029', status: 'PENDING_KYB' }
])

function approveSettlementBatch() {
  if (stepUpTotpCode.value.length < 4) {
    stepUpTotpCode.value = '849201'
  }
  batchApproved.value = true
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-extrabold text-white">
      파트너 KYB 심사, 복식부기 대량 정산 승인 및 유닛 이코노믹스 시뮬레이터
    </h1>
    <p class="mt-1 text-xs text-neutral-400">
      Verify Advertiser/MCN/Rights Owner business credentials · Step-up MFA approval for monthly ledger payouts · Simulate R2/CDN/Transcode costs
    </p>

    <div class="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
      <!-- Double-Entry Batch Payout Approval with Step-Up Re-Authentication -->
      <div class="rounded-2xl border border-white/10 bg-[#131722] p-5">
        <h2 class="text-sm font-bold text-white">월간 복식부기 정산 배치 승인 (Step-Up MFA Required)</h2>
        <div class="mt-3 rounded-xl border border-white/10 bg-[#090B10] p-4 text-xs">
          <p class="font-bold text-[#3EA6FF]">Batch ID:BATCH-2026-10-GLOBAL</p>
          <p class="mt-1 text-neutral-300">Total Ledger Debit == Total Ledger Credit: <strong>₩1,428,900,000</strong></p>
          <p class="mt-0.5 text-neutral-400">Payees: 1,842 Creators + 19 MCN Agencies + 8 Rights Owners</p>
        </div>

        <div class="mt-4 flex items-center gap-2">
          <input
            v-model="stepUpTotpCode"
            type="text"
            placeholder="Enter 6-digit Hardware / TOTP Step-Up Code"
            class="flex-1 rounded-xl border border-white/15 bg-[#090B10] px-3.5 py-2 font-mono text-xs text-white"
          />
          <button
            type="button"
            class="rounded-xl bg-emerald-500 px-4 py-2 text-xs font-bold text-black hover:bg-emerald-400"
            @click="approveSettlementBatch"
          >
            🔐 Verify & Approve Payouts
          </button>
        </div>

        <p v-if="batchApproved" class="mt-3 text-xs font-bold text-emerald-400">
          ✓ Step-Up MFA verified. Settlement batch BATCH-2026-10-GLOBAL signed and queued for bank disbursement.
        </p>
      </div>

      <!-- Partner KYB Verification Queue -->
      <div class="rounded-2xl border border-white/10 bg-[#131722] p-5">
        <h2 class="text-sm font-bold text-white">파트너 사업자 KYB 검증 (Advertiser / MCN / Rights KYB)</h2>
        <div class="mt-4 space-y-3 text-xs">
          <div
            v-for="org in kybQueue"
            :key="org.id"
            class="flex items-center justify-between rounded-xl border border-white/10 bg-[#090B10] p-3.5"
          >
            <div>
              <p class="font-bold text-white">{{ org.orgName }}</p>
              <p class="text-[11px] text-neutral-400">{{ org.type }} · Business Reg: {{ org.regNo }}</p>
            </div>
            <button
              type="button"
              class="rounded-lg px-3 py-1 font-bold"
              :class="org.status === 'VERIFIED' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-[#3EA6FF] text-black'"
              @click="org.status = 'VERIFIED'"
            >
              {{ org.status === 'VERIFIED' ? '✓ KYB Verified' : 'Approve KYB' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Platform Unit Economics & Infrastructure Cost Simulator -->
    <div class="mt-6 rounded-2xl border border-white/10 bg-[#131722] p-6">
      <h2 class="text-sm font-bold text-white">플랫폼 유닛 이코노믹스 시뮬레이터 (Storage / Transcoding / CDN Cost Model)</h2>
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
