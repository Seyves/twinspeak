import { httpClient } from '@/lib/httpClient'

const BACKEND_HOST = import.meta.env.PUBLIC_BACKEND_HOST

async function getWsTicket(): Promise<string> {
    const data = await httpClient.get('/ws-ticket').json<{ ticket: string }>()
    return data.ticket
}

export async function startSession(inLang: string, outLang: string) {
    const ticket = await getWsTicket()
    const query = new URLSearchParams() 

    query.set("inLang", inLang)
    query.set("outLang", outLang)
    query.set("ticket", ticket)

    const wsUrl = `wss://${BACKEND_HOST}/ws/session?${query.toString()}`

    return new WebSocket(wsUrl)
}
