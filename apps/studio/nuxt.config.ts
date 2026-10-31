export default defineNuxtConfig({
  future: {
    compatibilityVersion: 4
  },
  compatibilityDate: '2025-01-01',
  ssr: true,
  devServer: {
    port: 3001,
    host: '0.0.0.0'
  },
  modules: [
    '@nuxtjs/tailwindcss',
    '@pinia/nuxt'
  ],
  runtimeConfig: {
    internalApiUrl: process.env.NUXT_INTERNAL_API_URL || 'http://localhost:8080',
    public: {
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || '/api',
      wsBaseUrl: process.env.NUXT_PUBLIC_WS_BASE_URL || '',
      webUrl: process.env.NUXT_PUBLIC_WEB_URL || 'https://global-video-platform.dnjstjqppk.workers.dev',
      partnersUrl: process.env.NUXT_PUBLIC_PARTNERS_URL || 'https://global-video-platform-partners.dnjstjqppk.workers.dev',
      adminUrl: process.env.NUXT_PUBLIC_ADMIN_URL || 'https://global-video-platform-admin.dnjstjqppk.workers.dev',
      studioDomain: 'global-video-platform-studio.dnjstjqppk.workers.dev'
    }
  },
  app: {
    head: {
      htmlAttrs: { lang: 'ko', class: 'dark' },
      titleTemplate: '%s — StreamVerse Creator Studio',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'robots', content: 'noindex, nofollow' }
      ]
    }
  },
  typescript: {
    strict: true,
    shim: false
  },
  nitro: {
    preset: process.env.NITRO_PRESET || (process.env.CF_PAGES ? 'cloudflare-pages' : 'cloudflare-module')
  }
})
