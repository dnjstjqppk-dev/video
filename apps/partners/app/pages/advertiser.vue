<script setup lang="ts">
import type { AdCampaign, AdDecisionResponse } from '@gvp/shared-types'
import { AdFormat, CampaignStatus } from '@gvp/shared-types'

const api = usePartnerApi()

useSeoMeta({
  title: 'Dedicated Advertiser Portal — StreamVerse Media Buying'
})

// 1. Fetch campaigns
const { data: campaignsData, pending, refresh } = await useAsyncData('advertiser-campaigns', async () => {
  try {
    return await api.getPartnerCampaigns()
  } catch {
    return []
  }
})

const campaigns = computed<AdCampaign[]>(() => campaignsData.value || [])

// 2. Computed KPI aggregates
const totalBudget = computed(() => campaigns.value.reduce((acc, c) => acc + (c.totalBudget || 0), 0))
const totalSpent = computed(() => campaigns.value.reduce((acc, c) => acc + (c.spentAmount || 0), 0))
const totalImpressions = computed(() => campaigns.value.reduce((acc, c) => acc + (c.impressions || 0), 0))
const totalClicks = computed(() => campaigns.value.reduce((acc, c) => acc + (c.clicks || 0), 0))
const averageCtr = computed(() => {
  if (totalImpressions.value === 0) return 0
  return (totalClicks.value / totalImpressions.value) * 100
})

// 3. Ad Preview Simulator
const previewSlot = ref('watch_preroll')
const previewCategory = ref('Technology')
const previewResult = ref<AdDecisionResponse | null>(null)
const isPreviewLoading = ref(false)

async function runAdPreview() {
  isPreviewLoading.value = true
  try {
    previewResult.value = await api.getAdPreview({
      slot_id: previewSlot.value,
      category: previewCategory.value
    })
  } catch (err: any) {
    previewResult.value = {
      ad_available: false,
      reason: err?.message || 'PREVIEW_FETCH_FAILED'
    }
  } finally {
    isPreviewLoading.value = false
  }
}

onMounted(() => {
  runAdPreview()
})
</script>

