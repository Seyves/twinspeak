import { createFileRoute } from '@tanstack/react-router'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { type Preferences } from '@/api/common'
import { useAtom } from 'jotai'
import { chatMessageSize, theme, type ChatMessageSize, type Theme } from '@/definitions/chat'
import { Mic } from 'lucide-react'
import { useEffect, useState } from 'react'
import Loader from '@/components/ui/loader'
import ErrorPage from '@/components/Error'
import { AnimatePresence, motion } from 'motion/react'
import { inputDeviceAtom, preferencesAtom, updatePreferencesAtom } from '@/atoms/preferences'
import SectionCard from '@/components/SectionCard'

export const Route = createFileRoute('/settings/preferences')({
    component: Preferences,
})

function Preferences() {
    const [inputDevice, setInputDevice] = useAtom(inputDeviceAtom)
    const [audioDevices, setAudioDevices] = useState<MediaDeviceInfo[]>([])
    const [{ data: prefs, isPending, isError, refetch }] = useAtom(preferencesAtom)
    const [{ mutate: setPrefs }] = useAtom(updatePreferencesAtom)

    useEffect(() => {
        try {
            getMicDevices()
        } catch (e) {
            console.error(e)
        }
    }, [])

    async function getMicDevices() {
        await navigator.mediaDevices.getUserMedia({ audio: true })
        const devices = await navigator.mediaDevices.enumerateDevices()
        const micDevices = devices.filter((d) => d.kind === 'audioinput')
        setAudioDevices(micDevices)
    }

    return (
        <div className="relative h-full">
            <AnimatePresence>
                {(function () {
                    if (isPending) return <Loader key="loader" />

                    if (isError) return <ErrorPage key="error" onRetry={refetch} />

                    return (
                        <motion.div key="content">
                            <SectionCard label="Appearance">
                                <div className="space-y-5">
                                    <div className="flex items-center justify-between gap-4">
                                        <label className="text font-medium">Message size</label>
                                        <Select
                                            value={prefs.chatMessageSize}
                                            onValueChange={(v) =>
                                                setPrefs({
                                                    ...prefs,
                                                    chatMessageSize: v as ChatMessageSize,
                                                })
                                            }
                                        >
                                            <SelectTrigger className="w-44 text-base">
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem
                                                    className="text-base"
                                                    value={chatMessageSize.sm}
                                                >
                                                    Small
                                                </SelectItem>
                                                <SelectItem
                                                    className="text-base"
                                                    value={chatMessageSize.md}
                                                >
                                                    Medium
                                                </SelectItem>
                                                <SelectItem
                                                    className="text-base"
                                                    value={chatMessageSize.lg}
                                                >
                                                    Large
                                                </SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>

                                    <div className="flex items-center justify-between gap-4">
                                        <label className="font-medium">Theme</label>
                                        <Select
                                            value={prefs.theme}
                                            onValueChange={(v) =>
                                                setPrefs({
                                                    ...prefs,
                                                    theme: v as Theme,
                                                })
                                            }
                                        >
                                            <SelectTrigger className="w-44 text-base">
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem
                                                    className="text-base"
                                                    value={theme.dark}
                                                >
                                                    Dark
                                                </SelectItem>
                                                <SelectItem
                                                    className="text-base"
                                                    value={theme.light}
                                                >
                                                    Light
                                                </SelectItem>
                                                <SelectItem
                                                    className="text-base"
                                                    value={theme.system}
                                                >
                                                    System
                                                </SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>
                                </div>
                            </SectionCard>

                            <SectionCard label="Audio">
                                <div className="flex items-center justify-between gap-4">
                                    <div className="flex items-center gap-2">
                                        <Mic className="w-4 h-4 text-muted-foreground" />
                                        <label className="text font-medium">Microphone</label>
                                    </div>
                                    <Select value={inputDevice} onValueChange={setInputDevice}>
                                        <SelectTrigger className="w-44">
                                            <SelectValue placeholder="Default" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="default">Default</SelectItem>
                                            {audioDevices
                                                .filter((d) => d.deviceId !== '')
                                                .map((device, idx) => (
                                                    <SelectItem
                                                        key={device.deviceId}
                                                        value={device.deviceId}
                                                    >
                                                        {device.label || `Microphone ${idx + 1}`}
                                                    </SelectItem>
                                                ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                            </SectionCard>
                        </motion.div>
                    )
                })()}
            </AnimatePresence>
        </div>
    )
}
