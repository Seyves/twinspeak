import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import { resendVerificationEmail, signOut } from '@/api/auth'
import { AnimatedBackground } from '@/components/animated-background'
import { MailWarning, LogOut } from 'lucide-react'
import { atomWithMutation } from 'jotai-tanstack-query'
import { useAtom } from 'jotai'

export const Route = createFileRoute('/verify-email/')({
    component: VerifyEmail,
})

const resendVerificationEmailAtom = atomWithMutation(() => ({
    mutationKey: ['send-verification-email'],
    mutationFn: resendVerificationEmail,
}))

function VerifyEmail() {
    const navigate = useNavigate()
    const [{ mutateAsync, isPending }] = useAtom(resendVerificationEmailAtom)

    const handleResend = async () => {
        toast.promise(mutateAsync, {
            loading: 'Working on it...',
            success: () => {
                navigate({ to: '/' })
                return 'Verification email sent! Check your inbox.'
            },
            error: 'Something went wrong :(',
            position: 'top-right',
        })
    }

    const handleSignOut = async () => {
        await signOut()
        navigate({ to: '/auth' })
    }

    return (
        <div className="relative w-full h-screen flex flex-col items-center justify-center overflow-y-auto">
            <AnimatedBackground />

            <div className="relative z-10 w-full max-w-md px-4 py-4">
                <div className="backdrop-blur-xl bg-card/40 border border-border/50 rounded-2xl p-6 sm:p-8 shadow-2xl">
                    <div className="text-center mb-4">
                        <div className="flex justify-center mb-4">
                            <MailWarning className="size-12 sm:size-16" />
                        </div>
                        <h1 className="text-xl sm:text-2xl font-semibold bg-linear-to-r from-primary to-accent bg-clip-text text-transparent mb-2">
                            Verify your email
                        </h1>
                        <p className="text-muted-foreground text-sm sm:text-base">
                            We sent a verification link to your email address. Please check your
                            inbox and click the link to continue.
                        </p>
                    </div>

                    <div className="space-y-2">
                        <Button
                            onClick={handleResend}
                            disabled={isPending}
                            variant="outline"
                            className="w-full"
                        >
                            {isPending ? 'Sending...' : 'Resend verification email'}
                        </Button>
                        <Button
                            variant="outline"
                            onClick={handleSignOut}
                            className="w-full text-destructive border-destructive/20 hover:bg-destructive/10 hover:text-destructive hover:border-destructive/40"
                        >
                            <LogOut className="w-3.5 h-3.5" />
                            Sign out
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    )
}
