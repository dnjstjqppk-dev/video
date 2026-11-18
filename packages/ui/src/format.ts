/** Locale-aware metric formatting (tabular-nums display). */
export function formatViewCount(n: number, locale = 'ko'): string {
  if (!Number.isFinite(n) || n < 0) return locale === 'ko' ? '0회' : '0 views'
  if (locale === 'ko') {
    if (n >= 100_000_000) return `${trimNum(n / 100_000_000)}억회`
    if (n >= 10_000) return `${trimNum(n / 10_000)}만회`
    if (n >= 1_000) return `${trimNum(n / 1_000)}천회`
    return `${n}회`
  }
  if (n >= 1_000_000) return `${trimNum(n / 1_000_000)}M views`
  if (n >= 1_000) return `${trimNum(n / 1_000)}K views`
  return `${n} views`
}

export function formatSubscriberCount(n: number, locale = 'ko'): string {
  const v = formatViewCount(n, locale)
  return locale === 'ko' ? v.replace(/회$/, '명') : v.replace(/ views$/, '')
}

export function formatRelativeTime(iso: string, locale = 'ko', now = Date.now()): string {
  const t = new Date(iso).getTime()
  if (!Number.isFinite(t)) return locale === 'ko' ? '날짜 미상' : 'unknown'
  const s = Math.max(0, Math.floor((now - t) / 1000))
  const u = (ko: string, en: string) => (locale === 'ko' ? ko : en)
  if (s < 60) return u('방금 전', 'just now')
  const m = Math.floor(s / 60)
  if (m < 60) return u(`${m}분 전`, `${m}m ago`)
  const h = Math.floor(m / 60)
  if (h < 24) return u(`${h}시간 전`, `${h}h ago`)
  const d = Math.floor(h / 24)
  if (d < 7) return u(`${d}일 전`, `${d}d ago`)
  const w = Math.floor(d / 7)
  if (w < 5) return u(`${w}주 전`, `${w}w ago`)
  const mo = Math.floor(d / 30)
  if (mo < 12) return u(`${mo}개월 전`, `${mo}mo ago`)
  return u(`${Math.floor(mo / 12)}년 전`, `${Math.floor(mo / 12)}y ago`)
}

export function formatTimecode(sec: number): string {
  if (!Number.isFinite(sec) || sec < 0) return '0:00'
  const t = Math.floor(sec)
  const h = Math.floor(t / 3600)
  const m = Math.floor((t % 3600) / 60)
  const s = t % 60
  const ss = s < 10 ? `0${s}` : `${s}`
  if (h > 0) return `${h}:${m < 10 ? `0${m}` : m}:${ss}`
  return `${m}:${ss}`
}

export function trimNum(n: number): string {
  return n >= 100 ? `${Math.round(n)}` : `${Math.round(n * 10) / 10}`
}
