import { getCreditGrants, getMe } from '@/api/user'
import { atomWithQuery } from 'jotai-tanstack-query'

export const accountAtom = atomWithQuery(() => ({
    queryKey: ['account'],
    queryFn: async () => await Promise.all([getMe(), getCreditGrants()]),
}))
