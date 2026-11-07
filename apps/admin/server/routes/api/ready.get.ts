export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const upstream = ((config.internalApiUrl as string) || 'http://localhost:8080').replace(/\/$/, '')
  try {
    return await proxyRequest(event, `${upstream}/ready`, { fetchOptions: { redirect: 'manual' } })
  } catch (err: unknown) {
    const e = err as { response?: { status?: number } }
    throw createError({ statusCode: e?.response?.status || 502, statusMessage: 'API readiness check failed' })
  }
})
