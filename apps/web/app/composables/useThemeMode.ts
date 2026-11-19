/** Dark-first theme mode: light / dark / system, syncs color-mode + data-theme. */
export type ThemePreference = 'light' | 'dark' | 'system'

export function useThemeMode() {
  const colorMode = useColorMode()
  const preference = computed<ThemePreference>(() => {
    const p = colorMode.preference
    return p === 'dark' || p === 'light' || p === 'system' ? p : 'system'
  })
  const isDark = computed(() => colorMode.value === 'dark')

  function apply(mode: ThemePreference) {
    colorMode.preference = mode
    if (import.meta.client) {
      const resolved = mode === 'system'
        ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
        : mode
      document.documentElement.dataset.theme = resolved
      document.documentElement.classList.toggle('dark', resolved === 'dark')
      document.documentElement.classList.toggle('light', resolved === 'light')
    }
  }

  function toggle() { apply(isDark.value ? 'light' : 'dark') }

  onMounted(() => {
    apply(preference.value)
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
      if (colorMode.preference === 'system') apply('system')
    })
  })
  return { isDark, preference, apply, toggle }
}
