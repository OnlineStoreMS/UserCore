import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// 启动前设置：VITE_API_GATEWAY=http://localhost:8088 npm run dev
const apiTarget = process.env.VITE_API_GATEWAY || 'http://localhost:8091'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5174,
    proxy: {
      '/api': {
        target: apiTarget,
        changeOrigin: true,
      },
    },
  },
})
