import { httpClient } from '@/lib/httpClient'

export type ProcessSpeechResponse = {
    id: string
    transcription: string
    translation: string
}

export async function processSpeech(
    blob: Blob,
    inputLang: string,
    outputLang: string
): Promise<ProcessSpeechResponse> {
    const formData = new FormData()
    formData.append('audio_file', blob)

    return httpClient
        .post('speech/process', {
            searchParams: {
                inputLang,
                outputLang,
            },
            body: formData,
        })
        .json<ProcessSpeechResponse>()
}
