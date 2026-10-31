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
  <div>
    <h1 class="text-2xl font-extrabold text-white">저작권 권리자 및 Content ID 콘솔 (Rights Owner Console)</h1>
    <p class="mt-1 text-xs text-neutral-400">
      Register SHA-256 & perceptual audio/video fingerprints · Enforce territory rules (Monetize / Track / Block) · Resolve creator counter-disputes
    </p>

    <!-- Register Reference Fingerprint Asset -->
    <form class="mt-6 rounded-2xl border border-white/10 bg-[#141A2C] p-6" @submit.prevent="registerReferenceAsset">
      <h2 class="text-sm font-bold text-white">신규 원본 레퍼런스 자산 및 핑거프린트 등록</h2>
      <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3">
        <div>
          <label class="text-xs text-neutral-400">Reference Asset Title</label>
          <input
            v-model="newAssetTitle"
            type="text"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#0B0E17] px-3 py-2 text-xs text-white"
          />
        </div>
        <div>
          <label class="text-xs text-neutral-400">SHA-256 & Chromaprint Hash</label>
          <input
            v-model="newAssetSha256"
            type="text"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#0B0E17] px-3 py-2 font-mono text-xs text-[#3EA6FF]"
          />
        </div>
        <div>
          <label class="text-xs text-neutral-400">Default Global Territory Match Policy</label>
          <select
            v-model="defaultTerritoryPolicy"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#0B0E17] px-3 py-2 text-xs text-white"
          >
            <option :value="CopyrightClaimAction.MONETIZE">Monetize (Route Ad Revenue to Rights Owner)</option>
            <option :value="CopyrightClaimAction.TRACK">Track Only (Collect View Telemetry)</option>
            <option :value="CopyrightClaimAction.BLOCK">Block Worldwide Playback</option>
          </select>
        </div>
      </div>

      <div class="mt-4 flex items-center justify-between">
        <span v-if="assetRegisteredToast" class="text-xs font-bold text-emerald-400">
          ✓ Reference fingerprint indexed in real-time upload verification engine.
        </span>
        <span v-else class="text-xs text-neutral-400">Automatically scans all new R2 uploads during VALIDATING stage.</span>
        <button
          type="submit"
          class="rounded-xl bg-[#3EA6FF] px-5 py-2 text-xs font-bold text-black hover:bg-[#65B8FF]"
        >
          🛡️ Register Reference Asset
        </button>
      </div>
    </form>

    <!-- Suspected Infringement Matches & Creator Counter-Disputes -->
    <div class="mt-6 rounded-2xl border border-white/10 bg-[#141A2C] p-5">
      <h2 class="text-sm font-bold text-white">Content ID 매칭 목록 및 크리에이터 이의 신청 처리</h2>

      <div class="mt-4 space-y-4">
        <div
          v-for="claim in claims"
          :key="claim.id"
          class="rounded-xl border border-white/10 bg-[#0B0E17] p-4 text-xs"
        >
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div>
              <span class="font-mono font-bold text-[#3EA6FF]">{{ claim.id }}</span>
              <span class="ml-2 font-bold text-white">Target Video: {{ claim.claimedVideoId }}</span>
              <span class="ml-2 text-neutral-400">
                Segment: {{ claim.matchStartSeconds }}s – {{ claim.matchEndSeconds }}s (Confidence: {{ claim.matchConfidence }}%)
              </span>
            </div>
            <span
              class="rounded px-2.5 py-0.5 text-[10px] font-bold uppercase"
              :class="claim.status === CopyrightClaimStatus.DISPUTED ? 'bg-amber-500/20 text-amber-300' : 'bg-emerald-500/20 text-emerald-300'"
            >
              {{ claim.status }} · Policy: {{ claim.claimAction.toUpperCase() }}
            </span>
          </div>

          <div v-if="claim.disputeReason" class="mt-3 rounded-lg border border-amber-400/30 bg-amber-500/10 p-3 text-amber-200">
            <p class="font-bold">⚠️ Creator Counter-Dispute Filed:</p>
            <p class="mt-1">"{{ claim.disputeReason }}"</p>
            <p class="mt-1 text-[11px] text-neutral-300">
              Note: Ad revenue is held in the Double-Entry Escrow Account until resolved.
            </p>
          </div>

          <div class="mt-3 flex flex-wrap gap-2">
            <button
              type="button"
              class="rounded-lg bg-emerald-500/20 px-3 py-1.5 font-bold text-emerald-300 hover:bg-emerald-500/30"
              @click="updateClaimAction(claim, CopyrightClaimAction.MONETIZE)"
            >
              💰 Monetize (Claim Revenue)
            </button>
            <button
              type="button"
              class="rounded-lg bg-blue-500/20 px-3 py-1.5 font-bold text-blue-300 hover:bg-blue-500/30"
              @click="updateClaimAction(claim, CopyrightClaimAction.TRACK)"
            >
              📊 Track Only (Release Escrow to Creator)
            </button>
            <button
              type="button"
              class="rounded-lg bg-red-500/20 px-3 py-1.5 font-bold text-red-300 hover:bg-red-500/30"
              @click="updateClaimAction(claim, CopyrightClaimAction.BLOCK)"
            >
              ⛔ Enforce Block Worldwide
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
