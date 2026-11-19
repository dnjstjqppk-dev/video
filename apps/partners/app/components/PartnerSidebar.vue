<script setup lang="ts">
defineProps<{ collapsed: boolean; mobileOpen: boolean; orgName: string; orgId: string }>()
const route = useRoute()

const navItems = [
  { path: '/', label: '종합 현황', icon: 'M3 3h7v7H3V3zm11 0h7v7h-7V3zm0 11h7v7h-7v-7zM3 14h7v7H3v-7z' },
  { path: '/advertiser', label: '광고주', icon: 'M12 2L2 7l10 5 10-5-10-5zm0 15l-10-5v5l10 5 10-5v-5l-10 5z' },
  { path: '/campaigns', label: '캠페인', icon: 'M3 11l18-5v12L3 14v-3zm0 3l4 6h2l-3-6' },
  { path: '/mcn', label: 'MCN', icon: 'M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2m8-10a4 4 0 100-8 4 4 0 000 8zm14 10v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75' },
  { path: '/rights', label: '저작권', icon: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z' },
  { path: '/settlements', label: '정산', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9h6m-6 4h6' }
]
</script>

<template>
  <aside
    :class="[mobileOpen ? 'translate-x-0' : '-translate-x-full', collapsed ? 'md:w-[76px] md:px-2' : 'md:w-60 md:px-3']"
    class="fixed inset-y-0 left-0 z-50 w-64 shrink-0 overflow-y-auto border-r border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-3 transition-all md:static md:z-auto md:translate-x-0"
  >
    <div class="border-b border-[var(--gvp-border)] py-4" :class="collapsed ? 'md:px-0' : 'px-1'">
      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-3" :class="collapsed ? 'md:border-0 md:bg-transparent md:p-0 md:text-center' : ''">
        <p class="text-[10px] font-semibold uppercase tracking-wider text-[var(--gvp-brand-hover)]" :class="collapsed ? 'md:hidden' : ''">Organization</p>
        <p class="mt-1 truncate text-xs font-semibold text-[var(--gvp-text)]" :class="collapsed ? 'md:hidden' : ''" :title="orgName">{{ orgName || '조직 정보를 불러오는 중' }}</p>
        <p v-if="orgId" class="mt-0.5 truncate font-mono text-[11px] text-[var(--gvp-text-secondary)]" :class="collapsed ? 'md:hidden' : ''">ID: {{ orgId }}</p>
      </div>
    </div>
    <nav class="space-y-0.5 py-3" aria-label="파트너 메뉴">
      <NuxtLink
        v-for="item in navItems"
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
