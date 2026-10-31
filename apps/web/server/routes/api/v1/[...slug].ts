export default defineEventHandler((event) => {
  setResponseStatus(event, 404)
  setResponseHeaders(event, {
    'Content-Type': 'application/json; charset=utf-8',
    'Access-Control-Allow-Origin': '*'
  })
  return {
    error: {
      code: 'NOT_FOUND',
      message: `Endpoint '${getRequestURL(event).pathname}' is not found.`,
      status: 404
    }
  }
})
