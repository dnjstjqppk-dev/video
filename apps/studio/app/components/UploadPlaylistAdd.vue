<script setup lang="ts">
const props = defineProps<{ videoId: string | null }>()
const api = useStudioApi()
const title = ref('')
const busy = ref(false)
const notice = ref<string | null>(null)
const failed = ref<string | null>(null)

async function add() {
  const name = title.value.trim()
  if (!props.videoId || !name || busy.value) return
  busy.value = true
  notice.value = null
  failed.value = null
  try {
    const pl = await api.createPlaylist({ title: name, visibility: 'private' })
    await api.addPlaylistItem(pl.id, props.videoId)
    notice.value = `'${pl.title}' 재생목록에 추가했습니다.`
    title.value = ''
  } catch {
    failed.value = '재생목록 추가에 실패했습니다.'
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div>
    <p v-if="!videoId" class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-4 text-center text-xs text-[var(--gvp-text-secondary)]">
      업로드를 시작하면 재생목록에 담을 수 있습니다.
    </p>
    <div v-else class="flex gap-1.5">
      <input
        v-model="title"
        type="text"
        placeholder="새 재생목록 이름"
        aria-label="새 재생목록 이름"
        class="h-9 min-w-0 flex-1 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-3 text-xs outline-none placeholder:text-[var(--gvp-text-muted)] focus:border-[var(--gvp-focus)]"
        @keydown.enter="add"
      />
      <button type="button" :disabled="busy || !title.trim()" class="h-9 shrink-0 rounded-full bg-[var(--gvp-studio)] px-4 text-xs font-semibold text-white hover:opacity-90 disabled:opacity-50" @click="add">
        {{ busy ? '…' : '담기' }}
      </button>
    </div>
    <p v-if="notice" role="status" class="mt-1.5 text-xs font-medium text-[var(--gvp-success)]">{{ notice }}</p>
    <p v-if="failed" role="alert" class="mt-1.5 text-xs font-medium text-[var(--gvp-live)]">{{ failed }}</p>
  </div>
</template>
