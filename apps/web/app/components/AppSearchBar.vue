<script setup lang="ts">
const api = useApi()
const router = useRouter()
const route = useRoute()
const recent = useRecentSearch()
const query = ref((route.query.q as string) || '')
const suggestions = ref<string[]>([])
const open = ref(false)
const activeIdx = ref(-1)
const filters = ['전체', '동영상', '숏폼', '라이브', '재생목록']
const activeFilter = ref('전체')
let timer: ReturnType<typeof setTimeout> | null = null

watch(() => route.fullPath, () => { open.value = false; query.value = (route.query.q as string) || '' })
function fetchSug() {
  if (timer) clearTimeout(timer)
  const q = query.value.trim()
  if (!q) { suggestions.value = []; return }
  timer = setTimeout(async () => {
    try { suggestions.value = await api.getSearchSuggestions(q) } catch { suggestions.value = [] }
  }, 180)
}
function submit(term?: string) {
  const q = (term ?? query.value).trim()
  if (!q) return
  recent.add(q); open.value = false; query.value = q
  router.push({ path: '/search', query: { q, ...(activeFilter.value !== '전체' ? { type: activeFilter.value } : {}) } })
}
function onKey(e: KeyboardEvent) {
  const n = suggestions.value.length
  if (e.key === 'Escape') open.value = false
  if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
    e.preventDefault()
    activeIdx.value = e.key === 'ArrowDown' ? (activeIdx.value + 1) % Math.max(n, 1) : (activeIdx.value - 1 + n) % Math.max(n, 1)
  }
  if (e.key === 'Enter' && activeIdx.value >= 0 && suggestions.value[activeIdx.value]) submit(suggestions.value[activeIdx.value])
}
</script>

<template>
  <div v-click-outside="() => (open = false)" class="relative mx-4 flex min-w-0 max-w-xl flex-1">
    <form class="flex w-full gap-1.5" @submit.prevent="submit()">
      <input v-model="query" type="search" role="combobox" :aria-expanded="open" aria-label="검색"
        placeholder="동영상, 쇼츠, 라이브, 채널 검색..." @focus="open = true; fetchSug()" @input="fetchSug(); open = true" @keydown="onKey"
        class="h-9 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-3 text-sm text-[var(--gvp-text)] placeholder:text-[var(--gvp-text-muted)] focus:border-[var(--gvp-focus)] focus:outline-none" />
      <button type="submit" aria-label="검색 실행" class="inline-flex h-9 items-center rounded-md bg-[var(--gvp-brand)] px-3.5 text-white hover:bg-[var(--gvp-brand-hover)]">
        <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" /></svg>
      </button>
    </form>
    <div v-if="open && (query.trim() || recent.terms.value.length)" class="gvp-dropdown gvp-elevated absolute left-0 right-0 top-10 z-50 max-h-96 overflow-y-auto rounded-xl p-1.5">
      <div class="flex gap-1.5 overflow-x-auto px-1.5 pb-2 pt-1">
        <button v-for="f in filters" :key="f" type="button" :aria-pressed="activeFilter === f" @click="activeFilter = f"
          :class="activeFilter === f ? 'bg-[var(--gvp-text)] text-[var(--gvp-canvas)]' : 'bg-[var(--gvp-elevated)] text-[var(--gvp-text)]'"
          class="h-7 shrink-0 rounded-md px-2.5 text-xs font-medium">{{ f }}</button>
      </div>
      <template v-if="!query.trim()">
        <p class="px-2.5 pb-1 pt-1 text-[11px] font-semibold text-[var(--gvp-text-muted)]">최근 검색어</p>
        <div v-for="t in recent.terms.value" :key="t" class="group flex items-center rounded-lg hover:bg-[var(--gvp-elevated)]">
          <button type="button" class="flex flex-1 items-center gap-2.5 px-2.5 py-2 text-left text-sm text-[var(--gvp-text)]" @mousedown.prevent="submit(t)">{{ t }}</button>
          <button type="button" :aria-label="`${t} 삭제`" class="mr-1 rounded p-1 text-[var(--gvp-text-muted)] hover:text-[var(--gvp-text)]" @mousedown.prevent="recent.remove(t)">✕</button>
        </div>
      </template>
      <template v-else>
        <button v-for="(s, i) in suggestions" :key="s" type="button" :class="i === activeIdx ? 'bg-[var(--gvp-elevated)]' : ''"
          class="flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-left text-sm text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]" @mousedown.prevent="submit(s)">{{ s }}</button>
      </template>
    </div>
  </div>
</template>
