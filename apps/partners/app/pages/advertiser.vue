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
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col gap-4 border-b border-slate-200 pb-5 md:flex-row md:items-center md:justify-between">
      <div>
        <div class="flex flex-wrap items-center gap-2">
          <span class="rounded-md border border-slate-200 bg-slate-100 px-2.5 py-0.5 text-xs font-medium text-slate-700">
            ARTICLE 6 COMPLIANT
          </span>
          <h1 class="text-xl font-semibold tracking-tight text-slate-900">독립 광고주 포털 (Advertiser Portal)</h1>
        </div>
        <p class="mt-1 text-sm text-slate-600 tabular-nums">
          Apex Global Media Group (Demo) · Enterprise Ad Account #ADV-DEMO-2026-001
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <NuxtLink
          to="/campaigns"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors"
        >
          + 신규 캠페인 등록 (New Campaign)
        </NuxtLink>
        <button
          type="button"
          class="inline-flex items-center justify-center gap-1.5 h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
          @click="refresh()"
        >
          <svg class="h-4 w-4 text-slate-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>새로고침</span>
        </button>
      </div>
    </div>

    <!-- Account Overview & Balance Card -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <span class="text-xs font-medium uppercase tracking-wider text-slate-500">광고 계정 잔액 (Prepaid Balance)</span>
        <p class="mt-2 text-2xl font-semibold text-slate-900 tabular-nums">₩142,500,000</p>
        <div class="mt-2 flex items-center gap-1.5 text-xs text-emerald-700">
          <span class="h-2 w-2 rounded-full bg-emerald-600" />
          <span>KRW 법인 충전 계좌 연동 완료</span>
        </div>
      </div>

      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <span class="text-xs font-medium uppercase tracking-wider text-slate-500">총 예산 집행액 (Spend / Budget)</span>
        <p class="mt-2 text-2xl font-semibold text-slate-900 tabular-nums">
          ₩{{ (totalSpent / 10000).toLocaleString() }}만
          <span class="text-xs font-normal text-slate-500 tabular-nums">/ ₩{{ (totalBudget / 10000).toLocaleString() }}만</span>
        </p>
        <div class="mt-2 flex items-center gap-1.5 text-xs text-slate-500 tabular-nums">
          <span>소진율: {{ totalBudget > 0 ? ((totalSpent / totalBudget) * 100).toFixed(1) : 0 }}%</span>
        </div>
      </div>

      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <span class="text-xs font-medium uppercase tracking-wider text-slate-500">총 검증 노출수 (Verified Impressions)</span>
        <p class="mt-2 text-2xl font-semibold text-slate-900 tabular-nums">{{ totalImpressions.toLocaleString() }}</p>
        <div class="mt-2 flex items-center gap-1.5 text-xs text-slate-500">
          <span>CPM 실시간 과금 반영</span>
        </div>
      </div>

      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <span class="text-xs font-medium uppercase tracking-wider text-slate-500">클릭수 &amp; 평균 클릭률 (CTR)</span>
        <p class="mt-2 text-2xl font-semibold text-slate-900 tabular-nums">
          {{ totalClicks.toLocaleString() }}
          <span class="text-sm font-normal text-slate-500 tabular-nums">({{ averageCtr.toFixed(2) }}%)</span>
        </p>
        <div class="mt-2 flex items-center gap-1.5 text-xs text-slate-500">
          <span>목적지 랜딩 유입 전환율</span>
        </div>
      </div>
    </div>

    <!-- Middle Section: Live Ad Preview Simulator & Active Campaigns -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Live Ad Preview Simulator -->
      <div class="rounded-lg border border-slate-200 bg-white p-6">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-semibold text-slate-900">광고 렌더링 시뮬레이터 (Ad Preview)</h2>
          <span class="rounded-md border border-slate-200 bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-600">Live Engine</span>
        </div>
        <p class="mt-1 text-xs text-slate-500">
          Test real creative rendering against Ad Decision Engine
        </p>

        <div class="mt-4 space-y-3">
          <div>
            <label class="text-xs font-medium text-slate-700">Slot ID</label>
            <select
              v-model="previewSlot"
              class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-xs text-slate-900 focus:border-slate-900 focus:bg-white focus:outline-none"
              @change="runAdPreview"
            >
              <option value="watch_preroll">watch_preroll (Main Player 5s Skippable)</option>
              <option value="home_feed_banner">home_feed_banner (Native Feed Card)</option>
              <option value="shorts_sponsored">shorts_sponsored (Shorts Vertical Feed)</option>
            </select>
          </div>

          <div>
            <label class="text-xs font-medium text-slate-700">Target Category</label>
            <select
              v-model="previewCategory"
              class="mt-1 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-xs text-slate-900 focus:border-slate-900 focus:bg-white focus:outline-none"
              @change="runAdPreview"
            >
              <option value="Technology">Technology &amp; Engineering</option>
              <option value="Gaming">Gaming &amp; Esports</option>
              <option value="Cinema">Cinema &amp; Entertainment</option>
            </select>
          </div>

          <div class="mt-4 rounded-lg border border-slate-200 bg-slate-50 p-4">
            <div v-if="isPreviewLoading" class="py-6 text-center text-xs text-slate-500 animate-pulse">
              의사결정 엔진 조회 중... (Querying Decision Engine)
            </div>
            <div v-else-if="previewResult?.ad_available && previewResult?.creative" class="space-y-3">
              <div class="flex items-center justify-between">
                <span class="rounded-md border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700">
                  MATCH SUCCESS
                </span>
                <span class="font-mono text-xs text-slate-500">{{ previewResult.creative.slotId }}</span>
              </div>

              <!-- Media Preview -->
              <div class="aspect-video w-full overflow-hidden rounded-lg border border-slate-200 bg-slate-100">
                <img
                  v-if="previewResult.creative.mediaUrl"
                  :src="previewResult.creative.mediaUrl"
                  :alt="previewResult.creative.title"
                  class="h-full w-full object-cover"
                />
              </div>

              <div>
                <h4 class="text-xs font-semibold text-slate-900">{{ previewResult.creative.title }}</h4>
                <p class="text-xs text-slate-600">Advertiser: {{ previewResult.creative.advertiserName }}</p>
                <div class="mt-2 flex items-center justify-between text-xs">
                  <span class="text-slate-500">CTA: {{ previewResult.creative.callToAction }}</span>
                  <a
                    :href="previewResult.creative.destinationUrl"
                    target="_blank"
                    class="font-medium text-slate-900 underline hover:text-slate-700"
                  >
                    목적지 URL
                  </a>
                </div>
              </div>
            </div>
            <div v-else class="py-6 text-center text-xs text-slate-600">
              미송출 (No-Fill): {{ previewResult?.reason || '적격 캠페인 없음' }}
            </div>
          </div>
        </div>
      </div>

      <!-- Active Campaigns Table -->
      <div class="overflow-hidden rounded-lg border border-slate-200 bg-white lg:col-span-2">
        <div class="p-6 pb-4">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-semibold text-slate-900">등록 캠페인 목록 (Ad Campaigns)</h2>
            <span class="text-xs text-slate-500 tabular-nums">{{ campaigns.length }} Campaigns Registered</span>
          </div>
          <p class="mt-1 text-xs text-slate-500">Real-time status, CPM/CPC bids and flight schedule</p>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left">
            <thead>
              <tr class="bg-slate-50 border-y border-slate-200 text-xs font-semibold text-slate-600 uppercase tracking-wider">
                <th class="py-2.5 px-4">캠페인명</th>
                <th class="py-2.5 px-4">형식</th>
                <th class="py-2.5 px-4">상태</th>
                <th class="py-2.5 px-4">입찰가 (CPM / CPC)</th>
                <th class="py-2.5 px-4">집행액</th>
                <th class="py-2.5 px-4">노출 / 클릭</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="c in campaigns" :key="c.id" class="hover:bg-slate-50">
                <td class="py-3 px-4 text-sm font-medium text-slate-900 border-b border-slate-100 max-w-[180px] truncate">
                  {{ c.name }}
                </td>
                <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
                  <span class="rounded-md bg-slate-100 px-2 py-0.5 font-mono text-xs uppercase text-slate-700">
                    {{ c.format }}
                  </span>
                </td>
                <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
                  <span
                    class="rounded-md px-2 py-0.5 text-xs font-medium uppercase"
                    :class="c.status === 'active' ? 'border border-emerald-200 bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-600'"
                  >
                    {{ c.status }}
                  </span>
                </td>
                <td class="py-3 px-4 font-mono text-sm text-slate-900 border-b border-slate-100 tabular-nums">
                  ₩{{ c.cpmBid?.toLocaleString() }} / ₩{{ c.cpcBid?.toLocaleString() }}
                </td>
                <td class="py-3 px-4 font-mono text-sm text-slate-900 border-b border-slate-100 tabular-nums">
                  ₩{{ (c.spentAmount || 0).toLocaleString() }}
                </td>
                <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
                  <span>{{ (c.impressions || 0).toLocaleString() }}</span>
                  <span class="text-slate-400"> / </span>
                  <span class="text-slate-600">{{ (c.clicks || 0).toLocaleString() }}</span>
                </td>
              </tr>
              <tr v-if="!campaigns.length">
                <td colspan="6" class="py-8 px-4 text-center text-sm text-slate-500">
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
