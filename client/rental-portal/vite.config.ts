import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      // Важно: /api/ws должен быть выше /api, иначе WebSocket уходит на userservice (8052) вместо координатора (8090)
      '/api/ws': {
        target: 'http://localhost:8090',
        changeOrigin: true,
        ws: true,
        rewrite: (path) => path.replace(/^\/api\/ws/, '/api/v1/ws'),
      },
      '/api': {
        target: 'http://localhost:8052',
        changeOrigin: true,
      },
    },
  },
})
