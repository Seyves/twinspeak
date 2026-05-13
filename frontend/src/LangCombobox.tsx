import {
    Combobox,
    ComboboxContent,
    ComboboxInput,
    ComboboxItem,
    ComboboxList,
} from '@/components/ui/combobox'
import { LangItem, langItems } from './lib/languages'

export default function LangCombobox(props: {
    lang: LangItem | null
    setLang: React.Dispatch<React.SetStateAction<LangItem | null>>
}) {
    return (
        <Combobox
            value={props.lang}
            items={langItems}
            onValueChange={props.setLang}
            itemToStringValue={(item: LangItem) => item.label}
        >
            <ComboboxInput className="mx-2 min-w-24 max-w-36 *:text-lg" placeholder="Select a framework" />
            <ComboboxContent>
                <ComboboxList>
                    {(item: LangItem) => (
                        <ComboboxItem className="text-lg" key={item.value} value={item}>
                            {item.label}
                        </ComboboxItem>
                    )}
                </ComboboxList>
            </ComboboxContent>
        </Combobox>
    )
}
