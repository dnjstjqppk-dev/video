import type { H3Event } from 'h3'

// Server routes must return JSON errors explicitly instead of throwing into
// the HTML error renderer: a thrown H3Error rendered as HTML crashes SSR
// payload serialization and masks the real status as a 500. The returned
// shape mirrors h3's own JSON error body so existing clients keep working.
export function apiError(event: H3Event, statusCode: number, code: string, message: string, statusMessage?: string) {
  setResponseStatus(event, statusCode)
  const text = statusMessage || message
  return {
    statusCode,
    statusMessage: text,
    message: text,
    data: { error: { code, message } }
  }
}

// Browser visits to /auth/* land back on the login page with ?error=;
// API clients keep receiving JSON errors with the real status code.
export function authError(event: H3Event, statusCode: number, code: string, message: string) {
  const accept = getRequestHeader(event, 'accept') || ''
  if (accept.includes('text/html')) {
    return sendRedirect(event, `/login?error=${encodeURIComponent(code)}`, 302)
  }
  return apiError(event, statusCode, code, message)
}
