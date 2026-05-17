import { Button } from '@/components/ui/button'
import { ArrowLeftRight } from 'lucide-react'
import LangCombobox from './LangCombobox'

export default function Languages(props: {
    setOwnerLang: React.Dispatch<React.SetStateAction<string>>
    ownerLang: string
    setCompanionLang: React.Dispatch<React.SetStateAction<string>>
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
                <LangCombobox setLang={props.setOwnerLang} lang={props.ownerLang} />
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
                <LangCombobox setLang={props.setCompanionLang} lang={props.companionLang} />
            </div>
        </div>
    )
}
