export const directions = {
    own: 0,
    companion: 1,
} as const

export type Direction = (typeof directions)[keyof typeof directions]

export const messageStatuses = {
    recording: 'recording',
    pending: 'pending',
    processed: 'processed',
    error: 'error',
} as const

export type MessageStatus = (typeof messageStatuses)[keyof typeof messageStatuses]

export type Message = {
    id: string
    direction: Direction
    status: MessageStatus
    transcription: string
    translation: string
}
