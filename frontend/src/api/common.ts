import { httpClient } from '@/lib/httpClient'

export async function ping() {
    return httpClient.get('/ping')
}

export async function getSupportedLanguages() {
    return await httpClient.get('/supported-languages').json<Record<string, string>>()
}
