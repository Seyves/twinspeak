import Loader from '@/components/ui/loader'
import { redirectToGoogleAuth, type GoogleCallbackParams } from '@/api/auth'
import { createFileRoute, useNavigate, useSearch } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useAtom } from 'jotai'
import { googleCallbackAtom } from '@/atoms/auth'
import ErrorPage from '@/components/Error'
import { HTTPError } from 'ky'

export const Route = createFileRoute('/auth/google-callback')({
    component: RouteComponent,
    validateSearch: (search): GoogleCallbackParams => {
        return {
            code: search.code ? String(search.code) : '',
            state: search.state ? String(search.state) : '',
        }
    },
})

function RouteComponent() {
    const navigate = useNavigate()
    const params = useSearch({ from: '/auth/google-callback' })
    const [{ mutate, isSuccess, isError, error }] = useAtom(googleCallbackAtom)

    useEffect(() => {
        mutate(params)
    }, [])

    useEffect(() => {
        if (!isSuccess) return
        navigate({ to: '/' })
    }, [isSuccess])

    if (isError) {
        if (error instanceof HTTPError && error.response.status === 401) {
            return (
                <ErrorPage
                    message="Your session has either expired or is invalid."
                    onRetry={redirectToGoogleAuth}
                />
            )
        }
        return <ErrorPage onRetry={redirectToGoogleAuth} />
    }

    return <Loader />
}
