/** Console theme: dark cinema default + light + system, persisted. */
export type ConsoleThemePreference = 'light' | 'dark' | 'system'
const KEY = 'gvp_console_theme'

function resolveSystem(): 'light' | 'dark' {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function useConsoleTheme() {
  const preference = useState<ConsoleThemePreference>('console-theme', () => 'dark')
  const mode = computed<'light' | 'dark'>(() =>
    preference.value === 'system'
      ? (import.meta.client ? resolveSystem() : 'dark')
      : preference.value,
  )

  function apply(next: ConsoleThemePreference) {
    preference.value = next
    if (import.meta.client) {
      const resolved = next === 'system' ? resolveSystem() : next
      document.documentElement.dataset.theme = resolved
      document.documentElement.classList.toggle('dark', resolved === 'dark')
      document.documentElement.classList.toggle('light', resolved === 'light')
      try { localStorage.setItem(KEY, next) } catch { /* quota */ }
    }
  }

  function toggle() { apply(mode.value === 'dark' ? 'light' : 'dark') }

  onMounted(() => {
    try {
      const saved = localStorage.getItem(KEY)
      apply(saved === 'light' || saved === 'system' ? saved : 'dark')
    } catch { apply('dark') }
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
      if (preference.value === 'system') apply('system')
    })
  })

  return { mode, preference, apply, toggle }
}
