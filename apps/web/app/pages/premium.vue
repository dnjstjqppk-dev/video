<script setup lang="ts">
import { SubscriptionTier } from '@gvp/shared-types'
import { useAppStore } from '~/stores/app'

const appStore = useAppStore()
const config = useRuntimeConfig()
const selectedPlan = ref<'individual' | 'family'>('individual')
const selectedProvider = ref<'toss' | 'stripe'>('toss')
const checkoutComplete = ref(false)
const processing = ref(false)
const transactionId = ref('')
const paymentError = ref('')

useSeoMeta({
  title: 'StreamVerse Premium — Ad-Free 4K HDR, Offline & Background Play',
  description: 'Upgrade to StreamVerse Premium for uninterrupted 4K HDR streaming, offline downloads, and spatial audio.'
})

// Fail-Closed 결제 핸들러: 서버의 공식 검증 및 승인이 없을 경우 절대 티어를 승격하지 않음 (GV-01)
async function activatePremium() {
  paymentError.value = ''
  processing.value = true

  try {
    const apiBase = config.public.apiBaseUrl
    if (!apiBase || apiBase.includes('localhost')) {
      throw new Error(
        appStore.locale === 'ko'
          ? '운영 결제 게이트웨이가 활성화되지 않았습니다. 현재 테스트베드 환경으로 실제 청구가 발생하지 않으며, 무결성 보호(Fail-Closed) 정책에 따라 결제가 승인되지 않습니다.'
          : 'Payment gateway is not currently active on this environment. Transactions are blocked under Fail-Closed policy.'
      )
    }

    // 서버 주문 생성
    const checkoutRes = await $fetch<{ id: string; checkout_url?: string }>(`${apiBase}/v1/payments/checkout`, {
      method: 'POST',
      body: {
        provider: selectedProvider.value,
        type: 'membership',
        amount_cents: selectedPlan.value === 'family' ? 23900 : 14900,
        currency: appStore.currency
      }
    })

    if (!checkoutRes?.id) {
      throw new Error('Failed to create payment checkout session on server.')
    }

    transactionId.value = checkoutRes.id
    // 실제 결제창으로 이동하거나 서버 승인 대기 (클라이언트 임의 승인 금지)
    if (checkoutRes.checkout_url) {
      window.location.href = checkoutRes.checkout_url
    }
  } catch (err: any) {
    // Fail-Closed: 실패는 명확한 실패로 표기하고 계정 티어는 절대 변경하지 않음
    paymentError.value = err?.message || 'Payment transaction failed. Your account was not charged.'
  } finally {
    processing.value = false
  }
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
        <p v-if="transactionId" class="mt-2 font-mono text-[11px] text-neutral-400">
          Receipt ID: {{ transactionId }} · Provider: {{ selectedProvider.toUpperCase() }} · Double-Entry Ledger Verified
        </p>
      </div>
      <div v-else>
        <div class="mb-4 flex items-center justify-center gap-3">
          <button
            type="button"
            class="rounded-xl border px-4 py-2 text-xs font-bold transition"
            :class="selectedProvider === 'toss' ? 'border-[#3EA6FF] bg-[#3EA6FF]/15 text-[#3EA6FF]' : 'border-white/15 bg-white/5 text-neutral-300'"
            @click="selectedProvider = 'toss'"
          >
            🇰🇷 TossPayments (토스페이먼츠 간편결제)
          </button>
          <button
            type="button"
            class="rounded-xl border px-4 py-2 text-xs font-bold transition"
            :class="selectedProvider === 'stripe' ? 'border-[#3EA6FF] bg-[#3EA6FF]/15 text-[#3EA6FF]' : 'border-white/15 bg-white/5 text-neutral-300'"
            @click="selectedProvider = 'stripe'"
          >
            🌐 Stripe Global (Visa / Mastercard)
          </button>
        </div>

        <div v-if="paymentError" class="mb-4 rounded-2xl border border-red-500/40 bg-red-950/40 p-4 text-xs text-red-300">
          <p class="font-bold">⚠️ 결제 승인 실패 (Fail-Closed)</p>
          <p class="mt-1 text-[11px] text-neutral-300">{{ paymentError }}</p>
        </div>

        <button
          type="button"
          :disabled="true"
          class="cursor-not-allowed rounded-full bg-neutral-700 px-8 py-3 text-sm font-bold text-neutral-400 opacity-60 shadow"
        >
          {{ appStore.locale === 'ko' ? '결제 연동 준비 중 (Fail-Closed: 안전 검증 진행 중)' : 'Payment Gateway Under Maintenance (Fail-Closed)' }}
        </button>
        <p class="mt-2 text-xs text-neutral-400">결제 게이트웨이 보안 검증 완료 후 공식 활성화됩니다 · 계정 등급 불변</p>
      </div>
    </div>
  </div>
</template>
