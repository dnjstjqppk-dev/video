<script setup lang="ts">
import type { UserProfile } from '@gvp/shared-types'
import VravraLogo from '@gvp/ui/src/components/VravraLogo.vue'

defineProps<{ sessionUser: UserProfile | null }>()
defineEmits<{ (e: 'toggle-drawer'): void; (e: 'toggle-collapse'): void; (e: 'logout'): void }>()
const config = useRuntimeConfig()
const theme = useConsoleTheme()
const searchTerm = ref('')

function submitSearch() {
  const q = searchTerm.value.trim()
  navigateTo(q ? { path: '/content', query: { q } } : '/content')
}
</script>

<template>
  <header class="sticky top-0 z-40 flex h-14 items-center gap-2 border-b border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-3 sm:px-4">
    <button type="button" aria-label="메뉴 열기" class="inline-flex h-9 w-9 items-center justify-center rounded-full hover:bg-[var(--gvp-elevated)] md:hidden" @click="$emit('toggle-drawer')">
      <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path stroke-linecap="round" d="M4 7h16M4 12h16M4 17h16" /></svg>
    </button>
    <button type="button" aria-label="사이드바 접기/펼치기" class="hidden h-9 w-9 items-center justify-center rounded-full hover:bg-[var(--gvp-elevated)] md:inline-flex" @click="$emit('toggle-collapse')">
      <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path stroke-linecap="round" d="M4 7h16M4 12h16M4 17h16" /></svg>
    </button>
    <NuxtLink to="/" class="flex shrink-0 items-center gap-2" aria-label="스튜디오 홈">
      <VravraLogo :size="26" suffix="studio" suffix-tone="studio" />
    </NuxtLink>
    <form class="mx-auto hidden w-full max-w-xl items-center md:flex" role="search" @submit.prevent="submitSearch">
      <input v-model="searchTerm" type="search" placeholder="내 콘텐츠 검색" aria-label="내 콘텐츠 검색" class="h-9 w-full rounded-full border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-4 text-sm text-[var(--gvp-text)] outline-none placeholder:text-[var(--gvp-text-muted)] focus:border-[var(--gvp-focus)]" />
    </form>
    <div class="ml-auto flex shrink-0 items-center gap-1.5 md:ml-0">
      <a :href="config.public.webUrl as string" target="_blank" rel="noopener noreferrer" aria-label="vravra 메인으로 돌아가기" title="vravra 메인으로 돌아가기" class="inline-flex h-9 w-9 items-center justify-center rounded-full text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]">
        <svg class="h-[18px] w-[18px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M3 10.5L12 3l9 7.5V21a1 1 0 0 1-1 1h-5.5v-6.5h-5V22H4a1 1 0 0 1-1-1V10.5z" /></svg>
      </a>
      <button type="button" :aria-label="theme.mode.value === 'dark' ? '라이트 모드로 전환' : '다크 모드로 전환'" class="inline-flex h-9 w-9 items-center justify-center rounded-full text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]" @click="theme.toggle()">
        {{ theme.mode.value === 'dark' ? '○' : '◐' }}
      </button>
      <StudioCreateMenu v-if="sessionUser" />
      <StudioProfileMenu :session-user="sessionUser" @logout="$emit('logout')" />
    </div>
  </header>
</template>
