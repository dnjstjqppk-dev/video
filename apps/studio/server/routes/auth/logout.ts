export default defineEventHandler((event) => {
  deleteCookie(event, 'gvp_session', { path: '/' })
  deleteCookie(event, 'gvp_studio_token', { path: '/' })
  deleteCookie(event, 'gvp_csrf', { path: '/' })
  if (event.method === 'GET') {
    return sendRedirect(event, '/', 302)
  }
  return { status: 'logged_out' }
})
