/** Recent search terms persisted in localStorage (max 8). */
const KEY = 'gvp_recent_search_v1'
const MAX = 8

export function useRecentSearch() {
  const terms = ref<string[]>([])

  function load() {
    if (!import.meta.client) return
    try {
      const raw = JSON.parse(localStorage.getItem(KEY) || '[]')
      terms.value = Array.isArray(raw) ? raw.filter((t): t is string => typeof t === 'string').slice(0, MAX) : []
    } catch { terms.value = [] }
  }

  function persist() {
    try { localStorage.setItem(KEY, JSON.stringify(terms.value)) } catch { /* quota */ }
  }

  function add(term: string) {
    const t = term.trim()
    if (!t) return
    terms.value = [t, ...terms.value.filter(x => x !== t)].slice(0, MAX)
    persist()
  }

  function remove(term: string) {
    terms.value = terms.value.filter(x => x !== term)
    persist()
  }

  function clear() { terms.value = []; persist() }

  onMounted(load)
  return { terms, add, remove, clear }
}
