<script setup lang="ts">
import type { UserProfile } from '@gvp/shared-types'
import VravraLogo from '@gvp/ui/src/components/VravraLogo.vue'

defineProps<{ sessionUser: UserProfile | null; orgName: string }>()
defineEmits<{ (e: 'toggle-drawer'): void; (e: 'toggle-collapse'): void; (e: 'logout'): void }>()
const theme = useConsoleTheme()
</script>

<template>
  <header class="sticky top-0 z-40 flex h-14 items-center gap-2 border-b border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-3 sm:px-4">
    <button type="button" aria-label="메뉴 열기" class="inline-flex h-9 w-9 items-center justify-center rounded-full hover:bg-[var(--gvp-elevated)] md:hidden" @click="$emit('toggle-drawer')">
      <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path stroke-linecap="round" d="M4 7h16M4 12h16M4 17h16" /></svg>
    </button>
    <button type="button" aria-label="사이드바 접기/펼치기" class="hidden h-9 w-9 items-center justify-center rounded-full hover:bg-[var(--gvp-elevated)] md:inline-flex" @click="$emit('toggle-collapse')">
      <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path stroke-linecap="round" d="M4 7h16M4 12h16M4 17h16" /></svg>
    </button>
    <NuxtLink to="/" class="flex shrink-0 items-center gap-2" aria-label="파트너 홈">
      <VravraLogo size="sm" />
    </NuxtLink>
    <span v-if="orgName" class="hidden min-w-0 truncate rounded-md border border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] px-2 py-0.5 text-xs font-medium text-[var(--gvp-brand-hover)] sm:inline" :title="orgName">{{ orgName }}</span>
    <div class="ml-auto flex shrink-0 items-center gap-1.5">
      <button type="button" :aria-label="theme.mode.value === 'dark' ? '라이트 모드로 전환' : '다크 모드로 전환'" class="inline-flex h-9 w-9 items-center justify-center rounded-full text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]" @click="theme.toggle()">
        {{ theme.mode.value === 'dark' ? '○' : '◐' }}
      </button>
      <PartnerProfileMenu :session-user="sessionUser" @logout="$emit('logout')" />
    </div>
  </header>
</template>
