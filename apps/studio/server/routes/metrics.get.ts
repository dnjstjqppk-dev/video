const startedAt = Date.now()

export default defineEventHandler((event) => {
  setResponseHeaders(event, {
    'Content-Type': 'text/plain; version=0.0.4; charset=utf-8',
    'Cache-Control': 'no-store'
  })
  const uptimeSeconds = Math.floor((Date.now() - startedAt) / 1000)
  return [
    '# HELP gvp_nuxt_up Nitro server up',
    '# TYPE gvp_nuxt_up gauge',
    'gvp_nuxt_up{surface="studio"} 1',
    '# HELP gvp_nuxt_uptime_seconds Nitro server uptime in seconds',
    '# TYPE gvp_nuxt_uptime_seconds gauge',
    `gvp_nuxt_uptime_seconds{surface="studio"} ${uptimeSeconds}`,
    ''
  ].join('\n')
})
