import { createApiClient, type ApiClient } from '@gvp/ts-api-client'

let studioApiClient: ApiClient | null = null

export function useStudioApi(): ApiClient {
  const config = useRuntimeConfig()
  const authCookie = useCookie<string | null>('gvp_studio_token', {
    sameSite: 'lax'
  })

  if (!studioApiClient || import.meta.server) {
    const baseUrl = import.meta.server
      ? (config.internalApiUrl as string) || (config.public.apiBaseUrl as string)
      : (config.public.apiBaseUrl as string)

    studioApiClient = createApiClient({
      baseUrl,
      getToken: () => authCookie.value ?? null
    })
  }

  return studioApiClient
}
