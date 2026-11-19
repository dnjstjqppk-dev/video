<script setup lang="ts">
import type { UserProfile } from '@gvp/shared-types'
import GvpChannelAvatar from '@gvp/ui/src/components/GvpChannelAvatar.vue'

defineProps<{ sessionUser: UserProfile | null }>()
defineEmits<{ (e: 'logout'): void }>()
const config = useRuntimeConfig()
const route = useRoute()
const open = ref(false)
const currency = useState<'KRW' | 'USD'>('studio-currency', () => 'KRW')
watch(() => route.path, () => { open.value = false })

const consoles = [
  { label: '시청자 포털', href: () => config.public.webUrl as string },
  { label: '파트너 콘솔', href: () => config.public.partnersUrl as string },
  { label: '관리자 콘솔', href: () => config.public.adminUrl as string }
]
</script>

<template>
  <div class="relative">
    <button type="button" class="rounded-full" :aria-expanded="open" aria-haspopup="menu" aria-label="프로필 메뉴" @click="open = !open" @keydown.escape="open = false">
      <GvpChannelAvatar :src="sessionUser?.avatarUrl" :name="sessionUser?.displayName || '?'" :size-px="32" />
    </button>
    <div v-if="open" class="absolute right-0 z-50 mt-2 w-64 overflow-hidden rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] py-2 shadow-lg" role="menu">
      <div class="border-b border-[var(--gvp-border)] px-4 pb-3 pt-1">
        <p class="truncate text-sm font-semibold text-[var(--gvp-text)]">{{ sessionUser?.displayName || '미로그인' }}</p>
        <p class="mt-0.5 text-xs text-[var(--gvp-text-secondary)]">{{ sessionUser ? `@${sessionUser.username} · ${sessionUser.role}` : '로그인이 필요합니다' }}</p>
      </div>
      <div class="border-b border-[var(--gvp-border)] py-1.5">
        <button type="button" class="flex w-full items-center justify-between px-4 py-2 text-xs text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]" @click="currency = currency === 'KRW' ? 'USD' : 'KRW'">
          <span>통화 표시</span><span class="font-semibold text-[var(--gvp-text)] tabular-nums">{{ currency === 'KRW' ? '₩ KRW' : '$ USD' }}</span>
        </button>
        <a v-for="c in consoles" :key="c.label" :href="c.href()" target="_blank" rel="noopener noreferrer" class="block px-4 py-2 text-xs text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]">{{ c.label }} ↗</a>
      </div>
      <div class="py-1.5">
        <button v-if="sessionUser" type="button" role="menuitem" class="block w-full px-4 py-2 text-left text-xs font-medium text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]" @click="$emit('logout')">로그아웃</button>
        <NuxtLink v-else to="/login" role="menuitem" class="block px-4 py-2 text-xs font-medium text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]">로그인</NuxtLink>
      </div>
    </div>
  </div>
</template>
