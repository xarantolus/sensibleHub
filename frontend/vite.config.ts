import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'
import { defineConfig } from 'vitest/config'

const backend = process.env.SH_BACKEND ?? 'http://localhost:128'

export default defineConfig({
  plugins: [
    vue(),
    VitePWA({
      strategies: 'injectManifest',
      srcDir: 'src/sw',
      filename: 'sw.ts',
      injectRegister: false,
      injectManifest: {
        globPatterns: ['**/*.{js,css,html,svg,png,ico,woff2}'],
        globIgnores: ['fav/apple-touch-icon-*', 'fav/mstile-*'],
      },
      manifest: {
        name: 'Sensible Hub',
        short_name: 'Sensible Hub',
        description: 'Your music collection',
        display: 'standalone',
        start_url: '/',
        theme_color: '#baffda',
        background_color: '#ffffff',
        icons: [
          { src: '/fav/android-chrome-192x192.png', sizes: '192x192', type: 'image/png' },
          { src: '/fav/android-chrome-512x512.png', sizes: '512x512', type: 'image/png' },
          { src: '/fav/android-chrome-512x512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    proxy: {
      '/api': backend,
      '/media': backend,
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  test: {
    environment: 'jsdom',
  },
})
