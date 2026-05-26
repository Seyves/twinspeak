import { refreshToken } from '@/api/auth'
import ky, { HTTPError } from 'ky'

const BACKEND_HOST = import.meta.env.PUBLIC_BACKEND_HOST

// Create ky instance with hooks for automatic token handling
export const httpClient = ky.create({
    prefix: `https://${BACKEND_HOST}`,
    retry: {
        shouldRetry: ({ error }) => {
            return error instanceof HTTPError && error.response.status === 401
        },
    },
    hooks: {
        beforeRetry: [
            async ({ error }) => {
                console.log(error)
                if (error instanceof HTTPError && error.response.status === 401) {
                    try {
                        await refreshToken()
                    } catch (e) {
                        window.location.href = '/auth'
                    }
                }
            },
        ],
    },
})
