import { Button } from '@/components/ui/button'
import { ArrowLeftRight, Settings } from 'lucide-react'
import LangCombobox from './LangCombobox'
import { Link } from '@tanstack/react-router'
import type { Preferences } from '@/api/common'
import type { MutateFunction } from 'jotai-tanstack-query'

export default function Languages(props: {
    languages: Record<string, string>
    prefs: Preferences
    setPrefs: MutateFunction<Preferences, unknown, Preferences, void>
}) {
    function swagLanguages() {
        props.setPrefs({
            ...props.prefs,
            inLang: props.prefs.outLang,
            outLang: props.prefs.inLang
        })
    }

    return (
        <div className="relative border-y border-border/50 bg-card px-6 py-5 flex justify-center items-center gap-4">
            <div className="flex-1 max-w-64">
                <LangCombobox
                    languages={props.languages}
                    setLang={(newLang) => {
                        props.setPrefs({
                            ...props.prefs,
                            inLang: newLang,
                        })
                    }}
                    lang={props.prefs.inLang}
                />
            </div>
            <Button
                variant="ghost"
                size="icon"
                className="rounded-full shrink-0 hover:bg-primary/10"
                onClick={swagLanguages}
                title="Swap languages"
            >
                <ArrowLeftRight className="w-5 h-5 text-primary" />
            </Button>
            <div className="flex-1 max-w-64">
                <LangCombobox
                    languages={props.languages}
                    setLang={(newLang) => {
                        props.setPrefs({
                            ...props.prefs,
                            outLang: newLang,
                        })
                    }}
                    lang={props.prefs.outLang}
                />
            </div>
            <div className="">
                <Link to={'/settings'}>
                    <Button
                        variant="ghost"
                        size="icon"
                        className="rounded-full hover:bg-primary/10"
                    >
                        <Settings className="size-5 text-muted-foreground" />
                    </Button>
                </Link>
            </div>
        </div>
    )
}
