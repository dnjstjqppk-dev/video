export default defineEventHandler(async (event) => {
  const host = (getRequestHeader(event, 'host') || '').split(':')[0]
  if (host !== 'localhost' && host !== '127.0.0.1') {
    throw createError({
      statusCode: 403,
      statusMessage: 'Development token login is only available on localhost'
    })
  }

  const body = await readBody<{ token?: string }>(event).catch((): { token?: string } => ({}))
  const token = (body?.token || '').trim()
  if (!token) {
    throw createError({ statusCode: 400, statusMessage: 'Token is required' })
  }

  setCookie(event, 'gvp_session', token, {
    httpOnly: true,
    sameSite: 'lax',
    secure: false,
    path: '/',
    maxAge: 86400
  })
  return { status: 'ok' }
})
