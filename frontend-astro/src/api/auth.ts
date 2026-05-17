// const backendUrl = import.meta.env.PUBLIC_BACKEND_URL as string
const backendUrl = "https://localhost:4321/api/v1"
const TOKEN_KEY = "authToken"

export function googleSignIn() {
    return window.location.replace(`${backendUrl}/auth/google/sign-in`)
}

type GoogleProcessCallbackResp = {
    accessToken: string
}

export async function googleProcessCallback(code: string, state: string) {
    const resp = await fetch(`${backendUrl}/auth/google/callback`, {
        method: 'POST',
        body: JSON.stringify({
            code,
            state,
        }),
    })

    return (await resp.json()) as GoogleProcessCallbackResp
}

export function setAuthToken(token: string): void {
    localStorage.setItem(TOKEN_KEY, token)
}

export function getAuthToken(): string | null {
    return localStorage.getItem(TOKEN_KEY)
}

export function clearAuthToken(): void {
    localStorage.removeItem(TOKEN_KEY)
}
