import { chatSide, type ChatSide, type Message } from '@/definitions/chat'
import { useCallback, useEffect, useRef, useState } from 'react'
import { eventTypes, type SpeechEvent } from '@/definitions/events'
import * as SpeechApi from '@/api/ws'

type UseGladiaRecorderOptions = {
    sampleRate?: number
    numChannels?: number
    chunkSize?: number
}

export function useRecorder({
    sampleRate = 16000,
    numChannels = 1,
}: UseGladiaRecorderOptions = {}) {
    const [recordingSide, setRecordingSide] = useState<ChatSide | null>(null)
    const [recordingTime, setRecordingTime] = useState(0)

    const audioContextRef = useRef<AudioContext | null>(null)
    const streamRef = useRef<MediaStream | null>(null)
    const sourceRef = useRef<MediaStreamAudioSourceNode | null>(null)
    const workletNodeRef = useRef<AudioWorkletNode | null>(null)

    const wsRef = useRef<WebSocket | null>(null)
    const isRecordingRef = useRef(false)

    const timerRef = useRef<number | null>(null)
    const startedAtRef = useRef<number>(0)

    const cleanup = useCallback(async () => {
        try {
            sourceRef.current?.disconnect()
            workletNodeRef.current?.disconnect()

            streamRef.current?.getTracks().forEach((track) => {
                track.stop()
            })

            if (audioContextRef.current) {
                await audioContextRef.current.close()
            }
        } catch (err) {
            console.error('Cleanup error:', err)
        }

        sourceRef.current = null
        workletNodeRef.current = null
        streamRef.current = null
        audioContextRef.current = null

        if (timerRef.current) {
            clearInterval(timerRef.current)
            timerRef.current = null
        }
    }, [])

    useEffect(() => {
        return () => {
            cleanup()
        }
    }, [cleanup])

    const startRecording = useCallback(
        async (
            recordingSide: ChatSide,
            ownerLang: string,
            companionLang: string,
            setRecordingMessage: (callback: (prev: Message) => Message) => void,
        ) => {
            if (isRecordingRef.current) {
                return
            }

            const [inLang, outLang] =
                recordingSide === chatSide.bottom
                    ? [ownerLang, companionLang]
                    : [companionLang, ownerLang]

            const ws = await SpeechApi.startSession(inLang, outLang, recordingSide)
            ws.binaryType = 'arraybuffer'

            ws.onopen = async () => {
                console.log('WebSocket connected')
                wsRef.current = ws
            }

            ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data) as SpeechEvent
                    switch (message.type) {
                        case eventTypes.liveTranscript:
                            setRecordingMessage((prev) => ({
                                ...prev,
                                transcription: prev.transcription + message.payload,
                            }))
                            break
                        case eventTypes.liveTranslate:
                            setRecordingMessage((prev) => ({
                                ...prev,
                                translation: prev.translation + message.payload,
                            }))
                            break
                        case eventTypes.finalTranscript:
                            setRecordingMessage((prev) => ({
                                ...prev,
                                transcription: message.payload,
                            }))
                            break
                        case eventTypes.finalTranslate:
                            setRecordingMessage((prev) => ({
                                ...prev,
                                translation: message.payload,
                                status: 'processed',
                            }))
                            break
                        case eventTypes.error:
                            setRecordingMessage((prev) => ({
                                ...prev,
                                status: 'error',
                            }))
                            stopRecording()
                            break
                    }
                    console.log('Gladia message:', message)
                } catch (err) {
                    console.error('Error parsing message:', err)
                }
            }

            ws.onerror = (event) => {
                console.error('WebSocket error:', event)
            }

            ws.onclose = () => {
                console.log('WebSocket closed')
                wsRef.current = null
            }

            const constraints: MediaTrackConstraints = {
                channelCount: numChannels,
                echoCancellation: false,
                noiseSuppression: false,
                autoGainControl: false,
            }

            const stream = await navigator.mediaDevices.getUserMedia({
                audio: constraints,
            })

            const audioContext = new AudioContext({
                sampleRate,
                latencyHint: 'interactive',
            })

            await audioContext.audioWorklet.addModule('./gladia-processor.js')

            const source = audioContext.createMediaStreamSource(stream)
            const workletNode = new AudioWorkletNode(audioContext, 'gladia-processor')

            workletNode.port.onmessage = (event) => {
                if (!isRecordingRef.current || !wsRef.current) {
                    return
                }

                const input = event.data as Float32Array

                // Convert Float32 to PCM 16-bit
                const pcmBuffer = float32ToPCM16(input)

                // Send binary PCM data to WebSocket
                try {
                    ws.send(pcmBuffer)
                } catch (err) {
                    console.error('Error sending audio to WebSocket:', err)
                }
            }

            source.connect(workletNode)
            workletNode.connect(audioContext.destination)

            streamRef.current = stream
            sourceRef.current = source
            workletNodeRef.current = workletNode
            audioContextRef.current = audioContext

            isRecordingRef.current = true
            setRecordingSide(recordingSide)

            startedAtRef.current = performance.now()

            timerRef.current = window.setInterval(() => {
                setRecordingTime((performance.now() - startedAtRef.current) / 1000)
            }, 100)
        },
        [numChannels, sampleRate],
    )

    const stopRecording = useCallback(async () => {
        if (!isRecordingRef.current) {
            throw new Error('Not recording')
        }
        if (wsRef.current) {
            wsRef.current.send('stop_recording')
        }

        isRecordingRef.current = false
        setRecordingSide(null)

        await cleanup()
        wsRef.current = null

        setRecordingTime(0)
    }, [cleanup])

    return {
        startRecording,
        stopRecording,
        recordingSide,
        recordingTime,
    }
}

function float32ToPCM16(samples: Float32Array): ArrayBuffer {
    const buffer = new ArrayBuffer(samples.length * 2)
    const view = new DataView(buffer)

    for (let i = 0; i < samples.length; i++) {
        let sample = samples[i]

        // Clamp to [-1, 1]
        sample = sample < -1 ? -1 : sample > 1 ? 1 : sample

        // Convert to 16-bit PCM
        const int16 = sample < 0 ? sample * 0x8000 : sample * 0x7fff

        view.setInt16(i * 2, int16, true)
    }

    return buffer
}
