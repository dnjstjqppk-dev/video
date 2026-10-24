import { createApiClient, type ApiClient } from '@gvp/ts-api-client'

let adminClient: ApiClient | null = null

export function useAdminApi(): ApiClient {
  const config = useRuntimeConfig()
  const authCookie = useCookie<string | null>('gvp_admin_session', {
    sameSite: 'strict',
    maxAge: 900 // 15-minute strict admin session TTL
  })

  if (!adminClient || import.meta.server) {
    const baseUrl = import.meta.server
      ? (config.internalApiUrl as string) || (config.public.apiBaseUrl as string)
      : (config.public.apiBaseUrl as string)

    adminClient = createApiClient({
      baseUrl,
      getToken: () => authCookie.value ?? null
    })
  }

  return adminClient
}
