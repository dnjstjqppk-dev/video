import { createApiClient, type ApiClient } from '@gvp/ts-api-client'

let partnerClient: ApiClient | null = null

export function usePartnerApi(): ApiClient {
  const config = useRuntimeConfig()
  const authCookie = useCookie<string | null>('gvp_partner_token', {
    sameSite: 'lax'
  })

  if (!partnerClient || import.meta.server) {
    const baseUrl = import.meta.server
      ? (config.internalApiUrl as string) || (config.public.apiBaseUrl as string)
      : (config.public.apiBaseUrl as string)

    partnerClient = createApiClient({
      baseUrl,
      getToken: () => authCookie.value ?? null
    })
  }

  return partnerClient
}
