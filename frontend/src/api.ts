type ProcessSpeechResponse = {
    id: string
    transcription: string
    translation: string
}

export async function processSpeech(blob: Blob, inputLang: string, outputLang: string) {
    const formData = new FormData()
    formData.append('audio_file', blob)

    const resp = await fetch(`/process-speech/?inputLang=${inputLang}&outputLang=${outputLang}`, {
        method: 'POST',
        body: formData,
    })

    return await resp.json() as ProcessSpeechResponse
}
