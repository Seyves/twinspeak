import ky from 'ky'

const BACKEND_HOST = import.meta.env.VITE_HTTP_BACKEND_HOST

const restrictedClient = ky.create({ prefix: BACKEND_HOST })

export function redirectToGoogleAuth() {
    return window.location.replace(`${BACKEND_HOST}/google-sign-in`)
}

export type GoogleCallbackParams = {
    code: string
    state: string
}

export async function googleProcessCallback(params: GoogleCallbackParams) {
    await restrictedClient.post('auth/google/callback', {
        json: params,
    })
}

export async function signIn(email: string, password: string) {
    await restrictedClient.post('auth/sign-in', {
        json: { email, password },
    })
}

export async function signUp(email: string, password: string) {
    await restrictedClient.post('auth/sign-up', {
        json: { email, password },
    })
}

export async function refreshToken() {
    await restrictedClient.post(`/auth/refresh`, {
        json: { refreshToken },
    })
}

export async function signOut() {
    await restrictedClient.post('auth/logout')
}
