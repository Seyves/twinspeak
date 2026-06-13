import { googleProcessCallback } from '@/api/auth'
import { atomWithMutation } from 'jotai-tanstack-query'

export const googleCallbackAtom = atomWithMutation(() => ({
    mutationKey: ['google-callback'],
    mutationFn: googleProcessCallback,
}))
