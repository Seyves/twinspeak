import { getSupportedLanguages } from '@/api/common'
import { atomWithQuery } from 'jotai-tanstack-query'

export const supportedLanguagesAtom = atomWithQuery(() => ({
    queryKey: ['supported-languages'],
    queryFn: getSupportedLanguages,
    staleTime: Infinity,
}))
