import { refreshToken } from '@/api/auth'
import ky, { HTTPError } from 'ky'

const BACKEND_URL = import.meta.env.PUBLIC_BACKEND_URL

// Create ky instance with hooks for automatic token handling
export const httpClient = ky.create({
    prefix: BACKEND_URL,
    hooks: {
        beforeRetry: [
            async ({ error }) => {
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
