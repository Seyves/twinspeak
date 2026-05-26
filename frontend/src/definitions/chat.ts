export const chatSide = {
    bottom: 'bottom',
    top: 'top',
} as const

export type ChatSide = (typeof chatSide)[keyof typeof chatSide]

export function reverseSide(size: ChatSide) {
    switch (size) {
        case chatSide.top:
            return chatSide.bottom
        case chatSide.bottom:
            return chatSide.top
    }
}

export const messageStatuses = {
    recording: 'recording',
    processed: 'processed',
    error: 'error',
} as const

export type MessageStatus = (typeof messageStatuses)[keyof typeof messageStatuses]

export type Message = {
    id: string
    sendedFrom: ChatSide
    status: MessageStatus
    transcription: string
    translation: string
}
