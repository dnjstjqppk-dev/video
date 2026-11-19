import { createApiClient, type ApiClient } from '@gvp/ts-api-client'

let clientInstance: ApiClient | null = null

export function useApi(): ApiClient {
  const config = useRuntimeConfig()
  const authCookie = useCookie<string | null>('gvp_access_token', {
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production'
  })
  // The httpOnly session cookie is readable here on the server only, which
  // authenticates SSR API calls; in the browser the /api proxy injects it.
  const sessionCookie = useCookie<string | null>('gvp_session')

  if (!clientInstance || import.meta.server) {
    const baseUrl = import.meta.server
      ? (config.internalApiUrl as string) || (config.public.apiBaseUrl as string)
      : (config.public.apiBaseUrl as string)

    clientInstance = createApiClient({
      baseUrl,
      getToken: () => authCookie.value ?? sessionCookie.value ?? null,
      onUnauthorized: () => {
        authCookie.value = null
      }
    })
  }

  return clientInstance
}
