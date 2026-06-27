import { getRouter } from '@/router'
import * as AuthApi from '@/api/auth'
import ky from 'ky'

let refreshPromise: Promise<void> | null = null

async function refreshTokenOnce() {
    if (!refreshPromise) {
        refreshPromise = AuthApi.refreshToken().finally(() => {
            refreshPromise = null
        })
    }

    return refreshPromise
}

// Create ky instance with hooks for automatic token handling
export const httpClient = ky.create({
    prefix: `/api/v1`,
    hooks: {
        afterResponse: [
            async ({ request, options, response }) => {
                switch (response.status) {
                    case 401: {
                        const router = getRouter()
                        // Check if this is already a retry to prevent infinite loops
                        if (request.headers.get('X-Retry-After-Refresh')) {
                            router.navigate({ to: '/auth' })
                            return response
                        }

                        try {
                            await refreshTokenOnce()
                        } catch (e) {
                            router.navigate({ to: '/auth' })
                            return response
                        }

                        const retryRequest = new Request(request, {
                            headers: {
                                ...Object.fromEntries(request.headers.entries()),
                                'X-Retry-After-Refresh': 'true',
                            },
                        })
                        return httpClient(retryRequest, options)
                    }
                    case 403: {
                        const clonedResp = response.clone()
                        const body = await clonedResp.text()
                        if (body === 'email not verified') {
                            const router = getRouter()
                            router.navigate({ to: '/verify-email' })
                        }
                        return response
                    }
                }
            },
        ],
    },
})
