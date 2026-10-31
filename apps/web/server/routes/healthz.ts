export default defineEventHandler((event) => {
  setResponseHeaders(event, {
    'Content-Type': 'application/json; charset=utf-8',
    'Cache-Control': 'no-cache, no-store, must-revalidate'
  })
  return {
    status: 'healthy',
    app: 'web',
    commit: '3f8c1ef',
    timestamp: new Date().toISOString(),
    services: {
      frontend: 'healthy',
      runtime: 'cloudflare-worker'
    }
  }
})
