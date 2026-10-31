export default defineEventHandler((event) => {
  const host = getRequestHeader(event, 'host')
  // D-08 / GV-08: Canonical redirect from web mirror to primary domain
  if (host && host.startsWith('global-video-platform-web.')) {
    const url = getRequestURL(event)
    const target = `https://global-video-platform.dnjstjqppk.workers.dev${url.pathname}${url.search}`
    return sendRedirect(event, target, 301)
  }
})
