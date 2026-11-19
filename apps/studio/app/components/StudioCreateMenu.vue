<script setup lang="ts">
const open = ref(false)
const route = useRoute()
watch(() => route.path, () => { open.value = false })

const items = [
  { path: '/upload', label: '동영상 업로드', hint: '멀티파트 업로드' },
  { path: '/live', label: '라이브 시작', hint: '라이브 관제실' }
]
</script>

<template>
  <div class="relative">
    <button
      type="button"
      class="inline-flex h-9 items-center gap-1.5 rounded-full bg-[var(--gvp-brand)] px-3.5 text-xs font-semibold text-white transition-colors hover:bg-[var(--gvp-brand-hover)]"
      :aria-expanded="open"
      aria-haspopup="menu"
      @click="open = !open"
      @keydown.escape="open = false"
    >
      <span class="text-base leading-none" aria-hidden="true">+</span>
      <span class="hidden sm:inline">만들기</span>
    </button>
    <div v-if="open" class="absolute right-0 z-50 mt-2 w-56 overflow-hidden rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] py-1.5 shadow-lg" role="menu">
      <NuxtLink
        v-for="it in items"
        :key="it.path"
        :to="it.path"
        role="menuitem"
        class="block px-4 py-2.5 hover:bg-[var(--gvp-elevated)]"
        @click="open = false"
      >
        <span class="block text-sm font-medium text-[var(--gvp-text)]">{{ it.label }}</span>
        <span class="block text-xs text-[var(--gvp-text-secondary)]">{{ it.hint }}</span>
      </NuxtLink>
    </div>
  </div>
</template>
