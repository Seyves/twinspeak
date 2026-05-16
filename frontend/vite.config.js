import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import tailwindcss from '@tailwindcss/vite'
import { readFileSync } from 'fs'

export default defineConfig({
    plugins: [react(), tailwindcss()],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
        },
    },
    server: {
        https: {
            key: readFileSync('./key.pem'),
            cert: readFileSync('./cert.pem'),
        },
        host: '0.0.0.0',
        port: 5173,
        proxy: {
            '/process-speech': {
                target: 'http://backend:8080/',
            },
        },
    },
})
