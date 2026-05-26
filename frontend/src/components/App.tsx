import { ThemeProvider } from '@/components/theme-provider'
import Controls from '@/components/Controls'
import Languages from '@components/Languages'
import ChatMessage from '@/components/ui/chat-message'
import Chat from '@components/Chat'
import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { useGladiaRecorder } from '@/hooks/useGladiaRecorder'
import TextGenerateEffect from '@/components/ui/live-text-generate-effect'
import { useTwinChat } from '@/hooks/useTwinChat'
import { chatSide, messageStatuses } from '@/definitions/chat'
import type { ChatSide, Message } from '@/definitions/chat'
import ErrorState from '@/components/ui/error-state'
import { ping } from '@/api/common'
import { usePreferences } from '@/hooks/usePreferences'
import Ring from './ui/ring-spinner'
import useSupportedLanguages from '@/hooks/useAvailableLanguages'

function App() {
    const { messages, startRecordingMsg, errorProcessingMsg } = useTwinChat()

    const { start: startGladia, stop: stopGladia, recordingSide } = useGladiaRecorder()
    const languages = useSupportedLanguages()

    const { preferences } = usePreferences()

    const [ownLang, setOwnLang] = useState('')
    const [companionLang, setCompanionLang] = useState('')

    function setOwnLangCache(lang: string) {
        localStorage.setItem('ownLang', lang)
        setOwnLang(lang)
    }
    function setCompanionLangCache(lang: string) {
        localStorage.setItem('companionLang', lang)
        setCompanionLang(lang)
    }

    const ownChatRef = useRef<HTMLDivElement>(null)
    const companionChatRef = useRef<HTMLDivElement>(null)

    function afterInit() {
        setTimeout(() => {
            ownChatRef.current.scrollBy({
                top: ownChatRef.current.scrollHeight,
                behavior: 'instant',
            })
            companionChatRef.current.scrollBy({
                top: companionChatRef.current.scrollHeight,
                behavior: 'instant',
            })
        }, 10)
    }

    useEffect(() => {
        const cachedOwnLang = localStorage.getItem('ownLang') || 'en'
        setOwnLang(cachedOwnLang)
        const cachedCompanionLang = localStorage.getItem('companionLang') || 'fr'
        setCompanionLang(cachedCompanionLang)
        afterInit()
        ping()
    }, [])

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
    }, [messages])

    async function start(side: ChatSide) {
        try {
            startGladia(side, ownLang, companionLang, startRecordingMsg(side))
        } catch (err) {
            console.error('Error starting Gladia recording:', err)
            errorProcessingMsg(side)
        }
    }

    async function stop(side: ChatSide) {
        try {
            await stopGladia()
        } catch (err) {
            console.error('Error stopping Gladia recording:', err)
            errorProcessingMsg(side)
        }
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
                    return (
                        <TextGenerateEffect
                            words={'*Silence*'}
                            disable={true}
                        />
                    )
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
        <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
            <div className="relative font-sans font-normal h-dvh grid grid-rows-[1fr_auto_1fr] bg-background text-foreground">
                {/* Record button */}
                <div className="flex absolute top-0 z-10 rotate-180 right-1/2 translate-x-1/2 justify-center items-center p-4">
                    <div className="absolute backdrop-blur-sm bg-background/50 rounded-t-4xl inset-0 w-full h-full"></div>

                    <Controls
                        isRecording={recordingSide === chatSide.top}
                        disabled={recordingSide === chatSide.bottom}
                        start={start.bind(null, chatSide.top)}
                        stop={stop.bind(null, chatSide.top)}
                    />
                </div>
                {/* Companion section */}
                <div className="flex flex-col border-b border-border/50 min-h-0">
                    <Chat
                        ref={companionChatRef}
                        className="rotate-180 bg-background/50 pb-30 h-full"
                    >
                        {messages.map((msg) => (
                            <ChatMessage
                                key={msg.id}
                                type={msg.sendedFrom === chatSide.top ? 'outgoing' : 'incoming'}
                                size={preferences.chatFontSize}
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
                    languages={languages}
                    setOwnerLang={setOwnLangCache}
                    ownerLang={ownLang}
                    setCompanionLang={setCompanionLangCache}
                    companionLang={companionLang}
                />

                {/* Owner section */}
                <div className="flex relative flex-col min-h-0">
                    <Chat ref={ownChatRef} className="bg-background/50 pb-30 h-full">
                        {messages.map((msg) => (
                            <ChatMessage
                                key={msg.id}
                                type={msg.sendedFrom === chatSide.bottom ? 'outgoing' : 'incoming'}
                                size={preferences.chatFontSize}
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
                        start={start.bind(null, chatSide.bottom)}
                        stop={stop.bind(null, chatSide.bottom)}
                    />
                </div>
            </div>
        </ThemeProvider>
    )
}

export default App
