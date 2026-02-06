import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@/core': path.resolve(__dirname, './src/core'),
      '@/features': path.resolve(__dirname, './src/features'),
      '@/shared': path.resolve(__dirname, './src/shared'),
      '@/infrastructure': path.resolve(__dirname, './src/infrastructure'),
    },
  },
  server: {
    port: 5173,
    host: '0.0.0.0',
    strictPort: true, // ポートが使用中の場合エラーにする
    watch: {
      usePolling: true, // Dockerコンテナ内でのファイル監視を改善
    },
    hmr: {
      host: 'localhost', // HMR用のホスト
    },
    proxy: {
      '/api': {
        target: 'http://backend:8080', // Dockerネットワーク内のサービス名を使用
        changeOrigin: true,
      },
    },
  },
  define: {
    'process.env': JSON.stringify(process.env),
  },
  base: '/', // ベースパスを明示的に設定
});
