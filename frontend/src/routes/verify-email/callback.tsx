import { createFileRoute } from '@tanstack/react-router'
import { useEffect } from 'react'
import * as VerificationApi from '@/api/verification'
import { AnimatedBackground } from '@/components/animated-background'
import Loader from '@/components/ui/loader'
import { Button } from '@/components/ui/button'
import { CircleX, MailCheck } from 'lucide-react'
import { atomWithMutation } from 'jotai-tanstack-query'
import { useAtom } from 'jotai'

export const Route = createFileRoute('/verify-email/callback')({
    component: VerifyEmailCallback,
    validateSearch: (search: Record<string, unknown>) => {
        return {
            token: (search.token as string) || '',
        }
    },
})

const verifyEmailAtom = atomWithMutation(() => ({
    mutationKey: ['verify-email'],
    mutationFn: VerificationApi.verify,
}))

function VerifyEmailCallback() {
    const { token } = Route.useSearch()

    const [{ mutate, isIdle, isPending, isError }] = useAtom(verifyEmailAtom)

    useEffect(() => {
        if (!token) return
        mutate(token)
    }, [token])

    if (isIdle || isPending) return <Loader />

    if (!token || isError) {
        return (
            <div className="relative w-full h-screen flex flex-col items-center justify-center">
                <AnimatedBackground />
                <div className="relative z-10 w-full max-w-md px-4 py-4">
                    <div className="backdrop-blur-xl bg-card/40 border border-border/50 rounded-2xl p-6 sm:p-8 shadow-2xl">
                        <div className="relative z-10 text-center">
                            <div className="flex justify-center mb-4">
                                <CircleX className="size-12 sm:size-16 text-red-400" />
                            </div>
                            <h1 className="text-xl sm:text-2xl font-semibold text-foreground mb-2">
                                Invalid or expired verification link
                            </h1>
                            <p className="text-muted-foreground text-sm sm:text-base mb-4">
                                Please request a new verification email.
                            </p>
                            <Button
                                onClick={continueToApp}
                                className="w-full bg-linear-to-r from-primary to-accent hover:opacity-90"
                            >
                                Continue to app
                            </Button>
                        </div>
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className="relative w-full h-screen flex flex-col items-center justify-center overflow-y-auto">
            <AnimatedBackground />

            <div className="relative z-10 w-full max-w-md px-4 py-4">
                <div className="backdrop-blur-xl bg-card/40 border border-border/50 rounded-2xl p-6 sm:p-8 shadow-2xl">
                    <div className="text-center mb-4">
                        <div className="flex justify-center mb-4">
                            <MailCheck className="size-12 sm:size-16" />
                        </div>
                        <h1 className="text-xl sm:text-2xl font-semibold bg-linear-to-r from-primary to-accent bg-clip-text text-transparent mb-2">
                            Email verified!
                        </h1>
                        <p className="text-muted-foreground text-sm sm:text-base">
                            Your email has been successfully verified. You can now use TwinSpeak!
                        </p>
                    </div>

                    <Button
                        onClick={continueToApp}
                        className="w-full bg-linear-to-r from-primary to-accent hover:opacity-90"
                    >
                        Continue to app
                    </Button>
                </div>
            </div>
        </div>
    )
}

// To reset Referrer header
function continueToApp() {
    window.location.href = '/'
}
