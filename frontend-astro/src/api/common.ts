import { httpClient } from '@/lib/httpClient'

export async function ping() {
    return httpClient.get('/ping', {
        retry: {
            shouldRetry: () => true,
            limit: 3,
        },
    })
}
