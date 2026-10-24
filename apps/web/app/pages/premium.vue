<script setup lang="ts">
import { SubscriptionTier } from '@gvp/shared-types'
import { useAppStore } from '~/stores/app'

const appStore = useAppStore()
const selectedPlan = ref<'individual' | 'family'>('individual')
const checkoutComplete = ref(false)

useSeoMeta({
  title: 'StreamVerse Premium — Ad-Free 4K HDR, Offline & Background Play',
  description: 'Upgrade to StreamVerse Premium for uninterrupted 4K HDR streaming, offline downloads, and spatial audio.'
})

function activatePremium() {
  if (appStore.currentUser) {
    appStore.currentUser.subscriptionTier =
      selectedPlan.value === 'family' ? SubscriptionTier.PREMIUM_FAMILY : SubscriptionTier.PREMIUM
  }
  checkoutComplete.value = true
}
</script>

<template>
  <div class="mx-auto max-w-5xl px-4 py-12">
    <div class="text-center">
      <span class="rounded-full bg-[#FF0033]/20 px-3.5 py-1 text-xs font-extrabold uppercase tracking-wider text-[#FF0033]">
        StreamVerse Premium
      </span>
      <h1 class="mt-4 text-3xl font-extrabold sm:text-4xl">
        {{ appStore.locale === 'ko' ? '광고 없는 4K HDR 스트리밍과 오프라인 저장' : 'Ad-Free 4K HDR, Offline Downloads & Background Play' }}
      </h1>
      <p class="mx-auto mt-3 max-w-2xl text-sm text-neutral-400">
        55% of your Premium subscription revenue goes directly to the creators you watch, tracked transparently on our double-entry ledger.
      </p>
    </div>

    <!-- Plan Cards -->
    <div class="mt-10 grid grid-cols-1 gap-6 md:grid-cols-2">
      <!-- Individual Plan -->
      <div
        class="cursor-pointer rounded-3xl border p-6 transition"
        :class="selectedPlan === 'individual' ? 'border-[#FF0033] bg-[#1C1215]' : 'border-white/10 bg-[#161616]'"
        @click="selectedPlan = 'individual'"
      >
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-bold">Individual Premium</h2>
          <span class="text-2xl font-extrabold text-[#3EA6FF]">{{ appStore.formatCurrency(14900) }} / mo</span>
        </div>
        <ul class="mt-5 space-y-2.5 text-xs text-neutral-300">
          <li>✓ Zero pre-roll, mid-roll, or feed video ads</li>
          <li>✓ 4K 2160p 60fps High Bitrate & Spatial Audio</li>
          <li>✓ Offline encrypted video downloads on Web & Mobile</li>
          <li>✓ Background audio playback on iOS & Android</li>
        </ul>
      </div>

      <!-- Family Plan -->
      <div
        class="cursor-pointer rounded-3xl border p-6 transition"
        :class="selectedPlan === 'family' ? 'border-[#FF0033] bg-[#1C1215]' : 'border-white/10 bg-[#161616]'"
        @click="selectedPlan = 'family'"
      >
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-bold">Premium Family (Up to 5 Members)</h2>
          <span class="text-2xl font-extrabold text-[#3EA6FF]">{{ appStore.formatCurrency(23900) }} / mo</span>
        </div>
        <ul class="mt-5 space-y-2.5 text-xs text-neutral-300">
          <li>✓ Share with up to 5 household members</li>
          <li>✓ Independent watch histories & recommendations</li>
          <li>✓ Simultaneous 4K streams across 5 devices</li>
          <li>✓ Includes StreamVerse Music & Exclusive Live VODs</li>
        </ul>
      </div>
    </div>

    <!-- Checkout Action -->
    <div class="mt-8 rounded-3xl border border-white/10 bg-[#161616] p-6 text-center">
      <div v-if="checkoutComplete" class="text-emerald-400">
        <p class="text-lg font-bold">🎉 Welcome to StreamVerse Premium!</p>
        <p class="mt-1 text-xs text-neutral-300">
          Your account tier is now <strong>{{ appStore.currentUser?.subscriptionTier.toUpperCase() }}</strong>. Ad-free 4K playback is active immediately.
        </p>
      </div>
      <div v-else>
        <button
          type="button"
          class="rounded-full bg-[#FF0033] px-8 py-3 text-sm font-bold text-white shadow-lg transition hover:bg-[#D9002B]"
          @click="activatePremium"
        >
          {{ appStore.locale === 'ko' ? '프리미엄 멤버십 즉시 활성화' : 'Activate Premium Membership Now' }}
        </button>
        <p class="mt-2 text-xs text-neutral-400">Cancel anytime · Instant VAT invoice issued</p>
      </div>
    </div>
  </div>
</template>
