import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { verifyEmail } from '@/api/auth'
import { AnimatedBackground } from '@/components/animated-background'
import Loader from '@/components/ui/loader'
import { Button } from '@/components/ui/button'
import { CircleX, MailCheck } from 'lucide-react'

export const Route = createFileRoute('/verify-email/callback')({
    component: VerifyEmailCallback,
    validateSearch: (search: Record<string, unknown>) => {
        return {
            token: (search.token as string) || '',
        }
    },
})

type VerificationState = 'verifying' | 'success' | 'error'

function VerifyEmailCallback() {
    const { token } = Route.useSearch()
    const navigate = useNavigate()
    const [state, setState] = useState<VerificationState>('verifying')
    const [errorMessage, setErrorMessage] = useState<string>('')

    useEffect(() => {
        if (!token) {
            setState('error')
            setErrorMessage('Invalid verification link')
            return
        }

        verifyEmail(token)
            .then(() => {
                setState('success')
            })
            .catch((err) => {
                setState('error')
                setErrorMessage('Invalid or expired verification link')
            })
    }, [token])

    if (state === 'verifying') {
        return <Loader />
    }

    if (state === 'error') {
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
                                {errorMessage}
                            </h1>
                            <p className="text-muted-foreground text-sm sm:text-base">
                                Please request a new verification email.
                            </p>
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
                            Your email has been successfully verified. You can
                            now use TwinSpeak!
                        </p>
                    </div>

                    <Button
                        onClick={() => navigate({ to: '/' })}
                        className="w-full bg-linear-to-r from-primary to-accent hover:opacity-90"
                    >
                        Continue to app
                    </Button>
                </div>
            </div>
        </div>
    )
}
