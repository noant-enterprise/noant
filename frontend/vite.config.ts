import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';

export default defineConfig({
  plugins: [react()],
  optimizeDeps: {
    entries: ['index.html'],
  },
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
        secure: false,
        configure: (proxy, _options) => {
          proxy.on('error', (err, _req, _res) => {
            // Suppress ECONNRESET spam — only log real errors
            if ((err as any).code !== 'ECONNRESET') {
              console.warn('Proxy error:', err.message);
            }
          });
          proxy.on('proxyReq', (proxyReq, req, _res) => {
            console.log('→', req.method, req.url, '→', proxyReq.path);
          });
        },
      },
      '/ws': {
        target: 'ws://127.0.0.1:8080',
        ws: true,
        changeOrigin: true,
        secure: false,
        configure: (proxy, _options) => {
          proxy.on('error', (err, _req, _res) => {
            // Silently ignore ECONNRESET on WebSocket — it's normal during server restart
            if ((err as any).code !== 'ECONNRESET') {
              console.warn('WS Proxy error:', err.message);
            }
          });
          proxy.on('proxyReqWs', (_proxyReq, req, _socket, _options, _head) => {
            console.log('WS →', req.url);
          });
        },
      },
    },
  },
});