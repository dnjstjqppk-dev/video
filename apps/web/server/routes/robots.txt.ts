export default defineEventHandler((event) => {
  const config = useRuntimeConfig(event)
  const siteUrl = config.public.siteUrl || 'https://www.gvp.media'

  setHeader(event, 'Content-Type', 'text/plain; charset=utf-8')
  setHeader(event, 'Cache-Control', 'public, max-age=3600')

  return [
    'User-agent: *',
    'Allow: /',
    'Allow: /watch/',
    'Allow: /channels/',
    'Allow: /shorts',
    'Disallow: /api/',
    `Sitemap: ${siteUrl}/sitemap.xml`
  ].join('\n')
})
