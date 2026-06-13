import type { ChatMessageSize, Message, Theme } from '@/definitions/chat'
import { httpClient } from '@/lib/httpClient'

export type Preferences = {
    chatMessageSize: ChatMessageSize
    theme: Theme
    inLang: string
    outLang: string
}

export async function ping() {
    return httpClient.get('/ping')
}

export async function getSupportedLanguages() {
    return await httpClient.get('/supported-languages').json<Record<string, string>>()
}

export async function getPreferences() {
    return await httpClient.get('/preferences').json<Preferences>()
}

export async function getMessages() {
    return await httpClient.get('/messages').json<Message[]>()
}

export async function updatePreferences(prefs: Preferences) {
    return await httpClient.put('/preferences', {
        body: JSON.stringify(prefs),
    })
}
