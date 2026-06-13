import { defineConfig } from 'vite'
import { devtools } from '@tanstack/devtools-vite'

import { tanstackStart } from '@tanstack/react-start/plugin/vite'

import viteReact from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { nitro } from 'nitro/vite'
import { readFileSync, existsSync } from 'fs'

const httpsConfig =
    existsSync('./key.pem') && existsSync('./cert.pem')
        ? { key: readFileSync('./key.pem'), cert: readFileSync('./cert.pem') }
        : undefined

const config = defineConfig({
    resolve: {
        tsconfigPaths: true,
    },
    plugins: [
        devtools(),
        nitro({
            rollupConfig: {
                external: [/^@sentry\//],
            },
            routeRules: {
                '/api/v1/**': {
                    proxy: {
                        to: 'http://backend:8080/api/v1/**',
                        fetchOptions: { redirect: 'manual' },
                    },
                },
            },
        }),
        tailwindcss(),
        tanstackStart(),
        viteReact(),
    ],
    server: {
        // https: httpsConfig,
        host: '0.0.0.0',
        port: 4321,
        proxy: {
            '/api/v1/ws': {
                target: 'ws://backend:8080',
                ws: true,
                changeOrigin: true,
            },
        },
    },
})

export default config
