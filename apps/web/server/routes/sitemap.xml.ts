export default defineEventHandler((event) => {
  const config = useRuntimeConfig(event)
  const siteUrl = (config.public.siteUrl || 'https://www.gvp.media').replace(/\/$/, '')
  const now = new Date().toISOString()

  const urls = [
    { loc: `${siteUrl}/`, priority: '1.0', changefreq: 'hourly' },
    { loc: `${siteUrl}/shorts`, priority: '0.9', changefreq: 'hourly' },
    { loc: `${siteUrl}/watch/vid_arch_2026`, priority: '0.9', changefreq: 'daily' },
    { loc: `${siteUrl}/watch/vid_seoul_4k`, priority: '0.9', changefreq: 'daily' },
    { loc: `${siteUrl}/watch/vid_esports_finals`, priority: '0.8', changefreq: 'daily' },
    { loc: `${siteUrl}/channels/ch_tech_kr`, priority: '0.8', changefreq: 'daily' },
    { loc: `${siteUrl}/live/live_main_01`, priority: '0.8', changefreq: 'always' },
    { loc: `${siteUrl}/premium`, priority: '0.7', changefreq: 'weekly' }
  ]

  setHeader(event, 'Content-Type', 'application/xml; charset=utf-8')
  setHeader(event, 'Cache-Control', 'public, max-age=1800')

  const xmlItems = urls
    .map(
      u => `  <url>
    <loc>${u.loc}</loc>
    <lastmod>${now}</lastmod>
    <changefreq>${u.changefreq}</changefreq>
    <priority>${u.priority}</priority>
  </url>`
    )
    .join('\n')

  return `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${xmlItems}
</urlset>`
})
