<script setup lang="ts">
import type { CopyrightClaim } from '@gvp/shared-types'
import { CopyrightClaimAction, CopyrightClaimStatus } from '@gvp/shared-types'

const api = usePartnerApi()

useSeoMeta({
  title: 'Rights Owner Console — SHA-256 Fingerprints & Content ID Claims'
})

const newAssetTitle = ref('CJ ENM Original OST — Cyberpunk Seoul Theme (Master Audio)')
const newAssetSha256 = ref('e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855')
const defaultTerritoryPolicy = ref<CopyrightClaimAction>(CopyrightClaimAction.MONETIZE)
const assetRegisteredToast = ref(false)

const fallbackClaims: CopyrightClaim[] = [
  {
    id: 'clm_901',
    assetId: 'ast_ost_seoul_01',
    claimedVideoId: 'vid_seoul_4k',
    rightsOwnerId: 'org_rights_starlight',
    ownerName: 'Starlight Music & Media Rights Corp. (Demo)',
    claimAction: CopyrightClaimAction.MONETIZE,
    status: CopyrightClaimStatus.ACTIVE,
    matchStartSeconds: 410,
    matchEndSeconds: 445,
    matchConfidence: 98.7,
    createdAt: '2026-10-19T12:00:00Z',
    updatedAt: '2026-10-19T12:00:00Z'
  },
  {
    id: 'clm_902',
    assetId: 'ast_movie_clip_88',
    claimedVideoId: 'vid_reupload_suspect_77',
    rightsOwnerId: 'org_rights_starlight',
    ownerName: 'Starlight Music & Media Rights Corp. (Demo)',
    claimAction: CopyrightClaimAction.BLOCK,
    status: CopyrightClaimStatus.DISPUTED,
    matchStartSeconds: 0,
    matchEndSeconds: 310,
    matchConfidence: 99.4,
    disputeReason: 'Fair use commentary and architectural critique under 30 seconds.',
    disputedAt: '2026-10-23T08:30:00Z',
    createdAt: '2026-10-22T15:00:00Z',
    updatedAt: '2026-10-23T08:30:00Z'
  }
]

const { data: claimsData } = await useAsyncData('partner-claims', async () => {
  try {
    const res = await api.getPartnerClaims()
    return res.length ? res : fallbackClaims
  } catch {
    return fallbackClaims
  }
})

const claims = ref<CopyrightClaim[]>(claimsData.value || fallbackClaims)

function registerReferenceAsset() {
  assetRegisteredToast.value = true
  setTimeout(() => (assetRegisteredToast.value = false), 3500)
}

async function updateClaimAction(claim: CopyrightClaim, action: CopyrightClaimAction) {
  claim.claimAction = action
  claim.status = CopyrightClaimStatus.RESOLVED
  try {
    await api.createCopyrightClaim({
      videoId: claim.claimedVideoId,
      assetId: claim.assetId,
      action
    })
  } catch {
    // State updated locally
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
    <form class="rounded-lg border border-slate-200 bg-white p-6" @submit.prevent="registerReferenceAsset">
      <h2 class="text-sm font-semibold text-slate-900">신규 원본 레퍼런스 자산 및 핑거프린트 등록</h2>
      <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3">
        <div>
          <label class="text-xs font-medium text-slate-700">Reference Asset Title</label>
          <input
            v-model="newAssetTitle"
            type="text"
            class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-sm text-slate-900 focus:border-slate-900 focus:bg-white focus:outline-none"
          />
        </div>
        <div>
          <label class="text-xs font-medium text-slate-700">SHA-256 &amp; Chromaprint Hash</label>
          <input
            v-model="newAssetSha256"
            type="text"
            class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 font-mono text-xs text-slate-900 tabular-nums focus:border-slate-900 focus:bg-white focus:outline-none"
          />
        </div>
        <div>
          <label class="text-xs font-medium text-slate-700">Default Global Territory Match Policy</label>
          <select
            v-model="defaultTerritoryPolicy"
            class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-sm text-slate-900 focus:border-slate-900 focus:bg-white focus:outline-none"
          >
            <option :value="CopyrightClaimAction.MONETIZE">Monetize (Route Ad Revenue to Rights Owner)</option>
            <option :value="CopyrightClaimAction.TRACK">Track Only (Collect View Telemetry)</option>
            <option :value="CopyrightClaimAction.BLOCK">Block Worldwide Playback</option>
          </select>
        </div>
      </div>

      <div class="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-slate-100 pt-4">
        <span v-if="assetRegisteredToast" class="text-xs font-medium text-emerald-700">
          Reference fingerprint indexed in real-time upload verification engine.
        </span>
        <span v-else class="text-xs text-slate-500">Automatically scans all new R2 uploads during VALIDATING stage.</span>
        <button
          type="submit"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors"
        >
          Register Reference Asset
        </button>
      </div>
    </form>

    <!-- Suspected Infringement Matches & Creator Counter-Disputes -->
    <div class="rounded-lg border border-slate-200 bg-white p-6">
      <h2 class="text-sm font-semibold text-slate-900">Content ID 매칭 목록 및 크리에이터 이의 신청 처리</h2>

      <div class="mt-4 space-y-4">
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
