const BACKEND_URL = 'https://localhost:4321/api/v1'
const TOKEN_KEY = 'authToken'

function getToken(): string | null {
    return localStorage.getItem(TOKEN_KEY)
}

function setToken(token: string): void {
    localStorage.setItem(TOKEN_KEY, token)
}

async function refreshToken(): Promise<string | null> {
    const token = getToken()
    if (!token) return null

    try {
        const resp = await fetch(`${BACKEND_URL}/auth/refresh`, {
            method: 'POST',
        })

        if (resp.ok) {
            const data = await resp.json()
            setToken(data.accessToken)
            return data.accessToken
        }
    } catch (error) {
        console.error('Token refresh failed:', error)
    }

    return null
}

export async function httpClient(url: string, options: RequestInit = {}): Promise<Response> {
    const token = getToken()
    const headers = new Headers(options.headers || {})

    if (token) {
        headers.set('Authorization', `Bearer ${token}`)
    }

    let response = await fetch(url, {
        ...options,
        headers,
    })

    // Handle 401: refresh token and retry once
    if (response.status === 401 && token) {
        const newToken = await refreshToken()
        if (newToken) {
            headers.set('Authorization', `Bearer ${newToken}`)
            response = await fetch(url, {
                ...options,
                headers,
            })
        }
    }

    return response
}
