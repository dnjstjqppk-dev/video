import * as Sentry from '@sentry/nuxt'

const config = useRuntimeConfig()

Sentry.init({
  dsn: (config.sentryDsn as string | undefined) || undefined,
  tracesSampleRate: 0.1,
})
