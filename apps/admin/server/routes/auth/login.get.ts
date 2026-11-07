export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const oidc = config.oidc as { issuer?: string; clientId?: string }

  if (!oidc?.issuer || !oidc?.clientId) {
    throw createError({
      statusCode: 503,
      statusMessage: 'OIDC provider is not configured for this environment'
    })
  }

  const state = crypto.randomUUID()
  const verifier = `${crypto.randomUUID()}${crypto.randomUUID()}`.replace(/-/g, '')
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier))
  const challenge = Buffer.from(digest)
    .toString('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')

  const redirectUri = `${getRequestURL(event).origin}/auth/callback`
  setCookie(event, 'gvp_oauth', JSON.stringify({ state, verifier, redirectUri }), {
    httpOnly: true,
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production',
    path: '/',
    maxAge: 600
  })

  const authorizeUrl =
    `${oidc.issuer.replace(/\/$/, '')}/oauth/v2/authorize` +
    `?response_type=code&client_id=${encodeURIComponent(oidc.clientId)}` +
    `&redirect_uri=${encodeURIComponent(redirectUri)}` +
    `&scope=${encodeURIComponent('openid profile email offline_access')}` +
    `&state=${encodeURIComponent(state)}` +
    `&code_challenge=${encodeURIComponent(challenge)}` +
    `&code_challenge_method=S256`

  return sendRedirect(event, authorizeUrl, 302)
})
