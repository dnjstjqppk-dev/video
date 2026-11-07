const SAFE_METHODS = new Set(['GET', 'HEAD', 'OPTIONS'])

function assertSameOriginOrCsrfToken(event: Parameters<Parameters<typeof defineEventHandler>[0]>[0]): void {
  if (SAFE_METHODS.has(event.method)) return
  const host = getRequestHeader(event, 'host') || ''
  const origin = getRequestHeader(event, 'origin') || ''
  const referer = getRequestHeader(event, 'referer') || ''
  const matchesHost = (value: string): boolean => {
    try {
      return new URL(value).host === host
    } catch {
      return false
    }
  }
  const originOk = (origin !== '' && matchesHost(origin)) || (origin === '' && referer !== '' && matchesHost(referer))
  const csrfCookie = getCookie(event, 'gvp_csrf')
  const csrfHeader = getRequestHeader(event, 'x-csrf-token')
  const tokenOk = !!csrfCookie && !!csrfHeader && csrfCookie === csrfHeader
  if (!originOk && !tokenOk) {
    throw createError({
      statusCode: 403,
      statusMessage: 'CSRF validation failed',
      data: { error: { code: 'CSRF_VALIDATION_FAILED', message: 'Cross-site request rejected' } }
    })
  }
}

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const upstream = ((config.internalApiUrl as string) || 'http://localhost:8080').replace(/\/$/, '')
  const target = `${upstream}${event.path.replace(/^\/api/, '')}`

  // Mutations require same-origin proof or a double-submit CSRF token.
  assertSameOriginOrCsrfToken(event)

  // Inject the httpOnly session token so browser JS never handles credentials.
  if (!getRequestHeader(event, 'authorization')) {
    const session = getCookie(event, 'gvp_session')
    if (session) {
      event.node.req.headers['authorization'] = `Bearer ${session}`
    }
  }

  try {
    return await proxyRequest(event, target, {
      fetchOptions: {
        redirect: 'manual'
      }
    })
  } catch (err: any) {
    const status = err?.response?.status || err?.statusCode || 502
    setResponseStatus(event, status)
    setResponseHeaders(event, {
      'Content-Type': 'application/json; charset=utf-8'
    })
    return {
      error: {
        code: err?.data?.error?.code || (status === 502 ? 'UPSTREAM_UNAVAILABLE' : 'UPSTREAM_ERROR'),
        message: err?.data?.error?.message || 'Upstream API request failed',
        request_id: getRequestHeader(event, 'x-request-id') || `req_proxy_${Date.now()}`
      }
    }
  }
})
