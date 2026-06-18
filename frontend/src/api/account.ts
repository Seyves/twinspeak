import type { ChatMessageSize, Message, Theme } from '@/definitions/chat'
import { httpClient } from '@/lib/httpClient'

export type UserInfo = {
    email: string
    profilePicture: string | null
    createdAt: string
}

export type CreditGrant = {
    id: string
    amount: number
    remainingAmount: number
    type: 'monthly' | 'topup'
    expiresAt: string | null
    createdAt: string
}

export type Preferences = {
    chatMessageSize: ChatMessageSize
    theme: Theme
    inLang: string
    outLang: string
}

export async function getAccount(): Promise<UserInfo> {
    const response = await httpClient.get('/account')
    return response.json()
}

export async function getPreferences() {
    return await httpClient.get('/account/preferences').json<Preferences>()
}

export async function updatePreferences(prefs: Preferences) {
    return await httpClient.put('/account/preferences', {
        body: JSON.stringify(prefs),
    })
}

export async function getMessages() {
    return await httpClient.get('/account/messages').json<Message[]>()
}

export async function clearChat() {
    return await httpClient.post('/account/clear-chat')
}

export async function getCredits(): Promise<CreditGrant[]> {
    const response = await httpClient.get('/account/credits')
    return response.json()
}
