<script setup lang="ts">
import type { AdCampaign } from '@gvp/shared-types'
import { AdFormat, CampaignStatus } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'

const api = usePartnerApi()

useSeoMeta({
  title: 'Advertiser Console — Ad Campaigns, CPM/CPC Bidding & Pacing'
})

const newCampaignName = ref('Galaxy S26 Ultra 4K HDR Launch — Pre-Roll & Shorts')
const newFormat = ref<AdFormat>(AdFormat.PREROLL)
const dailyBudget = ref(5000000)
const totalBudget = ref(150000000)
const cpmBid = ref(14500)
const cpcBid = ref(850)
const targetCategory = ref('tech,gaming,cinema')
const targetCountries = ref('KR,US,JP')

const { data: campaignsData, pending: campaignsPending, error: campaignsError } = await useAsyncData('partner-campaigns', async () => {
  return await api.getPartnerCampaigns()
})

const campaigns = ref<AdCampaign[]>(campaignsData.value || [])

const errorMessage = ref<string | null>(null)
const successMessage = ref<string | null>(null)
const creativeFile = ref<File | null>(null)
const creativeHeadline = ref('')
const creativeDestinationUrl = ref('')
const isUploading = ref(false)

const creativeAccept = computed(() =>
  newFormat.value === AdFormat.FEED ? 'image/jpeg,image/png,image/gif,image/webp' : 'video/mp4,video/webm,video/quicktime'
)

function onCreativeFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  creativeFile.value = input.files?.[0] || null
}

async function createCampaign() {
  errorMessage.value = null
  successMessage.value = null

  const payload: Partial<AdCampaign> = {
    name: newCampaignName.value,
    format: newFormat.value,
    status: CampaignStatus.DRAFT,
    dailyBudget: dailyBudget.value,
    totalBudget: totalBudget.value,
    cpmBid: cpmBid.value,
    cpcBid: cpcBid.value,
    targetCategories: targetCategory.value.split(',').map(s => s.trim()),
    targetCountries: targetCountries.value.split(',').map(s => s.trim())
  }
  try {
    const created = await api.createPartnerCampaign(payload)
    campaigns.value.unshift(created)
    successMessage.value = `캠페인 [${created.name}]이 초안으로 등록되었습니다.`
    if (creativeFile.value) {
      await uploadCreativeFor(created.id)
    }
  } catch (err: any) {
    errorMessage.value = `캠페인 생성 실패: ${err?.message || 'API 서버 통신 오류'}`
  }
}

async function uploadCreativeFor(campaignId: string) {
  if (!creativeFile.value) return
  isUploading.value = true
  try {
    const creative = await api.uploadCreative(campaignId, creativeFile.value, {
      headline: creativeHeadline.value || newCampaignName.value,
      destinationUrl: creativeDestinationUrl.value || undefined
    })
    successMessage.value = `소재 [${creative.id}]가 업로드되어 심사 대기 중입니다.`
    creativeFile.value = null
  } catch (err: any) {
    errorMessage.value = `소재 업로드 실패: ${err?.message || 'API 서버 통신 오류'}`
  } finally {
    isUploading.value = false
  }
}

