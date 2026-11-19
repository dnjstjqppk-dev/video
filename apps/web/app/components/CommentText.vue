<script setup lang="ts">
const props = defineProps<{ text: string }>()
const emit = defineEmits<{ (e: 'seek', seconds: number): void }>()

interface Seg { kind: 'text' | 'stamp'; value: string; seconds?: number }

function toSeconds(stamp: string): number {
  const parts = stamp.split(':').map(Number)
  if (parts.length === 2) return (parts[0] || 0) * 60 + (parts[1] || 0)
  return (parts[0] || 0) * 3600 + (parts[1] || 0) * 60 + (parts[2] || 0)
}

const segments = computed<Seg[]>(() => {
  const out: Seg[] = []
  const re = /\b(\d{1,3}:\d{2}(?::\d{2})?)\b/g
  let last = 0
  let m: RegExpExecArray | null
  while ((m = re.exec(props.text)) !== null) {
    const stamp = m[1] || ''
    if (!stamp) continue
    if (m.index > last) out.push({ kind: 'text', value: props.text.slice(last, m.index) })
    out.push({ kind: 'stamp', value: stamp, seconds: toSeconds(stamp) })
    last = m.index + stamp.length
  }
  if (last < props.text.length) out.push({ kind: 'text', value: props.text.slice(last) })
  return out
})
</script>

<template>
  <span class="whitespace-pre-line"><template v-for="(s, i) in segments" :key="i"><button
    v-if="s.kind === 'stamp'"
    type="button"
    class="tabular-nums font-medium text-[var(--gvp-brand-hover)] hover:underline"
    @click="emit('seek', s.seconds || 0)"
  >{{ s.value }}</button><template v-else>{{ s.value }}</template></template></span>
</template>
