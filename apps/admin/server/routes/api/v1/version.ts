export default defineEventHandler((event) => {
  const config = useRuntimeConfig()
  setResponseHeaders(event, {
    'Content-Type': 'application/json; charset=utf-8',
    'Cache-Control': 'no-store, no-cache, must-revalidate',
    'Access-Control-Allow-Origin': '*'
  })
  return {
    app: 'admin',
    commit: (config.public.deployCommit as string) || 'unknown',
    builtAt: (config.public.deployBuiltAt as string) || 'unknown',
    status: 'healthy'
  }
})
