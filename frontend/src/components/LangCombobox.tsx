import {
    Combobox,
    ComboboxContent,
    ComboboxInput,
    ComboboxItem,
    ComboboxList,
    ComboboxClear,
} from '@/components/ui/combobox'
import { useRef } from 'react'

type LangItem = {
    label: string
    value: string
}

export default function LangCombobox(props: {
    languages: Record<string, string>
    lang: string
    setLang: (lang: string) => void
}) {
    const langItems = Object.keys(props.languages).map((key) => ({
        label: props.languages[key],
        value: key,
    }))

    const item: LangItem = {
        label: props.languages[props.lang],
        value: props.lang,
    }

    const clearRef = useRef<HTMLButtonElement>(null)

    return (
        <Combobox
            value={item}
            items={langItems}
            onValueChange={(val) => {
                if (!val) return
                props.setLang(val.value ?? item.value)
            }}
            itemToStringValue={(item: LangItem) => item.label}
        >
            {/* Hack for clean on focus to work. Getting this behaviour without it is pain in the ass. */}
            <ComboboxClear ref={clearRef} className="hidden" />
            <ComboboxInput
                onFocus={() => clearRef.current?.click()}
                className="*:text-base *:font-medium"
                placeholder="Select language"
            />
            <ComboboxContent side="top">
                <ComboboxList>
                    {(item: LangItem) => (
                        <ComboboxItem className="text-base" key={item.value} value={item}>
                            {item.label}
                        </ComboboxItem>
                    )}
                </ComboboxList>
            </ComboboxContent>
        </Combobox>
    )
}
