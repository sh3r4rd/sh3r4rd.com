import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // The dashboard API only allows the https://sh3r4rd.com origin, so calls
      // made straight from http://localhost:5173 are blocked by CORS. In dev the
      // app targets this relative prefix instead (VITE_API_BASE_URL in
      // .env.development) and the dev server forwards the request server-side,
      // where CORS does not apply. `changeOrigin` rewrites the Host header so
      // the API Gateway custom domain routes the request.
      '/dashboard-api': {
        target: 'https://dashboard-api.sh3r4rd.com',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/dashboard-api/, ''),
      },
      // The resume form's /requests endpoint, on a separate API. That one does
      // allow any origin, so this proxy is for consistency rather than to unblock
      // it — both API bases are wired the same way and configured in one place.
      '/requests-api': {
        target: 'https://api.sh3r4rd.com',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/requests-api/, ''),
      },
    },
  },
})
