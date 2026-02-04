import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
      imports: ['vue', 'vue-router', 'pinia'],
      dts: 'src/auto-imports.d.ts',
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts',
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3100,
    host: '0.0.0.0', // 允许外部访问（Docker 需要）
    strictPort: false, // 端口被占用时自动尝试下一个
    proxy: {
      '/api': {
        // Docker 环境使用服务名，本地使用 localhost
        target: process.env.DOCKER_ENV ? 'http://backend:10010' : 'http://localhost:10010',
        changeOrigin: true,
        ws: true, // 启用 WebSocket 代理
      },
    },
    watch: {
      usePolling: true, // Docker 环境需要轮询模式
    },
  },
})