async function submitForReview(cmp: AdCampaign) {
  errorMessage.value = null
  try {
    const updated = await api.updateCampaignStatus(cmp.id, 'pending_review')
    const idx = campaigns.value.findIndex(c => c.id === cmp.id)
    const current = idx >= 0 ? campaigns.value[idx] : undefined
    if (current) campaigns.value[idx] = { ...current, status: updated.status }
    successMessage.value = `캠페인 [${cmp.name}]이 심사에 제출되었습니다.`
  } catch (err: any) {
    errorMessage.value = `심사 제출 실패: ${err?.message || 'API 서버 통신 오류'}`
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="border-b border-slate-200 pb-5">
      <h1 class="text-xl font-semibold tracking-tight text-slate-900">광고주 캠페인 매니저 (Advertiser Campaign Console)</h1>
      <p class="mt-1 text-sm text-slate-600">
        Create Pre-Roll, Mid-Roll, and Shorts campaigns · Configure CPM/CPC bidding &amp; daily budget pacing
      </p>
    </div>

    <!-- Alert Notifications -->
    <div v-if="errorMessage" class="rounded-lg border border-red-200 bg-red-50 p-4 text-xs font-medium text-red-700">
      {{ errorMessage }}
    </div>
    <div v-if="successMessage" class="rounded-lg border border-emerald-200 bg-emerald-50 p-4 text-xs font-medium text-emerald-700">
      {{ successMessage }}
    </div>

    <!-- Create Campaign Form -->
    <form class="rounded-lg border border-slate-200 bg-white p-6" @submit.prevent="createCampaign">
      <h2 class="text-sm font-semibold text-slate-900">신규 광고 캠페인 생성 (<code class="font-mono text-xs text-slate-600">POST /v1/partners/campaigns</code>)</h2>
      <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3">
        <div class="md:col-span-2">
          <label class="text-xs font-medium text-slate-700">Campaign Name</label>
          <input
            v-model="newCampaignName"
            type="text"
            class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-sm text-slate-900 focus:border-slate-900 focus:bg-white focus:outline-none"
          />
        </div>
        <div>
          <label class="text-xs font-medium text-slate-700">Placement Format</label>
          <select
            v-model="newFormat"
            class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-sm text-slate-900 focus:border-slate-900 focus:bg-white focus:outline-none"
          >
            <option :value="AdFormat.PREROLL">Pre-Roll (In-Stream Skippable)</option>
            <option :value="AdFormat.MIDROLL">Mid-Roll (Watch Page Break)</option>
            <option :value="AdFormat.SHORTS">Vertical Shorts Interstitial</option>
            <option :value="AdFormat.FEED">Home Feed Sponsored Card</option>
          </select>
        </div>
        <div>
          <label class="text-xs font-medium text-slate-700">Daily Budget (KRW)</label>
          <input
            v-model.number="dailyBudget"
            type="number"
            class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-sm text-slate-900 tabular-nums focus:border-slate-900 focus:bg-white focus:outline-none"
          />
        </div>
        <div>
          <label class="text-xs font-medium text-slate-700">CPM Bid (KRW per 1,000 impressions)</label>
          <input
            v-model.number="cpmBid"
            type="number"
            class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-sm text-slate-900 tabular-nums focus:border-slate-900 focus:bg-white focus:outline-none"
          />
        </div>
        <div>
          <label class="text-xs font-medium text-slate-700">CPC Bid (KRW per Click)</label>
          <input
            v-model.number="cpcBid"
            type="number"
            class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-sm text-slate-900 tabular-nums focus:border-slate-900 focus:bg-white focus:outline-none"
          />
        </div>
      </div>

      <div class="mt-5 grid grid-cols-1 gap-4 border-t border-slate-100 pt-4 md:grid-cols-3">
        <div>
          <label class="text-xs font-medium text-slate-700">Creative File (mp4/webm/mov or jpg/png/gif/webp)</label>
          <input
            type="file"
            :accept="creativeAccept"
            class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-xs text-slate-900 focus:border-slate-900 focus:bg-white focus:outline-none"
            @change="onCreativeFileChange"
          />
        </div>
        <div>
          <label class="text-xs font-medium text-slate-700">Creative Headline</label>
          <input
            v-model="creativeHeadline"
            type="text"
            placeholder="Optional — defaults to campaign name"
            class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-sm text-slate-900 focus:border-slate-900 focus:bg-white focus:outline-none"
          />
        </div>
        <div>
          <label class="text-xs font-medium text-slate-700">Destination URL</label>
          <input
            v-model="creativeDestinationUrl"
            type="url"
            placeholder="https://…"
            class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-sm text-slate-900 focus:border-slate-900 focus:bg-white focus:outline-none"
          />
        </div>
      </div>

      <div class="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-slate-100 pt-4">
        <p class="text-xs text-slate-500">
          Campaigns start as <code class="rounded-md bg-slate-100 px-1.5 py-0.5 font-mono text-slate-800">draft</code> — upload a creative, then submit for review.
        </p>
        <button
          type="submit"
          :disabled="isUploading"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors disabled:opacity-60"
        >
          {{ isUploading ? 'Uploading Creative…' : 'Create Draft Campaign' }}
        </button>
      </div>
    </form>

    <!-- Active Campaigns Table -->
    <div v-if="campaignsPending" class="rounded-lg border border-slate-200 bg-white p-8">
      <div class="h-6 animate-pulse rounded-md bg-slate-100" />
    </div>
    <div v-else-if="campaignsError" class="rounded-lg border border-red-200 bg-red-50 p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-red-700">캠페인 목록을 불러오지 못했습니다.</p>
    </div>
    <GvpEmptyState
      v-else-if="campaigns.length === 0"
      title="등록된 캠페인이 없습니다"
      description="위 양식으로 첫 캠페인을 초안으로 등록하세요."
    />

    <div v-else class="overflow-x-auto rounded-lg border border-slate-200 bg-white">
      <table class="w-full text-left">
        <thead>
          <tr class="bg-slate-50 border-y border-slate-200 text-xs font-semibold text-slate-600 uppercase tracking-wider">
            <th class="py-2.5 px-4">Campaign</th>
            <th class="py-2.5 px-4">Placement</th>
            <th class="py-2.5 px-4">Budget Pacing</th>
            <th class="py-2.5 px-4">CPM / CPC</th>
            <th class="py-2.5 px-4">Impressions</th>
            <th class="py-2.5 px-4">Clicks (CTR)</th>
            <th class="py-2.5 px-4">Status</th>
            <th class="py-2.5 px-4">Review</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          <tr v-for="cmp in campaigns" :key="cmp.id" class="hover:bg-slate-50">
            <td class="py-3 px-4 text-sm font-medium text-slate-900 border-b border-slate-100">{{ cmp.name }}</td>
            <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
              <span class="rounded-md bg-slate-100 px-2 py-0.5 font-mono text-xs uppercase text-slate-700">{{ cmp.format }}</span>
            </td>
            <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
              ₩{{ cmp.spentAmount.toLocaleString() }} / ₩{{ cmp.totalBudget.toLocaleString() }}
            </td>
            <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
              ₩{{ cmp.cpmBid.toLocaleString() }} / ₩{{ cmp.cpcBid.toLocaleString() }}
            </td>
            <td class="py-3 px-4 font-mono text-sm text-slate-900 border-b border-slate-100 tabular-nums">
              {{ cmp.impressions.toLocaleString() }}
            </td>
            <td class="py-3 px-4 font-mono text-sm text-slate-900 border-b border-slate-100 tabular-nums">
              {{ cmp.clicks.toLocaleString() }}
              ({{ cmp.impressions ? ((cmp.clicks / cmp.impressions) * 100).toFixed(2) : '0.00' }}%)
            </td>
            <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
              <span class="rounded-md border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-xs font-medium uppercase text-emerald-700">
                {{ cmp.status }}
              </span>
            </td>
            <td class="py-3 px-4 border-b border-slate-100">
              <button
                v-if="cmp.status === 'draft'"
                type="button"
                class="rounded-md border border-slate-200 bg-white px-2.5 py-1 text-xs font-medium text-slate-700 hover:bg-slate-100"
                @click="submitForReview(cmp)"
              >
                Submit for review
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
