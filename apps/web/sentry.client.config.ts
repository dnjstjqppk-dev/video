import * as Sentry from '@sentry/nuxt'

const dsn = useRuntimeConfig().public.sentryDsn as string | undefined

Sentry.init({
  dsn: dsn || undefined,
  tracesSampleRate: 0.1,
  replaysSessionSampleRate: 0,
  replaysOnErrorSampleRate: 0,
})
