<script setup lang="ts">
import { ref } from 'vue'
import GvpDropdown from './GvpDropdown.vue'
import { nextFocusIndex } from '../dropdown'

export interface KebabItem { id: string; label: string; danger?: boolean; hint?: string }

defineProps<{ items: KebabItem[]; label: string }>()
defineEmits<{ (e: 'action', id: string): void }>()

const dd = ref<InstanceType<typeof GvpDropdown> | null>(null)
const focusIdx = ref(-1)

function onKey(e: KeyboardEvent, count: number) {
  if (['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(e.key)) {
    e.preventDefault()
    focusIdx.value = nextFocusIndex(e.key, focusIdx.value, count)
    document.getElementById(`kebab-${focusIdx.value}`)?.focus()
  }
}
function pick(id: string, emit: (e: 'action', id: string) => void) { emit('action', id); dd.value?.close() }
</script>

<template>
  <GvpDropdown ref="dd" :label="label" :width-px="220">
    <template #trigger>
      <span class="gvp-kebab" aria-hidden="true">⋮</span>
    </template>
    <template #default="{ close }">
      <div role="menu" @keydown="onKey($event, items.length)">
        <button v-for="(it, i) in items" :id="`kebab-${i}`" :key="it.id" type="button" role="menuitem"
          class="gvp-kebab__it" :class="{ 'gvp-kebab__it--danger': it.danger }"
          @click="pick(it.id, $emit); close()">
          <span>{{ it.label }}</span>
          <kbd v-if="it.hint">{{ it.hint }}</kbd>
        </button>
      </div>
    </template>
  </GvpDropdown>
</template>

<style scoped>
.gvp-kebab { display: inline-flex; align-items: center; justify-content: center; width: 32px; height: 32px; border-radius: 50%; font-size: 1rem; color: var(--gvp-text-secondary); }
.gvp-kebab:hover { background: var(--gvp-elevated); color: var(--gvp-text); }
.gvp-kebab__it { display: flex; align-items: center; justify-content: space-between; gap: 8px; width: 100%; padding: 8px 10px; border: 0; border-radius: 8px; background: none; color: var(--gvp-text); font-size: 0.8125rem; cursor: pointer; text-align: left; }
.gvp-kebab__it:hover { background: var(--gvp-elevated); }
.gvp-kebab__it--danger { color: var(--gvp-danger); }
.gvp-kebab__it kbd { font-family: var(--gvp-font-mono); font-size: 0.6875rem; color: var(--gvp-text-muted); }
</style>
