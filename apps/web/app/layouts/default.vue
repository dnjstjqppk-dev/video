<script setup lang="ts">
import BrandSplash from '@gvp/ui/src/components/BrandSplash.vue'
import { BRAND } from '@gvp/design-tokens/src/brand'

const config = useRuntimeConfig()
const store = useAppStore()
const ko = computed(() => store.locale === 'ko')
const route = useRoute()
// Cinema pages stay dark in both themes; the footer follows so the edge
// between content and chrome does not flash light.
const isCinema = computed(() => route.path === '/ai' || route.path.startsWith('/series/'))
</script>

<template>
  <div class="flex min-h-screen flex-col bg-[var(--gvp-canvas)] text-[var(--gvp-text)]">
    <BrandSplash variant="viewer" />
    <AppHeader />
    <div class="flex flex-1">
      <AppSideDrawer />
      <main class="min-w-0 flex-1 pb-16 md:pb-0">
        <slot />
        <footer
          class="border-t px-4 pb-6 sm:px-6"
          :class="isCinema
            ? 'mt-0 border-white/10 bg-neutral-950 pt-[72px]'
            : 'mt-12 border-[var(--gvp-border)] bg-[var(--gvp-surface)] pt-6'"
        >
          <div class="mx-auto max-w-7xl">
            <nav
              class="flex flex-wrap items-center gap-x-6 gap-y-2 text-xs"
              :class="isCinema ? 'text-white/60' : 'text-[var(--gvp-text-secondary)]'"
              aria-label="서비스 정보"
            >
              <span class="font-bold" :class="isCinema ? 'text-white' : 'text-[var(--gvp-text)]'">{{ BRAND.name }} {{ BRAND.nameKo }}</span>
              <NuxtLink to="/terms" :class="isCinema ? 'hover:text-white hover:underline' : 'hover:text-[var(--gvp-text)] hover:underline'">이용약관</NuxtLink>
              <NuxtLink to="/privacy" :class="isCinema ? 'hover:text-white hover:underline' : 'hover:text-[var(--gvp-text)] hover:underline'">개인정보처리방침</NuxtLink>
              <NuxtLink to="/premium" :class="isCinema ? 'hover:text-white hover:underline' : 'hover:text-[var(--gvp-text)] hover:underline'">Premium</NuxtLink>
              <a :href="config.public.studioUrl" target="_blank" rel="noopener noreferrer" :class="isCinema ? 'hover:text-white hover:underline' : 'hover:text-[var(--gvp-text)] hover:underline'">{{ BRAND.studio.name }}</a>
            </nav>
            <p class="mt-2 text-xs" :class="isCinema ? 'text-white/40' : 'text-[var(--gvp-text-muted)]'">{{ ko ? BRAND.sloganKo : BRAND.sloganEn }}</p>
            <p class="tabular-nums mt-1 text-xs" :class="isCinema ? 'text-white/40' : 'text-[var(--gvp-text-muted)]'">© 2026 {{ BRAND.name }}. {{ ko ? '계정 삭제 요청 시 30일 유예 후 파기됩니다.' : 'Data is purged 30 days after deletion requests.' }}</p>
          </div>
        </footer>
      </main>
    </div>
    <AppBottomTabs />
  </div>
</template>
