<script setup lang="ts">
const store = useAppStore()
const config = useRuntimeConfig()
const theme = useThemeMode()
const open = ref(false)
const modes = computed(() => [
  { id: 'viewer', label: '시청자 모드', href: '/', current: true },
  { id: 'studio', label: '파트너 스튜디오', href: config.public.studioUrl as string, current: false },
  { id: 'admin', label: '종합 관리자 콘솔', href: config.public.adminUrl as string, current: false },
])
async function logout() {
  open.value = false
  try { await $fetch('/auth/logout', { method: 'POST' }) } catch { /* cleared server-side */ }
  useCookie('gvp_access_token').value = null
  store.clearSession()
  await navigateTo('/')
}
</script>

<template>
  <div v-click-outside="() => (open = false)" class="relative">
    <button v-if="store.currentUser" type="button" :aria-expanded="open" aria-label="프로필 및 워크스페이스"
      class="flex items-center rounded-full border border-[var(--gvp-border-strong)] p-0.5 hover:border-[var(--gvp-text-muted)]" @click="open = !open">
      <img :src="store.currentUser?.avatarUrl" :alt="store.currentUser?.displayName" class="h-7 w-7 rounded-full object-cover" />
    </button>
    <NuxtLink v-else to="/login" class="inline-flex h-8 items-center rounded-md border border-[var(--gvp-border-strong)] px-3 text-xs font-medium text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]">로그인</NuxtLink>
    <div v-if="open && store.currentUser" role="menu" class="gvp-dropdown gvp-elevated absolute right-0 top-full z-50 mt-2 w-72 rounded-xl p-1.5">
      <div class="border-b border-[var(--gvp-border)] px-2.5 py-2">
        <p class="text-sm font-semibold text-[var(--gvp-text)]">{{ store.currentUser?.displayName }}</p>
        <p class="tabular-nums text-xs text-[var(--gvp-text-secondary)]">@{{ store.currentUser?.username }} · {{ store.currentUser?.subscriptionTier?.toUpperCase() }}</p>
      </div>
      <p class="px-2.5 pb-1 pt-2 text-[11px] font-semibold text-[var(--gvp-text-muted)]">워크스페이스 전환</p>
      <a v-for="m in modes" :key="m.id" :href="m.href" :target="m.current ? undefined : '_blank'" rel="noopener noreferrer"
        class="flex items-center justify-between rounded-lg px-2.5 py-2 text-sm hover:bg-[var(--gvp-elevated)]" :class="m.current ? 'font-semibold text-[var(--gvp-text)]' : 'text-[var(--gvp-text-secondary)]'" @click="open = false">
        {{ m.label }}<span v-if="m.current" class="h-2 w-2 rounded-full bg-[var(--gvp-accent)]" />
      </a>
      <div class="my-1 border-t border-[var(--gvp-border)]" />
      <button type="button" class="flex w-full items-center justify-between rounded-lg px-2.5 py-2 text-sm text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]" @click="theme.toggle()">
        테마: {{ theme.isDark.value ? '다크' : '라이트' }}<span>{{ theme.isDark.value ? '◐' : '○' }}</span>
      </button>
      <button type="button" class="flex w-full items-center justify-between rounded-lg px-2.5 py-2 text-sm text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]" @click="store.toggleLocale(); open = false">
        언어: {{ store.locale === 'ko' ? '한국어' : 'English' }}<span>›</span>
      </button>
      <NuxtLink to="/premium" class="block rounded-lg px-2.5 py-2 text-sm font-medium text-[var(--gvp-brand)] hover:bg-[var(--gvp-elevated)]" @click="open = false">StreamVerse Premium</NuxtLink>
      <button type="button" class="w-full rounded-lg px-2.5 py-2 text-left text-sm text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]" @click="logout()">로그아웃</button>
    </div>
  </div>
</template>
