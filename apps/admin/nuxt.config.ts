export default defineNuxtConfig({
  future: {
    compatibilityVersion: 4
  },
  compatibilityDate: '2025-01-01',
  ssr: true,
  devServer: {
    port: 3003,
    host: '0.0.0.0'
  },
  modules: [
    '@nuxtjs/tailwindcss',
    '@pinia/nuxt'
  ],
  runtimeConfig: {
    internalApiUrl: process.env.NUXT_INTERNAL_API_URL || 'http://localhost:8080',
    adminSessionTimeoutMinutes: 15,
    requireWebAuthnMfa: true,
    public: {
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
      adminDomain: 'admin.gvp.media',
      securityPolicy: 'ZERO_TRUST_STEP_UP_2FA'
    }
  },
  app: {
    head: {
      htmlAttrs: { lang: 'ko', class: 'dark' },
      titleTemplate: '%s — StreamVerse Super Admin Governance Console',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'robots', content: 'noindex, nofollow, noarchive' }
      ]
    }
  },
  typescript: {
    strict: true,
    shim: false
  }
})
