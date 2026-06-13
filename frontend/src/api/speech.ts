import type { ChatSide } from '@/definitions/chat'
import { httpClient } from '@/lib/httpClient'

const BACKEND_HOST = import.meta.env.VITE_WS_BACKEND_HOST

async function getWsTicket(): Promise<string> {
    const data = await httpClient.get('/ws-ticket').json<{ ticket: string }>()
    return data.ticket
}

export async function startSession(inLang: string, outLang: string, side: ChatSide) {
    const ticket = await getWsTicket()
    const query = new URLSearchParams()

    query.set('ticket', ticket)
    query.set('inLang', inLang)
    query.set('outLang', outLang)
    query.set('chatSide', side)

    const wsUrl = `${BACKEND_HOST}/ws/session?${query.toString()}`

    return new WebSocket(wsUrl)
}
