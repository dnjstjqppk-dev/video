/** v-click-outside: close dropdowns on outside pointer + Escape. */
export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.directive('click-outside', {
    mounted(el: HTMLElement, binding: { value: () => void }) {
      const onDown = (e: PointerEvent) => {
        if (!el.contains(e.target as Node)) binding.value()
      }
      const onKey = (e: KeyboardEvent) => {
        if (e.key === 'Escape') binding.value()
      }
      ;(el as unknown as { __co?: () => void }).__co = () => {
        document.removeEventListener('pointerdown', onDown)
        document.removeEventListener('keydown', onKey)
      }
      document.addEventListener('pointerdown', onDown)
      document.addEventListener('keydown', onKey)
    },
    unmounted(el: HTMLElement) {
      ;(el as unknown as { __co?: () => void }).__co?.()
    },
  })
})
