// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  future: {
    compatibilityVersion: 4
  },
  compatibilityDate: '2025-01-01',
  ssr: true,
  devServer: {
    port: 3000,
    host: '0.0.0.0'
  },
  modules: [
    '@nuxtjs/tailwindcss',
    '@pinia/nuxt',
    '@nuxtjs/color-mode',
    '@nuxtjs/i18n'
  ],
  runtimeConfig: {
    internalApiUrl: process.env.NUXT_INTERNAL_API_URL || 'http://localhost:8080',
    public: {
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
      wsBaseUrl: process.env.NUXT_PUBLIC_WS_BASE_URL || 'ws://localhost:8080',
      siteUrl: process.env.NUXT_PUBLIC_SITE_URL || 'https://global-video-platform.dnjstjqppk.workers.dev',
      studioUrl: process.env.NUXT_PUBLIC_STUDIO_URL || 'https://global-video-platform-studio.dnjstjqppk.workers.dev',
      partnersUrl: process.env.NUXT_PUBLIC_PARTNERS_URL || 'https://global-video-platform-partners.dnjstjqppk.workers.dev',
      adminUrl: process.env.NUXT_PUBLIC_ADMIN_URL || 'https://global-video-platform-admin.dnjstjqppk.workers.dev',
      brandName: 'StreamVerse'
    }
  },
  colorMode: {
    classSuffix: '',
    preference: 'dark',
    fallback: 'dark'
  },
  i18n: {
    locales: [
      { code: 'ko', iso: 'ko-KR', name: '한국어' },
      { code: 'en', iso: 'en-US', name: 'English' }
    ],
    defaultLocale: 'ko',
    strategy: 'no_prefix'
  },
  app: {
    head: {
      htmlAttrs: {
        lang: 'ko',
        class: 'dark'
      },
      titleTemplate: '%s — StreamVerse Global Video Platform',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        {
          name: 'description',
          content: 'StreamVerse — Watch 4K HDR videos, Shorts, and ultra-low latency Live streams from global creators.'
        },
        { name: 'theme-color', content: '#0F0F0F' },
        { property: 'og:site_name', content: 'StreamVerse' },
        { property: 'og:type', content: 'website' },
        { name: 'twitter:card', content: 'summary_large_image' }
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }
      ]
    }
  },
  routeRules: {
    '/': { swr: 60 },
    '/watch/**': { ssr: true },
    '/channels/**': { ssr: true },
    '/shorts/**': { ssr: true },
    '/search': { ssr: true }
  },
  typescript: {
    strict: true,
    shim: false
  },
  nitro: {
    preset: process.env.NITRO_PRESET || (process.env.CF_PAGES ? 'cloudflare-pages' : 'cloudflare-module')
  }
})
