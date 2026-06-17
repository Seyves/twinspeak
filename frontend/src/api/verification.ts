import { httpClient } from '@/lib/httpClient'

export async function resend(): Promise<void> {
    await httpClient.post('/verification/resend')
}

export async function verify(token: string): Promise<void> {
    await httpClient.get('/verification/verify', {
        searchParams: { token },
    })
}
