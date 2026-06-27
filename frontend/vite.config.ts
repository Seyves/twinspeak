import { defineConfig } from 'vite'
import { devtools } from '@tanstack/devtools-vite'

import { tanstackStart } from '@tanstack/react-start/plugin/vite'

import viteReact from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { nitro } from 'nitro/vite'

const config = defineConfig(() => {
    const backendHost = process.env.VITE_BACKEND_HOST
    console.log('Building with backend host:', backendHost)

    return {
        resolve: {
            tsconfigPaths: true,
        },
        plugins: [
            devtools(),
            nitro({
                $development: {
                    routeRules: {
                        '/api/v1/**': {
                            proxy: {
                                to: `http://${backendHost}/api/v1/**`,
                                fetchOptions: { redirect: 'manual' },
                            },
                        },
                    },
                },
            }),
            tailwindcss(),
            tanstackStart(),
            viteReact(),
        ],
        // Server for dev mode
        server: {
            host: '0.0.0.0',
            port: 4321,
            proxy: {
                '/api/v1/ws/session': {
                    target: `ws://${backendHost}`,
                    ws: true,
                    changeOrigin: true,
                },
            },
        },
    }
})

export default config
