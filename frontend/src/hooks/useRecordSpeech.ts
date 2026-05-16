import { Direction } from '@/definitions/messages'
import { useCallback, useEffect, useRef, useState } from 'react'

type UseWavRecorderOptions = {
    sampleRate?: number
    numChannels?: number
}

export function useRecordSpeech({
    sampleRate = 16000,
    numChannels = 1,
}: UseWavRecorderOptions = {}) {
    const [recordingDirection, setRecordingDirection] = useState<Direction | null>(null)
    const [recordingTime, setRecordingTime] = useState(0)

    const audioContextRef = useRef<AudioContext | null>(null)
    const streamRef = useRef<MediaStream | null>(null)
    const sourceRef = useRef<MediaStreamAudioSourceNode | null>(null)
    const workletNodeRef = useRef<AudioWorkletNode | null>(null)

    const chunksRef = useRef<Float32Array[]>([])
    const totalLengthRef = useRef(0)

    const timerRef = useRef<number | null>(null)
    const startedAtRef = useRef<number>(0)

    const isRecordingRef = useRef(false)

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
            console.error(err)
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

    const start = useCallback(
        async (direction: Direction) => {
            if (isRecordingRef.current) {
                return
            }

            chunksRef.current = []
            totalLengthRef.current = 0

            const stream = await navigator.mediaDevices.getUserMedia({
                audio: {
                    channelCount: numChannels,
                    echoCancellation: false,
                    noiseSuppression: false,
                    autoGainControl: false,
                },
            })

            const audioContext = new AudioContext({
                sampleRate,
                latencyHint: 'interactive',
            })

            await audioContext.audioWorklet.addModule('./recorder-processor.js')

            const source = audioContext.createMediaStreamSource(stream)

            const workletNode = new AudioWorkletNode(audioContext, 'wav-recorder-processor')

            workletNode.port.onmessage = (event) => {
                if (!isRecordingRef.current) {
                    return
                }

                const input = event.data as Float32Array

                const copy = new Float32Array(input.length)
                copy.set(input)

                chunksRef.current.push(copy)
                totalLengthRef.current += copy.length
            }

            source.connect(workletNode)

            // Keeps processor alive across browsers
            workletNode.connect(audioContext.destination)

            streamRef.current = stream
            sourceRef.current = source
            workletNodeRef.current = workletNode
            audioContextRef.current = audioContext

            isRecordingRef.current = true

            startedAtRef.current = performance.now()

            timerRef.current = window.setInterval(() => {
                setRecordingTime((performance.now() - startedAtRef.current) / 1000)
            }, 100)

            setRecordingDirection(direction)
        },
        [numChannels, sampleRate],
    )

    const stop = useCallback(async (): Promise<Blob> => {
        if (!isRecordingRef.current) {
            throw new Error('Not recording')
        }

        isRecordingRef.current = false

        setRecordingDirection(null)

        await cleanup()

        const mergedBuffer = new Float32Array(totalLengthRef.current)

        let offset = 0

        for (const chunk of chunksRef.current) {
            mergedBuffer.set(chunk, offset)
            offset += chunk.length
        }

        const wavBlob = encodeWav({
            samples: mergedBuffer,
            sampleRate,
            numChannels,
        })

        chunksRef.current = []
        totalLengthRef.current = 0
        setRecordingTime(0)

        return wavBlob
    }, [cleanup, numChannels, sampleRate])

    const download = useCallback((blob: Blob, filename = 'recording.wav') => {
        const url = URL.createObjectURL(blob)

        const a = document.createElement('a')

        a.href = url
        a.download = filename
        a.style.display = 'none'

        document.body.appendChild(a)

        a.click()

        setTimeout(() => {
            URL.revokeObjectURL(url)
            document.body.removeChild(a)
        }, 100)
    }, [])

    return {
        start,
        stop,
        download,
        recordingTime,
        recordingDirection,
    }
}

function encodeWav({
    samples,
    sampleRate,
    numChannels,
}: {
    samples: Float32Array
    sampleRate: number
    numChannels: number
}) {
    const bytesPerSample = 2
    const blockAlign = numChannels * bytesPerSample
    const dataSize = samples.length * bytesPerSample

    const buffer = new ArrayBuffer(44 + dataSize)

    const view = new DataView(buffer)

    writeString(view, 0, 'RIFF')
    view.setUint32(4, 36 + dataSize, true)
    writeString(view, 8, 'WAVE')

    writeString(view, 12, 'fmt ')
    view.setUint32(16, 16, true)
    view.setUint16(20, 1, true)
    view.setUint16(22, numChannels, true)
    view.setUint32(24, sampleRate, true)
    view.setUint32(28, sampleRate * blockAlign, true)
    view.setUint16(32, blockAlign, true)
    view.setUint16(34, 16, true)

    writeString(view, 36, 'data')
    view.setUint32(40, dataSize, true)

    floatTo16BitPCM(view, 44, samples)

    return new Blob([buffer], {
        type: 'audio/wav',
    })
}

function floatTo16BitPCM(view: DataView, offset: number, input: Float32Array) {
    for (let i = 0; i < input.length; i++, offset += 2) {
        let sample = input[i]

        sample = sample < -1 ? -1 : sample > 1 ? 1 : sample

        const int16 = sample < 0 ? sample * 0x8000 : sample * 0x7fff

        view.setInt16(offset, int16, true)
    }
}

function writeString(view: DataView, offset: number, text: string) {
    for (let i = 0; i < text.length; i++) {
        view.setUint8(offset + i, text.charCodeAt(i))
    }
}
