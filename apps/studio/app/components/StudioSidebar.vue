<script setup lang="ts">
import type { UserProfile } from '@gvp/shared-types'
import GvpChannelAvatar from '@gvp/ui/src/components/GvpChannelAvatar.vue'

defineProps<{ collapsed: boolean; mobileOpen: boolean; sessionUser: UserProfile | null }>()
const route = useRoute()

const navLinks = [
  { path: '/', label: '대시보드', icon: 'M3 3h7v7H3V3zm11 0h7v7h-7V3zm0 11h7v7h-7v-7zM3 14h7v7H3v-7z' },
  { path: '/content', label: '콘텐츠', icon: 'M4 6h16v12H4V6zm2 2v8h12V8H6zm4 2l5 2-5 2v-4z' },
  { path: '/analytics', label: '분석', icon: 'M3 20h18M6 16v-4m4 4V8m4 8v-6m4 6V4' },
  { path: '/comments', label: '댓글', icon: 'M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2v10z' },
  { path: '/subtitles', label: '자막', icon: 'M4 6h16v12H4V6zm3 5h4m2 0h4m-10 3h6m2 0h2' },
  { path: '/copyright', label: '저작권', icon: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z' },
  { path: '/monetization', label: '수익 창출', icon: 'M12 2v20M17 5H9.5a3.5 3.5 0 000 7h5a3.5 3.5 0 010 7H6' },
  { path: '/customization', label: '맞춤설정', icon: 'M12 20h9M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z' },
  { path: '/upload', label: '업로드', icon: 'M12 16V4m0 0L8 8m4-4l4 4M4 17v2a1 1 0 001 1h14a1 1 0 001-1v-2' },
  { path: '/live', label: '라이브', icon: 'M12 8a4 4 0 100 8 4 4 0 000-8zm-7 4a7 7 0 0114 0' }
]
</script>

<template>
  <aside
    :class="[mobileOpen ? 'translate-x-0' : '-translate-x-full', collapsed ? 'md:w-[76px] md:px-2' : 'md:w-60 md:px-3']"
    class="fixed inset-y-0 left-0 z-50 w-64 shrink-0 overflow-y-auto border-r border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-3 transition-all md:static md:z-auto md:translate-x-0"
  >
    <div class="flex items-center gap-3 border-b border-[var(--gvp-border)] py-4" :class="collapsed ? 'md:justify-center md:px-0' : 'px-1'">
      <GvpChannelAvatar :src="sessionUser?.avatarUrl" :name="sessionUser?.displayName || 'S'" :size-px="36" />
      <div class="min-w-0" :class="collapsed ? 'md:hidden' : ''">
        <p class="truncate text-xs font-semibold text-[var(--gvp-text)]">{{ sessionUser?.displayName || '미로그인 크리에이터' }}</p>
        <p class="truncate text-[11px] text-[var(--gvp-text-secondary)]">{{ sessionUser ? `@${sessionUser.username}` : '로그인 후 채널 표시' }}</p>
      </div>
    </div>
    <nav class="space-y-0.5 py-3" aria-label="스튜디오 메뉴">
      <NuxtLink
        v-for="item in navLinks"
        :key="item.path"
        :to="item.path"
        :title="item.label"
        class="relative flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors"
        :class="[
          route.path === item.path
            ? 'bg-[var(--gvp-elevated)] font-semibold text-[var(--gvp-text)]'
            : 'text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]',
          collapsed ? 'md:justify-center md:px-0' : ''
        ]"
      >
        <span v-if="route.path === item.path" class="absolute left-0 top-1.5 h-6 w-1 rounded-r bg-[var(--gvp-brand)]" aria-hidden="true" />
        <svg class="h-[18px] w-[18px] shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" :d="item.icon" />
        </svg>
        <span class="truncate" :class="collapsed ? 'md:hidden' : ''">{{ item.label }}</span>
      </NuxtLink>
    </nav>
  </aside>
</template>
