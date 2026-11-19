<script setup lang="ts">
import type { Video } from '@gvp/shared-types'

const props = defineProps<{ video: Video }>()
const api = useApi()
const library = useLibraryStore()
const open = ref(false)
const newTitle = ref('')
const busy = ref(false)
const notice = ref<string | null>(null)
const savedHere = computed(() => library.saved.some(v => v.id === props.video.id))

function toggleWatchLater() {
  library.toggleSaved(props.video)
}

async function createAndAdd() {
  const title = newTitle.value.trim()
  if (!title || busy.value) return
  busy.value = true
  notice.value = null
  try {
    const pl = await api.createPlaylist({ title, visibility: 'private' })
    await api.addPlaylistItem(pl.id, props.video.id)
    notice.value = `'${pl.title}' 재생목록에 저장했습니다.`
    newTitle.value = ''
  } catch {
    notice.value = '재생목록 저장에 실패했습니다.'
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div v-click-outside="() => (open = false)" class="relative">
    <button
      type="button"
      :aria-expanded="open"
      class="flex items-center gap-1.5 rounded-full px-3.5 py-1.5 text-xs font-semibold transition"
      :class="savedHere ? 'bg-[var(--gvp-success-soft)] text-[var(--gvp-success)]' : 'bg-[var(--gvp-elevated)] text-[var(--gvp-text)] hover:bg-[var(--gvp-canvas)]'"
      @click="open = !open"
    >
      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
      </svg>
      <span>{{ savedHere ? '저장됨' : '저장' }}</span>
    </button>
    <div v-if="open" role="menu" class="gvp-dropdown gvp-elevated absolute right-0 top-full z-30 mt-2 w-64 rounded-xl p-2">
      <button type="button" role="menuitemcheckbox" :aria-checked="savedHere" class="flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-left text-xs text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]" @click="toggleWatchLater">
        <span class="flex h-4 w-4 items-center justify-center rounded border border-[var(--gvp-border-strong)] text-[10px]" :class="savedHere ? 'bg-[var(--gvp-brand)] text-white' : ''">{{ savedHere ? '✓' : '' }}</span>
        나중에 볼 동영상
      </button>
      <div class="my-1.5 border-t border-[var(--gvp-border)]" />
      <p class="px-2.5 pb-1 text-[11px] font-semibold text-[var(--gvp-text-muted)]">새 재생목록에 저장</p>
      <div class="flex gap-1.5 px-1.5 pb-1">
        <input
          v-model="newTitle"
          type="text"
          placeholder="재생목록 이름"
          aria-label="새 재생목록 이름"
          class="h-8 min-w-0 flex-1 rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-2.5 text-xs text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-brand)]"
          @keydown.enter="createAndAdd"
        />
        <button type="button" :disabled="busy || !newTitle.trim()" class="h-8 shrink-0 rounded-md bg-[var(--gvp-brand)] px-3 text-xs font-semibold text-white hover:bg-[var(--gvp-brand-hover)] disabled:opacity-50" @click="createAndAdd">
          {{ busy ? '…' : '만들기' }}
        </button>
      </div>
      <p v-if="notice" role="status" class="px-2.5 py-1 text-[11px] text-[var(--gvp-text-secondary)]">{{ notice }}</p>
    </div>
  </div>
</template>
