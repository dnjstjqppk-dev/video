<script setup lang="ts">
import { useAppStore } from '~/stores/app'

const appStore = useAppStore()
const config = useRuntimeConfig()
const step = ref(0)

useSeoMeta({
  title: '시작하기',
  description: 'StreamVerse 시작 가이드: 시청, 구독, 업로드, Premium까지 3단계로 안내합니다.'
})

const steps = computed(() => [
  {
    title: appStore.locale === 'ko' ? '관심사를 골라 피드를 맞추세요' : 'Pick interests to tune your feed',
    body: appStore.locale === 'ko'
      ? '홈 상단 카테고리 칩에서 기술·음악·게임 등을 고르면 추천이 즉시 바뀝니다. 관심 없음 피드백도 바로 반영됩니다.'
      : 'Choose Technology, Music, or Gaming chips on Home and recommendations adapt instantly.',
    cta: appStore.locale === 'ko' ? '홈 피드 보기' : 'Browse Home',
    to: '/'
  },
  {
    title: appStore.locale === 'ko' ? '채널을 구독하고 알림을 받으세요' : 'Subscribe and never miss uploads',
    body: appStore.locale === 'ko'
      ? '좋아하는 크리에이터를 구독하면 새 영상·라이브 소식을 모아볼 수 있습니다. 로그인이 필요합니다.'
      : 'Subscribe to creators to collect new videos and live alerts. Sign-in required.',
    cta: appStore.locale === 'ko' ? '인기 채널 보기' : 'Explore channels',
    to: '/feed/subscriptions'
  },
  {
    title: appStore.locale === 'ko' ? '크리에이터라면 스튜디오에서 시작' : 'Creators start in Studio',
    body: appStore.locale === 'ko'
      ? '스튜디오에서 업로드·분석·수익 정산을 한곳에서 관리하세요. 광고 없는 감상은 Premium으로 업그레이드할 수 있습니다.'
      : 'Upload, analytics, and earnings live in Studio. Upgrade to Premium for ad-free viewing.',
    cta: 'Creator Studio',
    href: config.public.studioUrl as string
  }
])

function next() {
  if (step.value < steps.value.length - 1) step.value += 1
}

function back() {
  if (step.value > 0) step.value -= 1
}

onMounted(() => {
  try {
    localStorage.setItem('gvp_onboarded', '1')
  } catch {
    // Private mode: onboarding simply shows again next visit.
  }
})
</script>

<template>
  <div class="mx-auto max-w-2xl px-4 py-16 sm:px-6">
    <p class="text-xs font-bold uppercase tracking-wider text-[#FF0033]">Welcome</p>
    <h1 class="mt-2 text-3xl font-extrabold">
      {{ appStore.locale === 'ko' ? 'StreamVerse 시작하기' : 'Get started with StreamVerse' }}
    </h1>

    <ol class="mt-6 flex items-center gap-2" aria-label="진행 단계">
      <li
        v-for="(_, i) in steps"
        :key="i"
        class="h-1.5 flex-1 rounded-full"
        :class="i <= step ? 'bg-[#FF0033]' : 'bg-white/10'"
        :aria-current="i === step ? 'step' : undefined"
      />
    </ol>

    <div class="mt-6 rounded-2xl border border-white/10 bg-[#161616] p-6" role="status">
      <p class="text-xs font-bold text-[#3EA6FF]">Step {{ step + 1 }} / {{ steps.length }}</p>
      <h2 class="mt-1 text-xl font-bold">{{ steps[step]?.title }}</h2>
      <p class="mt-2 text-sm text-neutral-300">{{ steps[step]?.body }}</p>
      <div class="mt-5 flex flex-wrap items-center gap-3">
        <NuxtLink
          v-if="steps[step]?.to"
          :to="steps[step]?.to"
          class="rounded-full bg-[#FF0033] px-6 py-2.5 text-sm font-bold text-white hover:bg-[#D9002B]"
        >
          {{ steps[step]?.cta }}
        </NuxtLink>
        <a
          v-else-if="steps[step]?.href"
          :href="steps[step]?.href"
          target="_blank"
          rel="noopener noreferrer"
          class="rounded-full bg-[#FF0033] px-6 py-2.5 text-sm font-bold text-white hover:bg-[#D9002B]"
        >
          {{ steps[step]?.cta }} ↗
        </a>
        <div class="flex gap-2">
          <button
            v-if="step > 0"
            type="button"
            class="rounded-full border border-white/15 px-5 py-2.5 text-sm font-bold text-neutral-200 hover:bg-white/10"
            @click="back"
          >
            {{ appStore.locale === 'ko' ? '이전' : 'Back' }}
          </button>
          <button
            v-if="step < steps.length - 1"
            type="button"
            class="rounded-full border border-white/15 px-5 py-2.5 text-sm font-bold text-neutral-200 hover:bg-white/10"
            @click="next"
          >
            {{ appStore.locale === 'ko' ? '다음' : 'Next' }}
          </button>
          <NuxtLink
            v-else
            to="/premium"
            class="rounded-full border border-[#3EA6FF]/50 px-5 py-2.5 text-sm font-bold text-[#3EA6FF] hover:bg-[#3EA6FF]/10"
          >
            Premium 보기
          </NuxtLink>
        </div>
      </div>
    </div>

    <p class="mt-4 text-center text-xs text-neutral-500">
      <NuxtLink to="/terms" class="underline hover:text-neutral-300">이용약관</NuxtLink>
      ·
      <NuxtLink to="/privacy" class="underline hover:text-neutral-300">개인정보처리방침</NuxtLink>
    </p>
  </div>
</template>
