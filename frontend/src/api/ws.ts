import type { ChatSide } from '@/definitions/chat'
import { httpClient } from '@/lib/httpClient'

async function getTicket(): Promise<string> {
    const data = await httpClient.get('/ws/ticket').json<{ ticket: string }>()
    return data.ticket
}

export async function startSession(inLang: string, outLang: string, side: ChatSide) {
    const ticket = await getTicket()
    const query = new URLSearchParams()

    query.set('ticket', ticket)
    query.set('inLang', inLang)
    query.set('outLang', outLang)
    query.set('chatSide', side)

    return new WebSocket(`/api/v1/ws/session?${query.toString()}`)
}
