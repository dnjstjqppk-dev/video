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
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
      webUrl: process.env.NUXT_PUBLIC_WEB_URL || 'http://localhost:3000',
      studioDomain: 'studio.gvp.media'
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
  }
})
