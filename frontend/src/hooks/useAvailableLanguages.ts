import { getSupportedLanguages } from '@/api/common'
import { useEffect, useState } from 'react'

export default function useSupportedLanguages() {
    const [languages, setLanguages] = useState<Record<string, string>>({})

    async function fetchLanguages() {
        const langs = await getSupportedLanguages()
        setLanguages(langs)
    }

    useEffect(() => {
        fetchLanguages()
    }, [])

    return languages
}
