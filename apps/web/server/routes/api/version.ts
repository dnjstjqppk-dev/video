export default defineEventHandler((event) => {
  setResponseHeaders(event, {
    'Content-Type': 'application/json; charset=utf-8',
    'Cache-Control': 'no-store, no-cache, must-revalidate',
    'Access-Control-Allow-Origin': '*'
  })
  return {
    app: 'web',
    commit: '3f8c1ef',
    builtAt: '2026-11-01T03:00:00Z',
    workerVersion: 'v2026.10.30',
    status: 'healthy'
  }
})
