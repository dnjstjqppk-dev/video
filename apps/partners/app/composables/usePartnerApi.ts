import { createApiClient, type ApiClient } from '@gvp/ts-api-client'

let partnerClient: ApiClient | null = null

export function usePartnerApi(): ApiClient {
  const config = useRuntimeConfig()
  const authCookie = useCookie<string | null>('gvp_partner_token', {
    sameSite: 'lax'
  })
  // The httpOnly session cookie is readable here on the server only, which
  // authenticates SSR API calls; in the browser the /api proxy injects it.
  const sessionCookie = useCookie<string | null>('gvp_session')

  if (!partnerClient || import.meta.server) {
    const baseUrl = import.meta.server
      ? (config.internalApiUrl as string) || (config.public.apiBaseUrl as string)
      : (config.public.apiBaseUrl as string)

    partnerClient = createApiClient({
      baseUrl,
      getToken: () => authCookie.value ?? sessionCookie.value ?? null
    })
  }

  return partnerClient
}
