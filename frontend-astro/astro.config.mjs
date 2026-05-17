// @ts-check
import { defineConfig } from 'astro/config'
import { readFileSync } from 'fs'

import tailwindcss from '@tailwindcss/vite'

import react from '@astrojs/react'

// https://astro.build/config
export default defineConfig({
    vite: {
        plugins: [tailwindcss()],
        server: {
            https: {
                key: readFileSync('./key.pem'),
                cert: readFileSync('./cert.pem'),
            },
            host: '0.0.0.0',
            port: 4321,
            proxy: {
                '/api/v1': {
                    target: 'http://backend:8080/',
                },
            },
        },
    },

    integrations: [react()],
})
