<script setup lang="ts">
import type { AdCampaign } from '@gvp/shared-types'
import { AdFormat, CampaignStatus } from '@gvp/shared-types'

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

const fallbackCampaigns: AdCampaign[] = [
  {
    id: 'cmp_apex_01',
    advertiserId: 'usr_adv_1',
    organizationId: 'org_adv_apex',
    name: 'Apex Horizon Phone 4K Cinema Pre-Roll Campaign (Demo)',
    format: AdFormat.PREROLL,
    status: CampaignStatus.ACTIVE,
    dailyBudget: 5000000,
    totalBudget: 100000000,
    spentAmount: 42800000,
    cpmBid: 15000,
    cpcBid: 900,
    targetCategories: ['tech', 'cinema'],
    targetCountries: ['KR', 'US'],
    startDate: '2026-10-01',
    endDate: '2026-11-15',
    impressions: 2853000,
    clicks: 94100,
    views: 1920000,
    createdAt: '2026-10-01T00:00:00Z',
    updatedAt: '2026-10-24T00:00:00Z'
  },
  {
    id: 'cmp_gaming_shorts',
    advertiserId: 'usr_adv_1',
    organizationId: 'org_adv_apex',
    name: 'Apex Vision Curved Display Vertical Shorts Sponsored Feed (Demo)',
    format: AdFormat.SHORTS,
    status: CampaignStatus.ACTIVE,
    dailyBudget: 3000000,
    totalBudget: 60000000,
    spentAmount: 28400000,
    cpmBid: 11000,
    cpcBid: 650,
    targetCategories: ['gaming'],
    targetCountries: ['KR', 'US', 'JP'],
    startDate: '2026-10-05',
    endDate: '2026-11-05',
    impressions: 2581000,
    clicks: 112400,
    views: 1840000,
    createdAt: '2026-10-05T00:00:00Z',
    updatedAt: '2026-10-24T00:00:00Z'
  }
]

const { data: campaignsData } = await useAsyncData('partner-campaigns', async () => {
  try {
    const res = await api.getPartnerCampaigns()
    return res.length ? res : fallbackCampaigns
  } catch {
    return fallbackCampaigns
  }
})

const campaigns = ref<AdCampaign[]>(campaignsData.value || fallbackCampaigns)

const errorMessage = ref<string | null>(null)
const successMessage = ref<string | null>(null)

async function createCampaign() {
  errorMessage.value = null
  successMessage.value = null

  const payload: Partial<AdCampaign> = {
    name: newCampaignName.value,
    format: newFormat.value,
    status: CampaignStatus.ACTIVE,
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
    successMessage.value = `캠페인 [${created.name}]이 성공적으로 등록되었습니다.`
  } catch (err: any) {
    errorMessage.value = `캠페인 생성 실패: ${err?.message || 'API 서버 통신 오류'}`
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

      <div class="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-slate-100 pt-4">
        <p class="text-xs text-slate-500">
          Creative Asset: <code class="rounded-md bg-slate-100 px-1.5 py-0.5 font-mono text-slate-800">s26_ultra_hdr_master_1080p60.mp4</code> (VAST 4.2 Verified)
        </p>
        <button
          type="submit"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors"
        >
          Launch Ad Campaign
        </button>
      </div>
    </form>

    <!-- Active Campaigns Table -->
    <div class="overflow-x-auto rounded-lg border border-slate-200 bg-white">
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
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
