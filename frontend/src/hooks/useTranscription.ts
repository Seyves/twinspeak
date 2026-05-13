import { useState, useRef, useCallback } from 'react'
import { TranscriptionStatus, WhisperMessage, TranscriptLine } from '@/types/transcription'

export function useTranscription() {
    const [status, setStatus] = useState<TranscriptionStatus>('idle')
    const [lines, setLines] = useState<TranscriptLine[]>([])
    const [error, setError] = useState<string | null>(null)

    const wsRef = useRef<WebSocket | null>(null)
    const audioContextRef = useRef<AudioContext | null>(null)
    const mediaStreamRef = useRef<MediaStream | null>(null)
    const workletNodeRef = useRef<AudioWorkletNode | null>(null)
    const sourceRef = useRef<AudioNode | null>(null)
    const isServerReadyRef = useRef(false)

    const start = useCallback(async () => {
        try {
            setStatus('connecting')
            setError(null)
            setLines([])

            // Open WebSocket connection
            const wsURL = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws/transcribe`
            const ws = new WebSocket(wsURL)
            wsRef.current = ws

            ws.onopen = async () => {
                try {
                    // Send initial config to WhisperLive
                    const config = {
                        uid: crypto.randomUUID(),
                        task: 'transcribe',
                        model: 'base',
                        use_vad: false,
                        language: 'ru',
                        send_last_n_segments: 10,
                        no_speech_thresh: 0.45,
                        clip_audio: false,
                        same_output_threshold: 10,
                        enable_translation: false,
                    }
                    ws.send(JSON.stringify(config))

                    // Request microphone access
                    const mediaStream = await navigator.mediaDevices.getUserMedia({
                        audio: {
                            echoCancellation: true,
                            noiseSuppression: true,
                        },
                    })
                    mediaStreamRef.current = mediaStream

                    // Create AudioContext
                    const audioContext = new (
                        window.AudioContext || (window as any).webkitAudioContext
                    )()
                    audioContextRef.current = audioContext

                    // Add AudioWorklet module
                    await audioContext.audioWorklet.addModule('/audiopreprocessor.js')

                    // Create AudioWorkletNode
                    const workletNode = new AudioWorkletNode(audioContext, 'audiopreprocessor')
                    const stream = audioContext.createMediaStreamSource(mediaStream)
                    stream.connect(workletNode)

                    workletNodeRef.current = workletNode

                    // Set up message handler for audio chunks
                    workletNode.port.onmessage = (event) => {
                        const data = event.data
                        const audio16k = data // Float32Array @ 16 kHz
                        if (isServerReadyRef.current && ws.readyState === WebSocket.OPEN) {
                            ws.send(audio16k)
                        }
                    }

                    // Create source and connect
                    sourceRef.current = stream

                    setStatus('recording')
                } catch (err) {
                    const message =
                        err instanceof Error ? err.message : 'Failed to initialize audio'
                    setError(message)
                    setStatus('error')
                    ws.close()
                }
            }

            ws.onmessage = (event) => {
                try {
                    let message: WhisperMessage

                    // Handle both string and ArrayBuffer data
                    if (typeof event.data === 'string') {
                        message = JSON.parse(event.data)
                    } else if (event.data instanceof ArrayBuffer) {
                        // Convert ArrayBuffer to string then parse
                        const decoder = new TextDecoder()
                        const text = decoder.decode(event.data)
                        message = JSON.parse(text)
                    } else {
                        return
                    }

                    console.log('Server response:', message)

                    if (message.message === 'SERVER_READY') {
                        isServerReadyRef.current = true
                        // Server is ready for audio
                    } else if (message.error) {
                        setError(message.error)
                        setStatus('error')
                    } else if (message.segments) {
                        const newLines: TranscriptLine[] = message.segments.map((seg, idx) => ({
                            id: seg.id ?? idx,
                            text: seg.text,
                            completed: seg.completed,
                        }))
                        setLines(newLines)
                    }
                } catch (err) {
                    console.error('Failed to parse message:', err)
                }
            }

            ws.onerror = () => {
                setError('WebSocket error occurred')
                setStatus('error')
            }

            ws.onclose = () => {
                setStatus('idle')
            }
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to start transcription'
            setError(message)
            setStatus('error')
        }
    }, [])

    const stop = useCallback(async () => {
        try {
            setStatus('stopping')

            // Send END_OF_AUDIO signal
            if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
                wsRef.current.send('END_OF_AUDIO')
            }

            // Stop audio stream
            if (mediaStreamRef.current) {
                mediaStreamRef.current.getTracks().forEach((track) => track.stop())
            }

            // Disconnect audio nodes
            if (sourceRef.current) {
                sourceRef.current.disconnect()
            }
            if (workletNodeRef.current) {
                workletNodeRef.current.disconnect()
            }

            // Close AudioContext
            if (audioContextRef.current && audioContextRef.current.state !== 'closed') {
                await audioContextRef.current.close()
            }

            // Close WebSocket
            if (wsRef.current) {
                wsRef.current.close()
            }


            setStatus('idle')
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Error stopping transcription'
            setError(message)
            setStatus('error')
        }
    }, [])

    return {
        status,
        lines,
        error,
        start,
        stop,
        isRecording: status === 'recording',
    }
}
