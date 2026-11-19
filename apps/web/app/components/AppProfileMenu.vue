<script setup lang="ts">
const store = useAppStore()
const config = useRuntimeConfig()
const api = useApi()
const theme = useThemeMode()
const open = ref(false)
const triedChannel = ref(false)
const themeOpts = [
  { id: 'light', label: '라이트' },
  { id: 'dark', label: '다크' },
  { id: 'system', label: '시스템' },
] as const
const isAdmin = computed(() => (store.currentUser?.role || '').includes('admin'))
watch(open, (v) => {
  if (v && store.currentUser && !triedChannel.value) {
    triedChannel.value = true
    void store.fetchMyChannel(api)
  }
})
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
      <NuxtLink v-if="store.myChannelId" :to="`/channels/${store.myChannelId}`" class="mt-1 block rounded-lg px-2.5 py-2 text-sm font-medium text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]" @click="open = false">내 채널 보기</NuxtLink>
      <a :href="config.public.studioUrl as string" target="_blank" rel="noopener noreferrer" class="block rounded-lg px-2.5 py-2 text-sm font-medium text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]" @click="open = false">vravra studio 바로가기 ↗</a>
      <a v-if="isAdmin" :href="config.public.adminUrl as string" target="_blank" rel="noopener noreferrer" class="block rounded-lg px-2.5 py-2 text-sm text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]" @click="open = false">관리자 콘솔 ↗</a>
      <div class="my-1 border-t border-[var(--gvp-border)]" />
      <p class="px-2.5 pb-1 pt-1 text-[11px] font-semibold text-[var(--gvp-text-muted)]">테마</p>
      <div class="flex gap-1 px-2.5 pb-1" role="group" aria-label="테마 선택">
        <button v-for="o in themeOpts" :key="o.id" type="button" :aria-pressed="theme.preference.value === o.id"
          :class="theme.preference.value === o.id ? 'bg-[var(--gvp-text)] text-[var(--gvp-canvas)]' : 'bg-[var(--gvp-elevated)] text-[var(--gvp-text-secondary)]'"
          class="h-7 flex-1 rounded-md text-xs font-medium" @click="theme.apply(o.id)">{{ o.label }}</button>
      </div>
      <button type="button" class="flex w-full items-center justify-between rounded-lg px-2.5 py-2 text-sm text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]" @click="store.toggleLocale(); open = false">
        언어: {{ store.locale === 'ko' ? '한국어' : 'English' }}<span>›</span>
      </button>
      <NuxtLink to="/premium" class="block rounded-lg px-2.5 py-2 text-sm font-medium text-[var(--gvp-brand)] hover:bg-[var(--gvp-elevated)]" @click="open = false">vravra Premium</NuxtLink>
      <button type="button" class="w-full rounded-lg px-2.5 py-2 text-left text-sm text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]" @click="logout()">로그아웃</button>
    </div>
  </div>
</template>