<template>
  <div>
    <!-- Header -->
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div>
        <div class="flex items-center gap-2">
          <span class="rounded bg-blue-500/20 px-2.5 py-0.5 text-xs font-bold text-blue-400">
            ARTICLE 6 COMPLIANT
          </span>
          <h1 class="text-2xl font-black text-white tracking-tight">독립 광고주 포털 (Advertiser Portal)</h1>
        </div>
        <p class="mt-1 text-xs text-neutral-400">
          Apex Global Media Group (Demo) · Enterprise Ad Account #ADV-DEMO-2026-001
        </p>
      </div>

      <div class="flex items-center gap-3">
        <NuxtLink
          to="/campaigns"
          class="rounded-xl bg-blue-600 px-4 py-2 text-xs font-bold text-white shadow-lg transition hover:bg-blue-500"
        >
          + 신규 캠페인 등록 (New Campaign)
        </NuxtLink>
        <button
          type="button"
          @click="refresh()"
          class="rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 hover:bg-white/10"
        >
          🔄 새로고침
        </button>
      </div>
    </div>

    <!-- Account Overview & Balance Card -->
    <div class="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-5">
        <span class="text-xs font-medium text-neutral-400">광고 계정 잔액 (Prepaid Balance)</span>
        <p class="mt-2 text-2xl font-black text-white">₩142,500,000</p>
        <div class="mt-2 flex items-center gap-1.5 text-[11px] text-emerald-400">
          <span>●</span>
          <span>KRW 법인 충전 계좌 연동 완료</span>
        </div>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-5">
        <span class="text-xs font-medium text-neutral-400">총 예산 집행액 (Spend / Budget)</span>
        <p class="mt-2 text-2xl font-black text-white">
          ₩{{ (totalSpent / 10000).toLocaleString() }}만
          <span class="text-xs font-normal text-neutral-400">/ ₩{{ (totalBudget / 10000).toLocaleString() }}만</span>
        </p>
        <div class="mt-2 flex items-center gap-1.5 text-[11px] text-neutral-400">
          <span>소진율: {{ totalBudget > 0 ? ((totalSpent / totalBudget) * 100).toFixed(1) : 0 }}%</span>
        </div>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-5">
        <span class="text-xs font-medium text-neutral-400">총 검증 노출수 (Verified Impressions)</span>
        <p class="mt-2 text-2xl font-black text-blue-400">{{ totalImpressions.toLocaleString() }}</p>
        <div class="mt-2 flex items-center gap-1.5 text-[11px] text-neutral-400">
          <span>CPM 실시간 과금 반영</span>
        </div>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-5">
        <span class="text-xs font-medium text-neutral-400">클릭수 & 평균 클릭률 (CTR)</span>
        <p class="mt-2 text-2xl font-black text-amber-400">
          {{ totalClicks.toLocaleString() }}
          <span class="text-sm font-normal text-neutral-400">({{ averageCtr.toFixed(2) }}%)</span>
        </p>
        <div class="mt-2 flex items-center gap-1.5 text-[11px] text-neutral-400">
          <span>목적지 랜딩 유입 전환율</span>
        </div>
      </div>
    </div>

    <!-- Middle Section: Live Ad Preview Simulator & Active Campaigns -->
    <div class="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Live Ad Preview Simulator -->
      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-6">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-bold text-white">광고 렌더링 시뮬레이터 (Ad Preview)</h2>
          <span class="rounded bg-white/10 px-2 py-0.5 text-[10px] text-neutral-300">Live Engine</span>
        </div>
        <p class="mt-1 text-xs text-neutral-400">
          Test real creative rendering against Ad Decision Engine
        </p>

        <div class="mt-4 space-y-3">
          <div>
            <label class="text-[11px] text-neutral-400">Slot ID</label>
            <select
              v-model="previewSlot"
              class="mt-1 w-full rounded-xl border border-white/15 bg-[#0B0E17] px-3 py-2 text-xs text-white"
              @change="runAdPreview"
            >
              <option value="watch_preroll">watch_preroll (Main Player 5s Skippable)</option>
              <option value="home_feed_banner">home_feed_banner (Native Feed Card)</option>
              <option value="shorts_sponsored">shorts_sponsored (Shorts Vertical Feed)</option>
            </select>
          </div>

          <div>
            <label class="text-[11px] text-neutral-400">Target Category</label>
            <select
              v-model="previewCategory"
              class="mt-1 w-full rounded-xl border border-white/15 bg-[#0B0E17] px-3 py-2 text-xs text-white"
              @change="runAdPreview"
            >
              <option value="Technology">Technology & Engineering</option>
              <option value="Gaming">Gaming & Esports</option>
              <option value="Cinema">Cinema & Entertainment</option>
            </select>
          </div>

          <div class="mt-4 rounded-xl border border-white/10 bg-[#0B0E17] p-4">
            <div v-if="isPreviewLoading" class="py-6 text-center text-xs text-neutral-400 animate-pulse">
              의사결정 엔진 조회 중... (Querying Decision Engine)
            </div>
            <div v-else-if="previewResult?.ad_available && previewResult?.creative" class="space-y-3">
              <div class="flex items-center justify-between">
                <span class="rounded bg-emerald-500/20 px-2 py-0.5 text-[10px] font-bold text-emerald-400">
                  MATCH SUCCESS
                </span>
                <span class="text-[10px] font-mono text-neutral-400">{{ previewResult.creative.slotId }}</span>
              </div>

              <!-- Media Preview -->
              <div class="aspect-video w-full overflow-hidden rounded-lg bg-black border border-white/10">
                <img
                  v-if="previewResult.creative.mediaUrl"
                  :src="previewResult.creative.mediaUrl"
                  :alt="previewResult.creative.title"
                  class="h-full w-full object-cover"
                />
              </div>

              <div>
                <h4 class="text-xs font-bold text-white">{{ previewResult.creative.title }}</h4>
                <p class="text-[11px] text-neutral-400">Advertiser: {{ previewResult.creative.advertiserName }}</p>
                <div class="mt-2 flex items-center justify-between text-[11px]">
                  <span class="text-neutral-400">CTA: {{ previewResult.creative.callToAction }}</span>
                  <a
                    :href="previewResult.creative.destinationUrl"
                    target="_blank"
                    class="text-blue-400 hover:underline"
                  >
                    목적지 URL ↗
                  </a>
                </div>
              </div>
            </div>
            <div v-else class="py-6 text-center text-xs text-amber-400">
              미송출 (No-Fill): {{ previewResult?.reason || '적격 캠페인 없음' }}
            </div>
          </div>
        </div>
      </div>

      <!-- Active Campaigns Table -->
      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-6 lg:col-span-2">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-bold text-white">등록 캠페인 목록 (Ad Campaigns)</h2>
          <span class="text-xs text-neutral-400">{{ campaigns.length }} Campaigns Registered</span>
        </div>
        <p class="mt-1 text-xs text-neutral-400">Real-time status, CPM/CPC bids and flight schedule</p>

        <div class="mt-4 overflow-x-auto">
          <table class="w-full text-left text-xs text-neutral-300">
            <thead class="border-b border-white/10 text-[11px] uppercase text-neutral-400">
              <tr>
                <th class="py-2.5">캠페인명</th>
                <th class="py-2.5">형식</th>
                <th class="py-2.5">상태</th>
                <th class="py-2.5">입찰가 (CPM / CPC)</th>
                <th class="py-2.5">집행액</th>
                <th class="py-2.5">노출 / 클릭</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-white/5 font-sans">
              <tr v-for="c in campaigns" :key="c.id" class="hover:bg-white/5">
                <td class="py-3 font-semibold text-white max-w-[160px] truncate">{{ c.name }}</td>
                <td class="py-3">
                  <span class="rounded bg-white/10 px-2 py-0.5 text-[10px] font-mono uppercase text-neutral-300">
                    {{ c.format }}
                  </span>
                </td>
                <td class="py-3">
                  <span
                    class="rounded px-2 py-0.5 text-[10px] font-bold uppercase"
                    :class="c.status === 'active' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-neutral-800 text-neutral-400'"
                  >
                    {{ c.status }}
                  </span>
                </td>
                <td class="py-3 font-mono text-[11px]">
                  ₩{{ c.cpmBid?.toLocaleString() }} / ₩{{ c.cpcBid?.toLocaleString() }}
                </td>
                <td class="py-3 font-mono text-neutral-200">
                  ₩{{ (c.spentAmount || 0).toLocaleString() }}
                </td>
                <td class="py-3 text-[11px]">
                  <span class="text-blue-400">{{ (c.impressions || 0).toLocaleString() }}</span> /
                  <span class="text-amber-400">{{ (c.clicks || 0).toLocaleString() }}</span>
                </td>
              </tr>
              <tr v-if="!campaigns.length">
                <td colspan="6" class="py-8 text-center text-xs text-neutral-500">
                  등록된 광고 캠페인이 없습니다. (+ 신규 캠페인 등록을 진행하세요)
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
