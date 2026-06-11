import { getRouter } from '@/router'
import { refreshToken } from '@/api/auth'
import ky, { HTTPError } from 'ky'

const BACKEND_HOST = import.meta.env.VITE_HTTP_BACKEND_HOST

let refreshPromise: Promise<void> | null = null

async function refreshTokenOnce() {
    if (!refreshPromise) {
        refreshPromise = refreshToken().finally(() => {
            refreshPromise = null
        })
    }

    return refreshPromise
}

// Create ky instance with hooks for automatic token handling
export const httpClient = ky.create({
    prefix: `https://${BACKEND_HOST}`,
    hooks: {
        afterResponse: [
            async ({ request, options, response }) => {
                if (response.status !== 401) {
                    return response
                }

                try {
                    await refreshTokenOnce()

                    return ky(request, options)
                } catch (e) {
                    if (e instanceof HTTPError && e.response.status === 401) {
                        const router = getRouter()
                        router.navigate({ to: '/auth' })
                        return
                    }
                    throw e
                }
            },
        ],
    },
})
