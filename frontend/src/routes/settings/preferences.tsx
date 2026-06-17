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
import { chatMessageSize, themes, type ChatMessageSize, type Theme } from '@/definitions/chat'
import { useEffect } from 'react'
import Loader from '@/components/ui/loader'
import ErrorPage from '@/components/Error'
import { AnimatePresence, motion } from 'motion/react'
import SectionCard from '@/components/SectionCard'
import { localThemeAtom } from '@/components/theme-provider'
import { atomWithQuery, atomWithMutation, queryClientAtom } from 'jotai-tanstack-query'
import { getPreferences, updatePreferences } from '@/api/common'

export const Route = createFileRoute('/settings/preferences')({
    component: Preferences,
})

export const preferencesAtom = atomWithQuery(() => ({
    queryKey: ['preferences'],
    queryFn: getPreferences,
}))

export const updatePreferencesAtom = atomWithMutation((get) => ({
    mutationKey: ['update-preferences'],
    mutationFn: async (prefs: Preferences) => {
        await updatePreferences(prefs)
        return prefs
    },
    onMutate: (data) => {
        const queryClient = get(queryClientAtom)
        queryClient.setQueryData(['preferences'], data)
    },
}))

function Preferences() {
    const [_, setLocalTheme] = useAtom(localThemeAtom)
    const [{ data: prefs, isSuccess, isPending, isError, refetch }] = useAtom(preferencesAtom)
    const [{ mutate: setPrefs }] = useAtom(updatePreferencesAtom)

    useEffect(() => {
        if (isSuccess) setLocalTheme(prefs.theme)
    }, [isSuccess])

    useEffect(() => {
        try {
            // getMicDevices()
        } catch (e) {
            console.error(e)
        }
    }, [])

    // async function getMicDevices() {
    //     await navigator.mediaDevices.getUserMedia({ audio: true })
    //     const devices = await navigator.mediaDevices.enumerateDevices()
    //     const micDevices = devices.filter((d) => d.kind === 'audioinput')
    //     setAudioDevices(micDevices)
    // }

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
                                            onValueChange={(v) => {
                                                const theme = v as Theme
                                                setLocalTheme(theme)
                                                setPrefs({ ...prefs, theme })
                                            }}
                                        >
                                            <SelectTrigger className="w-44 text-base">
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem
                                                    className="text-base"
                                                    value={themes.dark}
                                                >
                                                    Dark
                                                </SelectItem>
                                                <SelectItem
                                                    className="text-base"
                                                    value={themes.light}
                                                >
                                                    Light
                                                </SelectItem>
                                                <SelectItem
                                                    className="text-base"
                                                    value={themes.system}
                                                >
                                                    System
                                                </SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>
                                </div>
                            </SectionCard>
                        </motion.div>
                    )
                })()}
            </AnimatePresence>
        </div>
    )
}
