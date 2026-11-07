export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const oidc = config.oidc as { issuer?: string; clientId?: string; clientSecret?: string }

  if (!oidc?.issuer || !oidc?.clientId) {
    throw createError({
      statusCode: 503,
      statusMessage: 'OIDC provider is not configured for this environment'
    })
  }

  const query = getQuery(event)
  if (query.error) {
    throw createError({
      statusCode: 401,
      statusMessage: `OIDC authorization failed: ${String(query.error)}`
    })
  }

  const stored = getCookie(event, 'gvp_oauth')
  deleteCookie(event, 'gvp_oauth', { path: '/' })
  if (!stored || !query.code || !query.state) {
    throw createError({ statusCode: 400, statusMessage: 'Missing OAuth state' })
  }

  let parsed: { state?: string; verifier?: string; redirectUri?: string }
  try {
    parsed = JSON.parse(stored)
  } catch {
    throw createError({ statusCode: 400, statusMessage: 'Invalid OAuth state' })
  }
  if (!parsed.state || parsed.state !== query.state || !parsed.verifier || !parsed.redirectUri) {
    throw createError({ statusCode: 400, statusMessage: 'OAuth state mismatch' })
  }

  const tokenRes = await $fetch<{
    access_token: string
    expires_in?: number
    token_type?: string
  }>(`${oidc.issuer.replace(/\/$/, '')}/oauth/v2/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      code: String(query.code),
      redirect_uri: parsed.redirectUri,
      client_id: oidc.clientId,
      ...(oidc.clientSecret ? { client_secret: oidc.clientSecret } : {}),
      code_verifier: parsed.verifier
    }).toString()
  }).catch(() => null)

  if (!tokenRes?.access_token) {
    throw createError({ statusCode: 502, statusMessage: 'OIDC token exchange failed' })
  }

  setCookie(event, 'gvp_session', tokenRes.access_token, {
    httpOnly: true,
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production',
    path: '/',
    maxAge: tokenRes.expires_in && tokenRes.expires_in > 0 ? tokenRes.expires_in : 3600
  })
  setCookie(event, 'gvp_csrf', crypto.randomUUID(), {
    httpOnly: false,
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production',
    path: '/',
    maxAge: tokenRes.expires_in && tokenRes.expires_in > 0 ? tokenRes.expires_in : 3600
  })

  return sendRedirect(event, '/', 302)
})
