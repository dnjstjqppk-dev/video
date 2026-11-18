/** Dark-first theme mode: syncs color-mode + data-theme attr. */
export function useThemeMode() {
  const colorMode = useColorMode()
  const isDark = computed(() => colorMode.value !== 'light')

  function apply(mode: 'dark' | 'light') {
    colorMode.preference = mode
    if (import.meta.client) {
      document.documentElement.dataset.theme = mode
      document.documentElement.classList.toggle('dark', mode === 'dark')
      document.documentElement.classList.toggle('light', mode === 'light')
    }
  }

  function toggle() { apply(isDark.value ? 'light' : 'dark') }

  onMounted(() => apply(colorMode.value === 'light' ? 'light' : 'dark'))
  return { isDark, apply, toggle }
}
