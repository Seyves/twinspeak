import { messageStatuses } from '@/definitions/messages'
import type { Direction, Message } from '@/definitions/messages'
import { useState } from 'react'

export function useTwinChat() {
    const [messages, setMessages] = useState<Message[]>([])

    function startRecordingMsg(direction: Direction): void {
        setMessages((prev) => [
            ...prev,
            {
                id: crypto.randomUUID(),
                direction: direction,
                status: messageStatuses.recording,
                transcription: '',
                translation: '',
            },
        ])
    }

    function startProcessingMsg(direction: Direction): void {
        setMessages((prev) => {
            const recordingIdx = prev.findIndex((msg) => {
                return msg.status === messageStatuses.recording && msg.direction === direction
            })
            const msg = prev[recordingIdx]
            prev = [...prev]
            prev[recordingIdx] = {
                ...msg,
                status: messageStatuses.pending,
            }
            return prev
        })
    }

    function successProcessingMsg(direction: Direction, transcribed: string, translated: string) {
        setMessages((prev) => {
            const processingIdx = prev.findIndex((msg) => {
                return msg.status === messageStatuses.pending && msg.direction === direction
            })
            const msg = prev[processingIdx]
            prev = [...prev]
            prev[processingIdx] = {
                ...msg,
                transcription: transcribed,
                translation: translated,
                status: messageStatuses.processed,
            }
            return prev
        })
    }

    function errorProcessingMsg(direction: Direction) {
        setMessages((prev) => {
            const processingIdx = prev.findIndex((msg) => {
                return msg.status === messageStatuses.pending && msg.direction === direction
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
        startProcessingMsg,
        successProcessingMsg,
        errorProcessingMsg,
    }
}
