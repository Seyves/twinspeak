import Loader from '@/components/ui/loader'
import * as AuthApi from '@/api/auth'
import { createFileRoute, useSearch } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useAtom } from 'jotai'
import ErrorPage from '@/components/Error'
import { HTTPError } from 'ky'
import { atomWithMutation } from 'jotai-tanstack-query'

const googleCallbackAtom = atomWithMutation(() => ({
    mutationKey: ['google-callback'],
    mutationFn: AuthApi.googleProcessCallback,
}))

export const Route = createFileRoute('/auth/google-callback')({
    component: RouteComponent,
    validateSearch: (search): AuthApi.GoogleCallbackParams => {
        return {
            code: search.code ? String(search.code) : '',
            state: search.state ? String(search.state) : '',
        }
    },
})

function RouteComponent() {
    const params = useSearch({ from: '/auth/google-callback' })
    const [{ mutate, isSuccess, isError, error }] = useAtom(googleCallbackAtom)

    useEffect(() => {
        mutate(params)
    }, [])

    useEffect(() => {
        if (!isSuccess) return
        // Full page reload to remove Referrer header where google persists
        window.location.href = '/'
    }, [isSuccess])

    if (isError) {
        if (error instanceof HTTPError && error.response.status === 401) {
            return (
                <ErrorPage
                    message="Your session has either expired or is invalid."
                    onRetry={AuthApi.redirectToGoogleAuth}
                />
            )
        }
        return <ErrorPage onRetry={AuthApi.redirectToGoogleAuth} />
    }

    return <Loader />
}
