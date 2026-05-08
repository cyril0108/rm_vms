import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import mkcert from 'vite-plugin-mkcert'

// https://vite.dev/config/
export default defineConfig({
  base: "/web/",
  plugins: [
    vue(),
    vueDevTools(),
    mkcert(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    https: true,
    proxy: {
      '/web/config.js': {
        target: 'https://localhost:59180',
        changeOrigin: true,
        configure: (proxy, options) => {
                  proxy.on('proxyReq', (proxyReq, req, _res) => {
                    console.log('Sending Request to the Target:', req.method, req.url);
                  });
                  proxy.on('proxyRes', (proxyRes, req, _res) => {
                    console.log('Received Response from the Target:', proxyRes.statusCode, req.url);
                  });
                },
      },
      // Bonus: Proxy all your API calls to avoid CORS issues in dev mode
      '/api': {
        target: 'https://localhost:59180',
        changeOrigin: true,
      }
    }
  }
})
