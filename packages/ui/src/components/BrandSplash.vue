<script setup lang="ts">
import VravraLogo from './VravraLogo.vue'
import VravraStudioLogo from './VravraStudioLogo.vue'

withDefaults(defineProps<{ variant?: 'viewer' | 'studio' }>(), { variant: 'viewer' })

const show = ref(false)
const leaving = ref(false)
let timers: Array<ReturnType<typeof setTimeout>> = []

onMounted(() => {
  try {
    if (sessionStorage.getItem('vravra_splash_seen')) return
    sessionStorage.setItem('vravra_splash_seen', '1')
  } catch { return }
  show.value = true
  timers = [
    setTimeout(() => { leaving.value = true }, 1000),
    setTimeout(() => { show.value = false }, 1260),
  ]
})

onBeforeUnmount(() => { timers.forEach(clearTimeout) })
</script>

<template>
  <div
    v-if="show"
    class="fixed inset-0 z-[60] flex items-center justify-center"
    :class="leaving ? 'vravra-splash-leave' : ''"
    style="background-color: #0a0a0c;"
    role="status"
    aria-label="vravra"
  >
    <VravraLogo v-if="variant === 'viewer'" size="lg" animate show-tagline />
    <VravraStudioLogo v-else size="lg" animate show-tagline />
  </div>
</template>
