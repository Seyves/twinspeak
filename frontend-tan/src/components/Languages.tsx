import { Button } from '@/components/ui/button'
import { ArrowLeftRight, Settings } from 'lucide-react'
import LangCombobox from './LangCombobox'
import { Link } from '@tanstack/react-router'

export default function Languages(props: {
    languages: Record<string, string>
    setOwnerLang: (lang: string) => void
    ownerLang: string
    setCompanionLang: (lang: string) => void
    companionLang: string
}) {
    function swagLanguages() {
        const ownerLang = props.ownerLang
        props.setOwnerLang(props.companionLang)
        props.setCompanionLang(ownerLang)
    }

    return (
        <div className="relative border-y border-border/50 bg-linear-to-r from-card/30 via-card/20 to-card/30 px-6 py-5 flex justify-center items-center gap-4">
            <div className="flex-1 max-w-64">
                <LangCombobox
                    languages={props.languages}
                    setLang={props.setOwnerLang}
                    lang={props.ownerLang}
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
                    setLang={props.setCompanionLang}
                    lang={props.companionLang}
                />
            </div>
            <div className="">
                <Link to={"/settings"}>
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
