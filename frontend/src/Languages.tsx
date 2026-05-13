import { Button } from './components/ui/button'
import { ArrowLeftRight } from 'lucide-react'
import LangCombobox from './LangCombobox'
import { LangItem } from './lib/languages'

export default function Languages(props: {
    setOwnerLang: React.Dispatch<React.SetStateAction<LangItem | null>>
    ownerLang: LangItem | null
    setCompanionLang: React.Dispatch<React.SetStateAction<LangItem | null>>
    companionLang: LangItem | null
}) {
    function swagLanguages() {
        const ownerLang = props.ownerLang
        props.setOwnerLang(props.companionLang)
        props.setCompanionLang(ownerLang)
    }

    return (
        <div className="border-(--color-border) border bg-card px-4 py-2 flex justify-between items-center">
            <LangCombobox setLang={props.setOwnerLang} lang={props.ownerLang} />
            <Button variant="outline" size="icon" className="rounded-full" onClick={swagLanguages}>
                <ArrowLeftRight />
            </Button>
            <LangCombobox setLang={props.setCompanionLang} lang={props.companionLang} />
        </div>
    )
}
