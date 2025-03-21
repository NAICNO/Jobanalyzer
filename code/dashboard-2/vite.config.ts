import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'https://naic-monitor.uio.no/reports', // Target API
        changeOrigin: true, // this is necessary to avoid CORS issues
        secure: false, // if you are accessing a https endpoint, this may be necessary
        rewrite: (path) => path.replace(/^\/api/, ''), // Remove "/api" from the path
      },
      '/rest': {
        target: 'https://naic-monitor.uio.no', // Target API
        changeOrigin: true, // this is necessary to avoid CORS issues
        secure: false, // if you are accessing a https endpoint, this may be necessary
        rewrite: (path) => path.replace(/^\/rest/, ''), // Remove "/rest" from the path
      }
    }
  }
})
