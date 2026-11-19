<script setup lang="ts">
defineProps<{
  title: string
  description: string
  videoType: 'standard' | 'shorts'
  category: string
  tagsText: string
  typeLocked?: boolean
}>()
defineEmits<{
  (e: 'update:title', v: string): void
  (e: 'update:description', v: string): void
  (e: 'update:videoType', v: 'standard' | 'shorts'): void
  (e: 'update:category', v: string): void
  (e: 'update:tagsText', v: string): void
}>()
const categories = ['일반', '음악', '게임', '테크/개발', '팟캐스트', '뉴스', '교육', '엔터테인먼트']
</script>

<template>
  <div class="space-y-4">
    <div>
      <label class="block text-xs font-medium text-[var(--gvp-text)]" for="up-title">동영상 제목</label>
      <input
        id="up-title" :value="title" type="text" maxlength="100"
        class="mt-1.5 w-full rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-3 py-2 text-sm text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-focus)]"
        @input="$emit('update:title', ($event.target as HTMLInputElement).value)"
      />
    </div>
    <div>
      <label class="block text-xs font-medium text-[var(--gvp-text)]" for="up-desc">설명</label>
      <textarea
        id="up-desc" :value="description" rows="4"
        class="mt-1.5 w-full rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] p-3 text-sm text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-focus)]"
        @input="$emit('update:description', ($event.target as HTMLTextAreaElement).value)"
      />
    </div>
    <div>
      <label class="block text-xs font-medium text-[var(--gvp-text)]" for="up-tags">태그 (쉼표 구분, 최대 12개)</label>
      <input
        id="up-tags" :value="tagsText" type="text" placeholder="예: hls, ffmpeg, 자막"
        class="mt-1.5 w-full rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-3 py-2 text-sm text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-focus)]"
        @input="$emit('update:tagsText', ($event.target as HTMLInputElement).value)"
      />
    </div>
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div>
        <span class="block text-xs font-medium text-[var(--gvp-text)]" role="group" aria-label="영상 유형">영상 유형</span>
        <div class="mt-1.5 flex gap-1.5" role="radiogroup" aria-label="영상 유형">
          <button
            v-for="t in (['standard', 'shorts'] as const)" :key="t" type="button" role="radio" :aria-checked="videoType === t" :disabled="typeLocked"
            :class="videoType === t ? 'bg-[var(--gvp-text)] text-[var(--gvp-canvas)]' : 'bg-[var(--gvp-elevated)] text-[var(--gvp-text)]'"
            class="h-9 flex-1 rounded-md text-xs font-medium disabled:opacity-60" @click="$emit('update:videoType', t)"
          >
            {{ t === 'standard' ? '일반 동영상 (16:9)' : '숏폼 Shorts (9:16)' }}
          </button>
        </div>
        <p v-if="typeLocked" class="mt-1 text-[11px] text-[var(--gvp-text-secondary)]">유형은 업로드 시작 후 변경할 수 없습니다.</p>
      </div>
      <div>
        <label class="block text-xs font-medium text-[var(--gvp-text)]" for="up-cat">카테고리</label>
        <select
          id="up-cat" :value="category"
          class="mt-1.5 h-9 w-full rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 text-xs font-medium text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-focus)]"
          @change="$emit('update:category', ($event.target as HTMLSelectElement).value)"
        >
          <option v-for="c in categories" :key="c" :value="c">{{ c }}</option>
        </select>
      </div>
    </div>
  </div>
</template>
