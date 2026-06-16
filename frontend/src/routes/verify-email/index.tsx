import { createFileRoute } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { useState } from 'react'
import { toast } from 'sonner'
import { resendVerificationEmail } from '@/api/auth'
import { AnimatedBackground } from '@/components/animated-background'
import { MailWarning } from 'lucide-react'

export const Route = createFileRoute('/verify-email/')({
    component: VerifyEmail,
})

function VerifyEmail() {
    const [isResending, setIsResending] = useState(false)

    const handleResend = async () => {
        setIsResending(true)
        try {
            await resendVerificationEmail()
            toast.success('Verification email sent! Check your inbox.', {
                position: 'top-right',
                richColors: true,
            })
        } catch (error) {
            toast.error('Failed to send email. Please try again.', {
                position: 'top-right',
                richColors: true,
            })
        } finally {
            setIsResending(false)
        }
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

                    <div className="space-y-4">
                        <Button
                            onClick={handleResend}
                            disabled={isResending}
                            variant="outline"
                            className="w-full"
                        >
                            {isResending ? 'Sending...' : 'Resend verification email'}
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    )
}
