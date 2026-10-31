export default defineEventHandler((event) => {
  setResponseHeaders(event, {
    'Content-Type': 'application/json; charset=utf-8',
    'Cache-Control': 'public, max-age=60',
    'Access-Control-Allow-Origin': '*'
  })

  // Reuse handler or return home feed structure
  return $fetch('/api/v1/videos')
})
