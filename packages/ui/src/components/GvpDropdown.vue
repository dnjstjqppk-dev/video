<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { placementClass, resolvePlacement, shouldCloseOnKey, type DropdownPlacement } from '../dropdown'

const props = withDefaults(defineProps<{
  placement?: DropdownPlacement
  widthPx?: number
  label: string
}>(), { placement: 'bottom-end', widthPx: 240 })

const open = ref(false)
const activePlacement = ref(props.placement)
const root = ref<HTMLElement | null>(null)
const panel = ref<HTMLElement | null>(null)

function toggle() { open.value ? close() : doOpen() }
function doOpen() {
  open.value = true
  requestAnimationFrame(() => {
    const t = root.value?.getBoundingClientRect()
    const h = panel.value?.offsetHeight ?? 280
    if (t) activePlacement.value = resolvePlacement(props.placement, t, h, window.innerHeight)
  })
}
function close() { open.value = false }
function onDocClick(e: MouseEvent) { if (open.value && !root.value?.contains(e.target as Node)) close() }
function onKey(e: KeyboardEvent) { if (open.value && shouldCloseOnKey(e.key)) close() }

watch(open, (v) => {
  document.addEventListener('click', onDocClick)
  document.addEventListener('keydown', onKey)
  if (!v) { document.removeEventListener('click', onDocClick); document.removeEventListener('keydown', onKey) }
})
onBeforeUnmount(() => { document.removeEventListener('click', onDocClick); document.removeEventListener('keydown', onKey) })
defineExpose({ close, open: doOpen })
</script>

<template>
  <div ref="root" class="gvp-dd">
    <button type="button" :aria-label="label" :aria-expanded="open" aria-haspopup="menu" @click.stop="toggle">
      <slot name="trigger" :open="open" />
    </button>
    <div v-if="open" ref="panel" role="menu" class="gvp-dd__panel gvp-dropdown gvp-elevated"
      :class="placementClass(activePlacement)" :style="{ width: `${widthPx}px` }">
      <slot :close="close" />
    </div>
  </div>
</template>

<style scoped>
.gvp-dd { position: relative; display: inline-flex; }
.gvp-dd > button { display: inline-flex; background: none; border: 0; padding: 0; cursor: pointer; }
.gvp-dd__panel { position: absolute; z-index: 50; border-radius: 12px; padding: 6px; }
</style>
