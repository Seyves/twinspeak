import { ThemeProvider } from '@/components/theme-provider'
import Controls from './Controls'
import Languages from './Languages'
import ChatMessage from './components/ui/chat-message'
import Chat from './Chat'
import { useState } from 'react'
import { LangItem, langItems } from './lib/languages'
import LiveTextGenerateEffect from './components/ui/live-text-generate-effect'
import { useRecordSpeech } from './hooks/useRecordSpeech'
import { processSpeech } from './api'

function App() {
    const { start, stop, isRecording } = useRecordSpeech()

    const [ownerLang, setOwnerLang] = useState<LangItem | null>(
        langItems.find((item) => item.value == 'en') as LangItem,
    )
    const [companionLang, setCompanionLang] = useState<LangItem | null>(
        langItems.find((item) => item.value == 'zh') as LangItem,
    )

    const [ownerMessages, setOwnerMessages] = useState<string[]>([]) 
    const [companionMessages, setCompanionMessages] = useState<string[]>([]) 

    async function stopRecording() {
        const blob = await stop()
        const res = await processSpeech(blob, ownerLang?.value, companionLang?.value)
        setOwnerMessages((prev) => [...prev, res.transcription])
        setCompanionMessages((prev) => [...prev, res.translation])
    }

    return (
        <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
            <div className="font-sans overflow-hidden h-dvh font-normal grid grid-rows-[100px_1fr_auto_1fr_100px]">
                <Controls className="rotate-180" />
                <Chat className="rotate-180">
                    {companionMessages.map((msg) => <ChatMessage type='outgoing'>{msg}</ChatMessage>)}
                </Chat>
                <Languages
                    setOwnerLang={setOwnerLang}
                    ownerLang={ownerLang}
                    setCompanionLang={setCompanionLang}
                    companionLang={companionLang}
                />
                <Chat>
                    {ownerMessages.map((msg) => <ChatMessage type='outgoing'>{msg}</ChatMessage>)}
                </Chat>
                <Controls isRecording={isRecording} start={start} stop={stopRecording} />
            </div>
        </ThemeProvider>
    )
}

export default App
