<script setup lang="ts">
import type { CopyrightClaim } from '@gvp/shared-types'
import { CopyrightClaimAction, CopyrightClaimStatus } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'

const api = usePartnerApi()

useSeoMeta({
  title: 'Rights Owner Console — SHA-256 Fingerprints & Content ID Claims'
})

const claimError = ref<string | null>(null)

const { data: claimsData, pending: claimsPending, error: claimsError } = await useAsyncData('partner-claims', async () => {
  return await api.getPartnerClaims()
})

const claims = computed(() => claimsData.value || [])

async function updateClaimAction(claim: CopyrightClaim, action: CopyrightClaimAction) {
  claimError.value = null
  try {
    const updated = await api.createCopyrightClaim({
      videoId: claim.claimedVideoId,
      assetId: claim.assetId,
      action
    })
    claim.claimAction = updated.claimAction
    claim.status = updated.status
  } catch (e: unknown) {
    claimError.value = e instanceof Error ? e.message : '클레임 처리에 실패했습니다.'
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="border-b border-slate-200 pb-5">
      <h1 class="text-xl font-semibold tracking-tight text-slate-900">저작권 권리자 및 Content ID 콘솔 (Rights Owner Console)</h1>
      <p class="mt-1 text-sm text-slate-600">
        Register SHA-256 &amp; perceptual audio/video fingerprints · Enforce territory rules (Monetize / Track / Block) · Resolve creator counter-disputes
      </p>
    </div>

    <!-- Register Reference Fingerprint Asset -->
    <div class="rounded-lg border border-dashed border-slate-300 bg-slate-50 p-6 text-center">
      <h2 class="text-sm font-semibold text-slate-600">원본 레퍼런스 자산 등록 연동 준비 중</h2>
      <p class="mt-1 text-xs text-slate-500">핑거프린트 등록 API 연결 후 제공됩니다.</p>
    </div>

    <!-- Suspected Infringement Matches & Creator Counter-Disputes -->
    <div class="rounded-lg border border-slate-200 bg-white p-6">
      <h2 class="text-sm font-semibold text-slate-900">Content ID 매칭 목록 및 크리에이터 이의 신청 처리</h2>

      <p v-if="claimError" class="mt-3 rounded-md border border-red-200 bg-red-50 p-2 text-xs font-semibold text-red-700" role="alert">
        {{ claimError }}
      </p>
      <div v-if="claimsPending" class="mt-4 space-y-3">
        <div v-for="i in 2" :key="i" class="h-20 animate-pulse rounded-lg border border-slate-200 bg-slate-100" />
      </div>
      <div v-else-if="claimsError" class="mt-4 rounded-lg border border-red-200 bg-red-50 p-6 text-center" role="alert">
        <p class="text-xs font-semibold text-red-700">클레임 목록을 불러오지 못했습니다.</p>
      </div>
      <GvpEmptyState
        v-else-if="claims.length === 0"
        title="진행 중인 클레임이 없습니다"
        description="새 매칭이 접수되면 여기에 표시됩니다."
      />

      <div v-else class="mt-4 space-y-4">
        <div
          v-for="claim in claims"
          :key="claim.id"
          class="rounded-lg border border-slate-200 bg-slate-50 p-4 text-sm"
        >
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div class="flex flex-wrap items-center gap-2">
              <span class="rounded-md bg-slate-100 px-2 py-0.5 font-mono text-xs font-semibold text-slate-800 tabular-nums">{{ claim.id }}</span>
              <span class="font-semibold text-slate-900">Target Video: {{ claim.claimedVideoId }}</span>
              <span class="text-xs text-slate-600 tabular-nums">
                Segment: {{ claim.matchStartSeconds }}s – {{ claim.matchEndSeconds }}s (Confidence: {{ claim.matchConfidence }}%)
              </span>
            </div>
            <span
              class="rounded-md px-2.5 py-0.5 text-xs font-medium uppercase"
              :class="claim.status === CopyrightClaimStatus.DISPUTED ? 'border border-amber-200 bg-amber-50 text-amber-800' : 'border border-emerald-200 bg-emerald-50 text-emerald-700'"
            >
              {{ claim.status }} · Policy: {{ claim.claimAction.toUpperCase() }}
            </span>
          </div>

          <div v-if="claim.disputeReason" class="mt-3 rounded-lg border border-amber-200 bg-amber-50 p-3 text-xs text-amber-900">
            <p class="font-semibold">Creator Counter-Dispute Filed</p>
            <p class="mt-1 text-slate-700">"{{ claim.disputeReason }}"</p>
            <p class="mt-1 text-slate-500">
              Note: Ad revenue is held in the Double-Entry Escrow Account until resolved.
            </p>
          </div>

          <div class="mt-4 flex flex-wrap gap-2">
            <button
              type="button"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors"
              @click="updateClaimAction(claim, CopyrightClaimAction.MONETIZE)"
            >
              Monetize (Claim Revenue)
            </button>
            <button
              type="button"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
              @click="updateClaimAction(claim, CopyrightClaimAction.TRACK)"
            >
              Track Only (Release Escrow to Creator)
            </button>
            <button
              type="button"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
              @click="updateClaimAction(claim, CopyrightClaimAction.BLOCK)"
            >
              Enforce Block Worldwide
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
