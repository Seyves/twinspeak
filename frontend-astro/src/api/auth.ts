import ky from 'ky'

const BACKEND_URL = import.meta.env.PUBLIC_BACKEND_URL

const restrictedClient = ky.create({ prefix: BACKEND_URL })

export function redirectToGoogleAuth() {
    return window.location.replace(`${BACKEND_URL}/auth/google/sign-in`)
}

export async function googleProcessCallback(code: string, state: string) {
    await restrictedClient.post('auth/google/callback', {
        json: {
            code,
            state,
        },
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
