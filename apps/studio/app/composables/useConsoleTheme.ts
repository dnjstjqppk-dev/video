/** Console theme: Pure Crisp Light default, persisted toggle. */
const KEY = 'gvp_console_theme'

export function useConsoleTheme() {
  const mode = useState<'light' | 'dark'>('console-theme', () => 'light')

  function apply(next: 'light' | 'dark') {
    mode.value = next
    if (import.meta.client) {
      document.documentElement.dataset.theme = next
      document.documentElement.classList.toggle('dark', next === 'dark')
      document.documentElement.classList.toggle('light', next === 'light')
      try { localStorage.setItem(KEY, next) } catch { /* quota */ }
    }
  }

  function toggle() { apply(mode.value === 'dark' ? 'light' : 'dark') }

  onMounted(() => {
    try {
      const saved = localStorage.getItem(KEY)
      apply(saved === 'dark' ? 'dark' : 'light')
    } catch { apply('light') }
  })

  return { mode, apply, toggle }
}
