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
  css: ['@gvp/design-tokens/src/gvp.css'],
  runtimeConfig: {
    internalApiUrl: process.env.NUXT_INTERNAL_API_URL || 'http://localhost:8080',
    oidc: {
      issuer: process.env.NUXT_OIDC_ISSUER || '',
      clientId: process.env.NUXT_OIDC_CLIENT_ID || '',
      clientSecret: process.env.NUXT_OIDC_CLIENT_SECRET || ''
    },
    adminSessionTimeoutMinutes: 15,
    requireWebAuthnMfa: true,
    public: {
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || '/api',
      webUrl: process.env.NUXT_PUBLIC_WEB_URL || 'https://global-video-platform.dnjstjqppk.workers.dev',
      studioUrl: process.env.NUXT_PUBLIC_STUDIO_URL || 'https://global-video-platform-studio.dnjstjqppk.workers.dev',
      partnersUrl: process.env.NUXT_PUBLIC_PARTNERS_URL || 'https://global-video-platform-partners.dnjstjqppk.workers.dev',
      adminDomain: 'global-video-platform-admin.dnjstjqppk.workers.dev',
      securityPolicy: 'ZERO_TRUST_STEP_UP_2FA'
    }
  },
  app: {
    head: {
      htmlAttrs: { lang: 'ko' },
      titleTemplate: '%s — StreamVerse Super Admin Governance Console',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'robots', content: 'noindex, nofollow, noarchive' }
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'preconnect', href: 'https://cdn.jsdelivr.net' },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        { rel: 'stylesheet', href: 'https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/variable/pretendardvariable-dynamic-subset.min.css' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600;700&display=swap' }
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
