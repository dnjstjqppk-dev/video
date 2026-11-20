export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const upstream = ((config.internalApiUrl as string) || 'http://localhost:8080').replace(/\/$/, '')
  try {
    return await proxyRequest(event, `${upstream}/healthz`, { fetchOptions: { redirect: 'manual' } })
  } catch (err: unknown) {
    const e = err as { response?: { status?: number } }
    return apiError(event, e?.response?.status || 502, 'API_UNHEALTHY', 'API health check failed')
  }
})
