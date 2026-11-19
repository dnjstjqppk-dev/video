import { createApiClient, type ApiClient } from '@gvp/ts-api-client'

let studioApiClient: ApiClient | null = null

export function useStudioApi(): ApiClient {
  const config = useRuntimeConfig()
  const authCookie = useCookie<string | null>('gvp_studio_token', {
    sameSite: 'lax'
  })
  // The httpOnly session cookie is readable here on the server only, which
  // authenticates SSR API calls; in the browser the /api proxy injects it.
  const sessionCookie = useCookie<string | null>('gvp_session')

  if (!studioApiClient || import.meta.server) {
    const baseUrl = import.meta.server
      ? (config.internalApiUrl as string) || (config.public.apiBaseUrl as string)
      : (config.public.apiBaseUrl as string)

    studioApiClient = createApiClient({
      baseUrl,
      getToken: () => authCookie.value ?? sessionCookie.value ?? null
    })
  }

  return studioApiClient
}
