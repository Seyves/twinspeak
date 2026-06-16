import { createFileRoute } from '@tanstack/react-router'
import Controls from '@/components/Controls'
import Languages from '@/components/Languages'
import ChatMessage from '@/components/ui/chat-message'
import Chat from '@/components/Chat'
import { useEffect, useRef } from 'react'
import { useRecorder } from '@/hooks/useRecorder'
import TextGenerateEffect from '@/components/ui/live-text-generate-effect'
import { chatSide, messageStatuses } from '@/definitions/chat'
import type { ChatSide, Message } from '@/definitions/chat'
import ErrorState from '@/components/ui/error-state'
import ErrorPage from '@/components/Error'
import Ring from '@/components/ui/ring-spinner'
import { useAtom } from 'jotai'
import { preferencesAtom, updatePreferencesAtom } from '@/atoms/preferences'
import Loader from '@/components/ui/loader'
import { AnimatePresence } from 'motion/react'
import { messagesAtom } from '@/atoms/messages'
import { supportedLanguagesAtom } from '@/atoms/supported-languages'
import { queryClientAtom } from 'jotai-tanstack-query'

export const Route = createFileRoute('/')({
    component: Index,
})

const maxMessages = 10

function Index() {
    const [queryClient] = useAtom(queryClientAtom)

    const [msgs] = useAtom(messagesAtom)
    const [prefs] = useAtom(preferencesAtom)
    const [supportedLangs] = useAtom(supportedLanguagesAtom)

    const [{ mutate: setPrefs }] = useAtom(updatePreferencesAtom)

    const { startRecording, stopRecording, recordingSide } = useRecorder()

    const ownChatRef = useRef<HTMLDivElement>(null)
    const companionChatRef = useRef<HTMLDivElement>(null)

    async function refetch() {
        await Promise.all([msgs.refetch(), prefs.refetch(), supportedLangs.refetch()])
    }

    useEffect(() => {
        if (
            ownChatRef.current &&
            ownChatRef.current.scrollTop + 600 > ownChatRef.current.scrollHeight
        ) {
            ownChatRef.current.scrollBy({
                top: ownChatRef.current.scrollHeight,
                behavior: 'smooth',
            })
        }
        if (
            companionChatRef.current &&
            companionChatRef.current.scrollTop + 600 > companionChatRef.current.scrollHeight
        ) {
            companionChatRef.current.scrollBy({
                top: companionChatRef.current.scrollHeight,
                behavior: 'smooth',
            })
        }
    }, [msgs.data])

    async function startSpeech(side: ChatSide) {
        if (!prefs.isSuccess) return
        const id = crypto.randomUUID()

        // Adding new message with empty ring
        queryClient.setQueryData(['messages'], (prev: Message[]) => [
            ...limitMessages(prev),
            {
                id: id,
                sendedFrom: side,
                status: messageStatuses.recording,
                transcription: '',
                translation: '',
            },
        ])

        // Making setter for that new message
        const setMessage = (callback: (prev: Message) => Message) => {
            queryClient.setQueryData(['messages'], (prev: Message[]) =>
                prev.map((msg) => {
                    if (msg.id !== id) return msg
                    return callback(msg)
                }),
            )
        }

        startRecording(side, prefs.data.inLang, prefs.data.outLang, setMessage)
    }

    async function endSpeech(_: ChatSide) {
        await stopRecording()
    }

    function getMessageContent(side: ChatSide, msg: Message) {
        const isMineSide = msg.sendedFrom === side

        switch (msg.status) {
            case messageStatuses.recording:
                const text = isMineSide ? msg.transcription : msg.translation
                return (
                    <>
                        <TextGenerateEffect words={text} />
                        <Ring className="inline" />
                    </>
                )
            case messageStatuses.processed:
                if (msg.transcription == '') {
                    return <TextGenerateEffect words={'*Silence*'} disable={true} />
                }
                return (
                    <TextGenerateEffect
                        words={isMineSide ? msg.transcription : msg.translation}
                        disable={true}
                    />
                )
            case messageStatuses.error:
                return (
                    <ErrorState
                        className={isMineSide ? 'dark:text-foreground text-background' : ''}
                    >
                        <TextGenerateEffect words={"Couldn't process message"} />
                    </ErrorState>
                )
        }
    }

    return (
        <div className="relative h-full">
            <AnimatePresence>
                {(function () {
                    if (msgs.isPending || prefs.isPending || supportedLangs.isPending) {
                        return <Loader key="loader" />
                    }

                    if (msgs.isError || prefs.isError || supportedLangs.isError) {
                        return <ErrorPage key="error" onRetry={refetch} />
                    }

                    return (
                        <div
                            key="page"
                            className="relative font-sans font-normal h-dvh grid grid-rows-[1fr_auto_1fr] text-foreground"
                        >
                            {/* Record button */}
                            <div className="flex absolute top-0 z-10 rotate-180 right-1/2 translate-x-1/2 justify-center items-center p-4">
                                <div className="absolute backdrop-blur-sm bg-background/50 rounded-t-4xl inset-0 w-full h-full"></div>

                                <Controls
                                    isRecording={recordingSide === chatSide.top}
                                    disabled={recordingSide === chatSide.bottom}
                                    start={startSpeech.bind(null, chatSide.top)}
                                    stop={endSpeech.bind(null, chatSide.top)}
                                />
                            </div>
                            {/* Companion section */}
                            <div className="flex flex-col border-b border-border/50 min-h-0">
                                <Chat
                                    ref={companionChatRef}
                                    className="rotate-180 bg-background/50 pb-30 h-full"
                                >
                                    {[...msgs.data].reverse().map((msg) => (
                                        <ChatMessage
                                            key={msg.id}
                                            type={
                                                msg.sendedFrom === chatSide.top
                                                    ? 'outgoing'
                                                    : 'incoming'
                                            }
                                            size={prefs.data.chatMessageSize}
                                            className={
                                                msg.sendedFrom === chatSide.top &&
                                                msg.status === messageStatuses.error
                                                    ? 'bg-destructive'
                                                    : ''
                                            }
                                        >
                                            {getMessageContent(chatSide.top, msg)}
                                        </ChatMessage>
                                    ))}
                                </Chat>
                            </div>

                            {/* Center control section */}
                            <Languages
                                languages={supportedLangs.data}
                                setOwnerLang={(lang) =>
                                    setPrefs({
                                        ...prefs.data,
                                        inLang: lang,
                                    })
                                }
                                ownerLang={prefs.data.inLang}
                                setCompanionLang={(lang) =>
                                    setPrefs({
                                        ...prefs.data,
                                        outLang: lang,
                                    })
                                }
                                companionLang={prefs.data.outLang}
                            />

                            {/* Owner section */}
                            <div className="flex relative flex-col min-h-0">
                                <Chat ref={ownChatRef} className="bg-background/50 pb-30 h-full">
                                    {[...msgs.data].reverse().map((msg) => (
                                        <ChatMessage
                                            key={msg.id}
                                            type={
                                                msg.sendedFrom === chatSide.bottom
                                                    ? 'outgoing'
                                                    : 'incoming'
                                            }
                                            size={prefs.data.chatMessageSize}
                                            className={
                                                msg.sendedFrom === chatSide.bottom &&
                                                msg.status === messageStatuses.error
                                                    ? 'bg-destructive'
                                                    : ''
                                            }
                                        >
                                            {getMessageContent(chatSide.bottom, msg)}
                                        </ChatMessage>
                                    ))}
                                </Chat>
                            </div>

                            {/* Record button */}
                            <div className="flex absolute bottom-0 right-1/2 translate-x-1/2 justify-center items-center p-4">
                                <div className="absolute backdrop-blur-sm bg-background/50 rounded-t-4xl inset-0 w-full h-full"></div>

                                <Controls
                                    isRecording={recordingSide === chatSide.bottom}
                                    disabled={recordingSide === chatSide.top}
                                    start={startSpeech.bind(null, chatSide.bottom)}
                                    stop={endSpeech.bind(null, chatSide.bottom)}
                                />
                            </div>
                        </div>
                    )
                })()}
            </AnimatePresence>
        </div>
    )
}

function limitMessages(messages: Message[]) {
    const limit = maxMessages - 1
    if (messages.length > limit) {
        return messages.slice(messages.length - limit)
    }
    return messages
}
