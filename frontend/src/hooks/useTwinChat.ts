import { messageStatuses } from '@/definitions/chat'
import type { ChatSide, Message } from '@/definitions/chat'
import { useLayoutEffect, useState } from 'react'

export type MessageController = (callback: (prev: Message) => Message) => void

const maxMessages = 10

export function useTwinChat() {
    const [messages, setMessages] = useState<Message[]>([])

    useLayoutEffect(() => {
        initMessages()
    }, [])

    function initMessages() {
        const rawMessages = localStorage.getItem('messages') ?? '[]'
        const cachedMessages = JSON.parse(rawMessages) as Message[]
        setMessages(
            cachedMessages.map((msg) => {
                msg.status = messageStatuses.processed
                return msg
            }),
        )
    }

    function startRecordingMsg(side: ChatSide) {
        const id = crypto.randomUUID()

        setMessages((prev) => [
            ...limitMessages(prev),
            {
                id: id,
                sendedFrom: side,
                status: messageStatuses.recording,
                transcription: '',
                translation: '',
            },
        ])
        return function setMessage(c: (prev: Message) => Message) {
            return setMessages((prev) => {
                const result = prev.map((msg) => {
                    if (msg.id !== id) return msg
                    return c(msg)
                })
                localStorage.setItem('messages', JSON.stringify(result))
                return result
            })
        }
    }

    function errorProcessingMsg(side: ChatSide) {
        setMessages((prev) => {
            const processingIdx = prev.findIndex((msg) => {
                return msg.status === messageStatuses.pending && msg.sendedFrom === side
            })
            const msg = prev[processingIdx]
            prev = [...prev]
            prev[processingIdx] = {
                ...msg,
                transcription: '',
                translation: '',
                status: messageStatuses.error,
            }
            return prev
        })
    }

    return {
        messages,
        startRecordingMsg,
        errorProcessingMsg,
    }
}

function limitMessages(messages: Message[]) {
    const limit = maxMessages - 1
    if (messages.length > limit) {
        return messages.slice(messages.length - limit)
    }
    return messages
}
