import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import fs from 'node:fs'
import path from 'node:path'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd())
  const isDemoMode = env.VITE_DEMO_MODE === 'true'

  return {
  plugins: [
    react(),
    !isDemoMode && {
      name: 'remove-demo-assets',
      apply: 'build' as const,
      writeBundle(options) {
        const outDir = options.dir ?? 'dist'
        for (const file of ['mockServiceWorker.js', 'clusters.demo.json']) {
          fs.rmSync(path.join(outDir, file), { force: true })
        }
      },
    },
  ].filter(Boolean),
  define: mode === 'production' ? { 'console.log': '(() => {})' } : {},
  build: {
    rollupOptions: {
      input: {
        main: 'index.html',
        'silent-renew': 'silent-renew.html',
      },
    },
  },
  server: {
    proxy: {
      '/api': {
        target: 'https://naic-monitor.uio.no/reports',
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
      '/rest': {
        target: 'https://naic-monitor.uio.no',
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path.replace(/^\/rest/, ''),
      }
    }
  }
  }
}))
