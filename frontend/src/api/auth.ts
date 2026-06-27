import ky from 'ky'

const restrictedClient = ky.create({ prefix: '/api/v1' })

export function redirectToGoogleAuth() {
    return window.location.replace(`api/v1/auth/google/sign-in`)
}

export const emailAlreadyTaken = 'email already taken'
export const userNotFound = 'user not found'

export async function signIn(email: string, password: string) {
    await restrictedClient.post('/auth/sign-in', {
        json: { email, password },
    })
}

export async function signUp(email: string, password: string) {
    await restrictedClient.post('/auth/sign-up', {
        json: { email, password },
    })
}

export async function refreshToken() {
    await restrictedClient.post(`/auth/refresh`, {
        json: { refreshToken },
    })
}

export async function signOut() {
    await restrictedClient.post('/auth/logout')
}

export type GoogleCallbackParams = {
    code: string
    state: string
}

export async function googleProcessCallback(params: GoogleCallbackParams) {
    await restrictedClient.post('/auth/google/callback', {
        json: params,
    })
}

export async function requestPasswordReset(email: string): Promise<void> {
    await restrictedClient.post('/auth/password-reset/request', {
        json: { email },
    })
}

export async function confirmPasswordReset(token: string, password: string): Promise<void> {
    await restrictedClient.post('/auth/password-reset/confirm', {
        json: { token, password },
    })
}
