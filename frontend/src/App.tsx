import { ThemeProvider } from '@/components/theme-provider'
import Controls from './Controls'
import Languages from './Languages'
import ChatMessage from './components/ui/chat-message'
import Chat from './Chat'
import { useRef, useState } from 'react'
import { useRecordSpeech } from './hooks/useRecordSpeech'
import { processSpeech } from './api'
import Ring from './components/ui/ring-spinner'
import TextGenerateEffect from './components/ui/live-text-generate-effect'
import { useTwinChat } from './hooks/useTwinChat'
import { directions, Direction, Message, messageStatuses } from './definitions/messages'
import ErrorState from './components/ui/error-state'
import { Auth } from './components/Auth'

function App() {
    // TODO: Replace with actual auth state from context or API
    const [isAuthenticated, setIsAuthenticated] = useState(false)

    const {
        messages,
        startRecordingMsg,
        startProcessingMsg,
        successProcessingMsg,
        errorProcessingMsg,
    } = useTwinChat()

    const { start, stop, recordingDirection } = useRecordSpeech()
    const isRecording = recordingDirection !== null

    const [ownLang, setOwnLang] = useState('en')
    const [companionLang, setCompanionLang] = useState('zh')

    const ownChatRef = useRef<HTMLDivElement>(null)
    const companionChatRef = useRef<HTMLDivElement>(null)

    function startRecording(direction: Direction) {
        start(direction)
        startRecordingMsg(direction)
    }

    async function stopRecording(direction: Direction) {
        startProcessingMsg(direction)
        const blob = await stop()
        try {
            let res
            if (direction === directions.own) {
                res = await processSpeech(blob, ownLang, companionLang)
            } else {
                res = await processSpeech(blob, companionLang, ownLang)
            }
            successProcessingMsg(direction, res.transcription, res.translation)
        } catch (e) {
            errorProcessingMsg(direction)
        }
        setTimeout(() => {
            ownChatRef.current?.scrollBy({
                top: ownChatRef.current.scrollHeight,
                behavior: 'smooth',
            })
            companionChatRef.current?.scrollBy({
                top: companionChatRef.current.scrollHeight,
                behavior: 'smooth',
            })
        }, 300)
    }

    function getMessageContent(orientation: Direction, msg: Message) {
        switch (msg.status) {
            case messageStatuses.recording:
            case messageStatuses.pending:
                return <Ring />
            case messageStatuses.processed:
                if (orientation === msg.direction) {
                    return <TextGenerateEffect words={msg.transcription} />
                } else {
                    return <TextGenerateEffect words={msg.translation} />
                }
            case messageStatuses.error:
                return (
                    <ErrorState className={orientation === msg.direction ? 'text-foreground' : ''}>
                        <TextGenerateEffect words={'Could not process message'} />
                    </ErrorState>
                )
        }
    }

    if (!isAuthenticated) {
        return (
            <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
                <Auth />
            </ThemeProvider>
        )
    }

    return (
        <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
            <div className="relative font-sans overflow-hidden h-dvh font-normal grid grid-rows-[1fr_auto_1fr] bg-background text-foreground">
                {/* Record button */}
                <div className="flex absolute top-0 z-10 rotate-180 right-1/2 translate-x-1/2 justify-center items-center p-4">
                    <div className="absolute backdrop-blur-sm bg-background/50 rounded-t-4xl inset-0 w-full h-full"></div>

                    <Controls
                        isRecording={recordingDirection === directions.companion}
                        disabled={recordingDirection === directions.own}
                        start={startRecording.bind(null, directions.companion)}
                        stop={stopRecording.bind(null, directions.companion)}
                    />
                </div>
                {/* Companion section */}
                <div className="flex flex-col border-b border-border/50 min-h-0">
                    <Chat ref={companionChatRef} className="rotate-180 bg-background/50 pb-30 h-full">
                        {messages.map((msg) => (
                            <ChatMessage
                                key={msg.id}
                                type={msg.direction === directions.own ? 'incoming' : 'outgoing'}
                                className={
                                    msg.direction === directions.companion &&
                                    msg.status === messageStatuses.error
                                        ? 'bg-destructive'
                                        : ''
                                }
                            >
                                {getMessageContent(directions.companion, msg)}
                            </ChatMessage>
                        ))}
                    </Chat>
                </div>

                {/* Center control section */}
                <Languages
                    setOwnerLang={setOwnLang}
                    ownerLang={ownLang}
                    setCompanionLang={setCompanionLang}
                    companionLang={companionLang}
                />

                {/* Owner section */}
                <div className="flex relative flex-col min-h-0">
                    <Chat ref={ownChatRef} className="bg-background/50 pb-30 h-full">
                        {messages.map((msg) => (
                            <ChatMessage
                                key={msg.id}
                                type={
                                    msg.direction === directions.companion ? 'incoming' : 'outgoing'
                                }
                                className={
                                    msg.direction === directions.own &&
                                    msg.status === messageStatuses.error
                                        ? 'bg-destructive'
                                        : ''
                                }
                            >
                                {getMessageContent(directions.own, msg)}
                            </ChatMessage>
                        ))}
                    </Chat>
                </div>

                {/* Record button */}
                <div className="flex absolute bottom-0 right-1/2 translate-x-1/2 justify-center items-center p-4">
                    <div className="absolute backdrop-blur-sm bg-background/50 rounded-t-4xl inset-0 w-full h-full"></div>

                    <Controls
                        isRecording={recordingDirection === directions.own}
                        disabled={recordingDirection === directions.companion}
                        start={startRecording.bind(null, directions.own)}
                        stop={stopRecording.bind(null, directions.own)}
                    />
                </div>
            </div>
        </ThemeProvider>
    )
}

export default App
