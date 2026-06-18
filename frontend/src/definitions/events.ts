export const eventTypes = {
    liveTranscript: 0,
    liveTranslate: 1,
    finalTranscript: 2,
    finalTranslate: 3,
    error: 4,
} as const

export type EventType = (typeof eventTypes)[keyof typeof eventTypes]

type EventT<T, P> = {
    type: T
    payload: P
}

export type LiveTranscriptEvent = EventT<0, string>
export type LiveTranslateEvent = EventT<1, string>
export type FinalTranscriptEvent = EventT<2, string>
export type FinalTranslateEvent = EventT<3, string>
export type DurationEvent = EventT<4, number>

export type SpeechEvent =
    | LiveTranscriptEvent
    | LiveTranslateEvent
    | FinalTranscriptEvent
    | FinalTranslateEvent
    | DurationEvent
