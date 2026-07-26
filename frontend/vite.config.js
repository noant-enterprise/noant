import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { VitePWA } from 'vite-plugin-pwa';
import path from 'node:path';
export default defineConfig({
    plugins: [
        react(),
        VitePWA({
            strategies: 'injectManifest',
            srcDir: 'public',
            filename: 'sw.js',
            registerType: 'autoUpdate',
            injectManifest: {
                injectionPoint: undefined,
            },
            manifest: {
                name: 'Noant - AI Customer Support',
                short_name: 'Noant',
                description: 'AI-powered customer support platform for Nigerian businesses',
                theme_color: '#0f172a',
                background_color: '#0f172a',
                display: 'standalone',
                display_override: ['window-controls-overlay', 'standalone'],
                orientation: 'any',
                start_url: '/',
                scope: '/',
                categories: ['business', 'customer support', 'communication'],
                lang: 'en',
                dir: 'ltr',
                prefer_related_applications: false,
                icons: [
                    { src: '/favicon.jpg', sizes: '192x192', type: 'image/jpeg' },
                    { src: '/favicon.jpg', sizes: '512x512', type: 'image/jpeg' },
                    { src: '/favicon.jpg', sizes: '512x512', type: 'image/jpeg', purpose: 'maskable' },
                ],
                shortcuts: [
                    {
                        name: 'Conversations',
                        short_name: 'Chats',
                        url: '/chats',
                        icons: [{ src: '/favicon.jpg', sizes: '192x192' }],
                    },
                    {
                        name: 'Dashboard',
                        short_name: 'Dashboard',
                        url: '/dashboard',
                        icons: [{ src: '/favicon.jpg', sizes: '192x192' }],
                    },
                    {
                        name: 'Settings',
                        short_name: 'Settings',
                        url: '/settings',
                        icons: [{ src: '/favicon.jpg', sizes: '192x192' }],
                    },
                ],
            },
        }),
    ],
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
                configure: function (proxy, _options) {
                    proxy.on('error', function (err, _req, _res) {
                        // Suppress ECONNRESET spam — only log real errors
                        if (err.code !== 'ECONNRESET') {
                            console.warn('Proxy error:', err.message);
                        }
                    });
                    proxy.on('proxyReq', function (proxyReq, req, _res) {
                        console.log('→', req.method, req.url, '→', proxyReq.path);
                    });
                },
            },
            '/ws': {
                target: 'ws://127.0.0.1:8080',
                ws: true,
                changeOrigin: true,
                secure: false,
                configure: function (proxy, _options) {
                    proxy.on('error', function (err, _req, _res) {
                        // Silently ignore ECONNRESET on WebSocket — it's normal during server restart
                        if (err.code !== 'ECONNRESET') {
                            console.warn('WS Proxy error:', err.message);
                        }
                    });
                    proxy.on('proxyReqWs', function (_proxyReq, req, _socket, _options, _head) {
                        console.log('WS →', req.url);
                    });
                },
            },
        },
    },
});
