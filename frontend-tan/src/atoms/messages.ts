import { getMessages } from '@/api/common'
import { atomWithQuery } from 'jotai-tanstack-query'

export const messagesAtom = atomWithQuery(() => ({
    queryKey: ['messages'],
    queryFn: getMessages,
    staleTime: Infinity,
}))
