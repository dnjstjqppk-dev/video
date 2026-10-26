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
    id: 'cmp_samsung_01',
    advertiserId: 'usr_adv_1',
    organizationId: 'org_adv_samsung',
    name: 'Galaxy S26 Ultra 4K Cinema Pre-Roll Campaign',
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
    organizationId: 'org_adv_samsung',
    name: 'Odyssey OLED G9 Vertical Shorts Sponsored Feed',
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
  <div>
    <h1 class="text-2xl font-extrabold text-white">광고주 캠페인 매니저 (Advertiser Campaign Console)</h1>
    <p class="mt-1 text-xs text-neutral-400">
      Create Pre-Roll, Mid-Roll, and Shorts campaigns · Configure CPM/CPC bidding & daily budget pacing
    </p>

    <!-- Alert Notifications -->
    <div v-if="errorMessage" class="mt-4 rounded-xl border border-red-500/30 bg-red-950/40 p-4 text-xs font-semibold text-red-200">
      {{ errorMessage }}
    </div>
    <div v-if="successMessage" class="mt-4 rounded-xl border border-emerald-500/30 bg-emerald-950/40 p-4 text-xs font-semibold text-emerald-200">
      {{ successMessage }}
    </div>

    <!-- Create Campaign Form -->
    <form class="mt-6 rounded-2xl border border-white/10 bg-[#141A2C] p-6" @submit.prevent="createCampaign">
      <h2 class="text-sm font-bold text-white">신규 광고 캠페인 생성 (`POST /v1/partners/campaigns`)</h2>
      <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3">
        <div class="md:col-span-2">
          <label class="text-xs text-neutral-400">Campaign Name</label>
          <input
            v-model="newCampaignName"
            type="text"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#0B0E17] px-3.5 py-2 text-xs text-white"
          />
        </div>
        <div>
          <label class="text-xs text-neutral-400">Placement Format</label>
          <select
            v-model="newFormat"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#0B0E17] px-3 py-2 text-xs text-white"
          >
            <option :value="AdFormat.PREROLL">Pre-Roll (In-Stream Skippable)</option>
            <option :value="AdFormat.MIDROLL">Mid-Roll (Watch Page Break)</option>
            <option :value="AdFormat.SHORTS">Vertical Shorts Interstitial</option>
            <option :value="AdFormat.FEED">Home Feed Sponsored Card</option>
          </select>
        </div>
        <div>
          <label class="text-xs text-neutral-400">Daily Budget (KRW)</label>
          <input
            v-model.number="dailyBudget"
            type="number"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#0B0E17] px-3 py-2 text-xs text-white"
          />
        </div>
        <div>
          <label class="text-xs text-neutral-400">CPM Bid (KRW per 1,000 impressions)</label>
          <input
            v-model.number="cpmBid"
            type="number"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#0B0E17] px-3 py-2 text-xs text-white"
          />
        </div>
        <div>
          <label class="text-xs text-neutral-400">CPC Bid (KRW per Click)</label>
          <input
            v-model.number="cpcBid"
            type="number"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#0B0E17] px-3 py-2 text-xs text-white"
          />
        </div>
      </div>

      <div class="mt-4 flex items-center justify-between">
        <p class="text-xs text-neutral-400">
          Creative Asset: <code class="text-[#3EA6FF]">s26_ultra_hdr_master_1080p60.mp4</code> (VAST 4.2 Verified)
        </p>
        <button
          type="submit"
          class="rounded-xl bg-[#3EA6FF] px-5 py-2 text-xs font-bold text-black hover:bg-[#65B8FF]"
        >
          🚀 Launch Ad Campaign
        </button>
      </div>
    </form>

    <!-- Active Campaigns Table -->
    <div class="mt-6 overflow-x-auto rounded-2xl border border-white/10 bg-[#141A2C]">
      <table class="w-full text-left text-xs">
        <thead class="border-b border-white/10 bg-white/5 uppercase text-neutral-400">
          <tr>
            <th class="p-3.5">Campaign</th>
            <th class="p-3.5">Placement</th>
            <th class="p-3.5">Budget Pacing</th>
            <th class="p-3.5">CPM / CPC</th>
            <th class="p-3.5">Impressions</th>
            <th class="p-3.5">Clicks (CTR)</th>
            <th class="p-3.5">Status</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-white/10">
          <tr v-for="cmp in campaigns" :key="cmp.id" class="hover:bg-white/5">
            <td class="p-3.5 font-bold text-white">{{ cmp.name }}</td>
            <td class="p-3.5 uppercase text-[#3EA6FF]">{{ cmp.format }}</td>
            <td class="p-3.5">
              ₩{{ cmp.spentAmount.toLocaleString() }} / ₩{{ cmp.totalBudget.toLocaleString() }}
            </td>
            <td class="p-3.5">₩{{ cmp.cpmBid.toLocaleString() }} / ₩{{ cmp.cpcBid.toLocaleString() }}</td>
            <td class="p-3.5 font-mono">{{ cmp.impressions.toLocaleString() }}</td>
            <td class="p-3.5 font-mono">
              {{ cmp.clicks.toLocaleString() }}
              ({{ cmp.impressions ? ((cmp.clicks / cmp.impressions) * 100).toFixed(2) : '0.00' }}%)
            </td>
            <td class="p-3.5">
              <span class="rounded bg-emerald-500/20 px-2 py-0.5 text-[10px] font-bold uppercase text-emerald-400">
                {{ cmp.status }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
