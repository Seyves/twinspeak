import ky from 'ky'

const BACKEND_HOST = import.meta.env.PUBLIC_BACKEND_HOST

const restrictedClient = ky.create({ prefix: `https://${BACKEND_HOST}` })

export function redirectToGoogleAuth() {
    return window.location.replace(`https://${BACKEND_HOST}/auth/google/sign-in`)
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

export async function signOut() {
    await restrictedClient.post('auth/logout')
    window.location.href = '/auth'
}
