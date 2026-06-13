import { atomWithStorage } from 'jotai/utils'
import { atomWithQuery, atomWithMutation, queryClientAtom } from 'jotai-tanstack-query'
import { getPreferences, updatePreferences, type Preferences } from '@/api/common'

// Storing device preference only client side
export const inputDeviceAtom = atomWithStorage('inputDevice', 'default')

// Other should be stored server side
export const preferencesAtom = atomWithQuery(() => ({
    queryKey: ['preferences'],
    queryFn: getPreferences,
}))

export const updatePreferencesAtom = atomWithMutation((get) => ({
    mutationKey: ['preferences'],
    mutationFn: async (prefs: Preferences) => {
        await updatePreferences(prefs)
        return prefs
    },
    onMutate: (data) => {
        const queryClient = get(queryClientAtom)
        queryClient.setQueryData(['preferences'], data)
    },
}))
