import {
    Combobox,
    ComboboxContent,
    ComboboxInput,
    ComboboxItem,
    ComboboxList,
} from '@/components/ui/combobox'
import { langItems, languages } from '@/lib/languages'
import type { LangItem } from '@/lib/languages'

export default function LangCombobox(props: {
    lang: string
    setLang: React.Dispatch<React.SetStateAction<string>>
}) {
    const item: LangItem = {
        label: languages[props.lang],
        value: props.lang,
    }

    return (
        <Combobox
            value={item}
            items={langItems}
            onValueChange={(val) => props.setLang(val?.value!)}
            itemToStringValue={(item: LangItem) => item.label}
        >
            <ComboboxInput className="*:text-base *:font-medium" placeholder="Select language" />
            <ComboboxContent>
                <ComboboxList>
                    {(item: LangItem) => (
                        <ComboboxItem key={item.value} value={item}>
                            {item.label}
                        </ComboboxItem>
                    )}
                </ComboboxList>
            </ComboboxContent>
        </Combobox>
    )
}
