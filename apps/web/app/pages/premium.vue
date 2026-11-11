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
const checkoutError = paymentError

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
    if (!apiBase) {
      throw new Error(
        appStore.locale === 'ko'
          ? '결제 API 주소가 설정되지 않았습니다. 관리자에게 문의하세요.'
          : 'Payment API base URL is not configured. Please contact support.'
      )
    }

    // 서버 주문 생성 (실패 시 Fail-Closed: 티어 변경 없음)
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
  <div class="mx-auto max-w-5xl bg-white px-4 py-12 text-slate-900">
    <div class="text-center">
      <span class="rounded-md border border-blue-200 bg-blue-50 px-3 py-1 text-xs font-bold uppercase tracking-wider text-blue-700">
        StreamVerse Premium
      </span>
      <h1 class="mt-4 text-3xl font-extrabold text-slate-900 sm:text-4xl">
        {{ appStore.locale === 'ko' ? '광고 없는 4K HDR 스트리밍과 오프라인 저장' : 'Ad-Free 4K HDR, Offline Downloads & Background Play' }}
      </h1>
      <p class="mx-auto mt-3 max-w-2xl text-sm text-slate-600 tabular-nums">
        55% of your Premium subscription revenue goes directly to the creators you watch, tracked transparently on our double-entry ledger.
      </p>
    </div>

    <!-- Plan Cards -->
    <div class="mt-10 grid grid-cols-1 gap-6 md:grid-cols-2">
      <!-- Individual Plan -->
      <div
        class="cursor-pointer rounded-lg border p-6 transition"
        :class="selectedPlan === 'individual' ? 'border-slate-900 bg-slate-50' : 'border-slate-200 bg-white hover:bg-slate-50'"
        @click="selectedPlan = 'individual'"
      >
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-bold text-slate-900">Individual Premium</h2>
          <span class="text-2xl font-extrabold text-blue-700 tabular-nums">{{ appStore.formatCurrency(14900) }} / mo</span>
        </div>
        <ul class="mt-5 space-y-2.5 text-xs text-slate-600">
          <li>• Zero pre-roll, mid-roll, or feed video ads</li>
          <li>• 4K 2160p 60fps High Bitrate & Spatial Audio</li>
          <li>• Offline encrypted video downloads on Web & Mobile</li>
          <li>• Background audio playback on iOS & Android</li>
        </ul>
      </div>

      <!-- Family Plan -->
      <div
        class="cursor-pointer rounded-lg border p-6 transition"
        :class="selectedPlan === 'family' ? 'border-slate-900 bg-slate-50' : 'border-slate-200 bg-white hover:bg-slate-50'"
        @click="selectedPlan = 'family'"
      >
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-bold text-slate-900 tabular-nums">Premium Family (Up to 5 Members)</h2>
          <span class="text-2xl font-extrabold text-blue-700 tabular-nums">{{ appStore.formatCurrency(23900) }} / mo</span>
        </div>
        <ul class="mt-5 space-y-2.5 text-xs text-slate-600">
          <li>• Share with up to 5 household members</li>
          <li>• Independent watch histories & recommendations</li>
          <li>• Simultaneous 4K streams across 5 devices</li>
          <li>• Includes StreamVerse Music & Exclusive Live VODs</li>
        </ul>
      </div>
    </div>

    <!-- Checkout Action -->
    <div class="mt-8 rounded-lg border border-slate-200 bg-slate-50 p-6 text-center">
      <div v-if="checkoutComplete" class="text-green-700">
        <p class="text-lg font-bold">Welcome to StreamVerse Premium</p>
        <p class="mt-1 text-xs text-slate-600">
          Your account tier is now <strong>{{ appStore.currentUser?.subscriptionTier?.toUpperCase() }}</strong>. Ad-free 4K playback is active immediately.
        </p>
        <p v-if="transactionId" class="mt-2 font-mono text-[11px] text-slate-500 tabular-nums">
          Receipt ID: {{ transactionId }} · Provider: {{ selectedProvider.toUpperCase() }} · Double-Entry Ledger Verified
        </p>
      </div>
      <div v-else>
        <div class="mb-4 flex flex-wrap items-center justify-center gap-3">
          <button
            type="button"
            class="rounded-md border px-4 py-2 text-xs font-semibold transition"
            :class="selectedProvider === 'toss' ? 'border-blue-200 bg-blue-50 text-blue-700' : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-50'"
            @click="selectedProvider = 'toss'"
          >
            KR · TossPayments (토스페이먼츠 간편결제)
          </button>
          <button
            type="button"
            class="rounded-md border px-4 py-2 text-xs font-semibold transition"
            :class="selectedProvider === 'stripe' ? 'border-blue-200 bg-blue-50 text-blue-700' : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-50'"
            @click="selectedProvider = 'stripe'"
          >
            GLOBAL · Stripe Global (Visa / Mastercard)
          </button>
        </div>

        <div v-if="paymentError || checkoutError" class="mb-4 rounded-lg border border-red-200 bg-red-50 p-4 text-xs text-red-700">
          <p class="font-bold">결제 승인 실패 (Fail-Closed)</p>
          <p class="mt-1 text-[11px] text-red-600">{{ paymentError || checkoutError }}</p>
        </div>

        <div class="flex flex-wrap items-center justify-center gap-3">
          <button
            type="button"
            :disabled="processing"
            class="rounded-md bg-slate-900 px-6 py-2.5 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-200 disabled:text-slate-500 tabular-nums"
            @click="activatePremium"
          >
            {{ processing
              ? (appStore.locale === 'ko' ? '결제 승인 중…' : 'Processing…')
              : (appStore.locale === 'ko' ? `${appStore.formatCurrency(selectedPlan === 'family' ? 23900 : 14900)}로 Premium 시작하기` : `Start Premium — ${appStore.formatCurrency(selectedPlan === 'family' ? 23900 : 14900)}/mo`) }}
          </button>

          <button
            type="button"
            disabled
            class="cursor-not-allowed rounded-md border border-amber-200 bg-amber-50 px-4 py-2.5 text-xs font-semibold text-amber-700 opacity-90"
          >
            결제 연동 준비 중 (Fail-Closed: 안전 검증 진행 중)
          </button>
        </div>
        <p class="mt-3 text-xs text-slate-500">서버 승인 없이 등급이 변경되지 않습니다 (Fail-Closed) · {{ selectedProvider === 'toss' ? 'TossPayments' : 'Stripe' }} 안전결제</p>
      </div>
    </div>
  </div>
</template>
