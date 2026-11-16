declare const __DEPLOY_COMMIT__: string
declare const __DEPLOY_BUILT_AT__: string

export default defineEventHandler((event) => {
  setResponseHeaders(event, {
    'Content-Type': 'application/json; charset=utf-8',
    'Cache-Control': 'no-store, no-cache, must-revalidate',
    'Access-Control-Allow-Origin': '*'
  })
  return {
    app: 'admin',
    commit: __DEPLOY_COMMIT__,
    builtAt: __DEPLOY_BUILT_AT__,
    status: 'healthy'
  }
})
