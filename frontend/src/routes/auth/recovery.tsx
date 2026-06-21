import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import * as AuthApi from '@/api/auth'
import { AnimatedBackground } from '@/components/animated-background'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { CircleX, CheckCircle, Lock } from 'lucide-react'
import { atomWithMutation } from 'jotai-tanstack-query'
import { useAtom } from 'jotai'
import { Field, FieldError, FieldLabel } from '@/components/ui/field'
import { useForm } from '@tanstack/react-form'
import { toast } from 'sonner'
import { HTTPError } from 'ky'

export const Route = createFileRoute('/auth/recovery')({
    component: PasswordRecovery,
    validateSearch: (search: Record<string, unknown>) => {
        return {
            token: (search.token as string) || '',
        }
    },
})

const resetPasswordAtom = atomWithMutation(() => ({
    mutationKey: ['reset-password'],
    mutationFn: ({ token, password }: { token: string; password: string }) =>
        AuthApi.confirmPasswordReset(token, password),
}))

function PasswordRecovery() {
    const { token } = Route.useSearch()
    const [isSuccess, setIsSuccess] = useState(false)
    const [{ mutateAsync, isPending }] = useAtom(resetPasswordAtom)

    const form = useForm({
        defaultValues: {
            password: '',
            confirmPassword: '',
        },
        onSubmit: async ({ value }) => {
            if (!token) return

            toast.promise(() => mutateAsync({ token, password: value.password }), {
                loading: 'Resetting password...',
                success: () => {
                    setIsSuccess(true)
                    return 'Password updated successfully!'
                },
                error: async (e) => {
                    if (!(e instanceof HTTPError)) return 'Something went wrong'
                    switch (e.response.status) {
                        case 400:
                            return 'Invalid or expired reset link'
                        default:
                            return 'Something went wrong'
                    }
                },
                position: 'top-right',
            })
        },
    })

    if (!token) {
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
                                Invalid reset link
                            </h1>
                            <p className="text-muted-foreground text-sm sm:text-base mb-4">
                                This password reset link is invalid. Please request a new one.
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

    if (isSuccess) {
        return (
            <div className="relative w-full h-screen flex flex-col items-center justify-center overflow-y-auto">
                <AnimatedBackground />

                <div className="relative z-10 w-full max-w-md px-4 py-4">
                    <div className="backdrop-blur-xl bg-card/40 border border-border/50 rounded-2xl p-6 sm:p-8 shadow-2xl">
                        <div className="text-center mb-4">
                            <div className="flex justify-center mb-4">
                                <CheckCircle className="size-12 sm:size-16 text-green-400" />
                            </div>
                            <h1 className="text-xl sm:text-2xl font-semibold bg-linear-to-r from-primary to-accent bg-clip-text text-transparent mb-2">
                                Password updated!
                            </h1>
                            <p className="text-muted-foreground text-sm sm:text-base">
                                Your password has been successfully updated. You can now sign in
                                with your new password.
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

    return (
        <div className="bg-[#f1f2f7] dark:bg-[#060607] relative w-full h-screen flex flex-col items-center justify-center overflow-y-auto">
            <AnimatedBackground />

            {/* Content - scrollable on small screens */}
            <div className="relative z-10 w-full max-w-md px-4 py-4 sm:py-8 my-auto">
                <div className="backdrop-blur-xl bg-card/80 dark:bg-card/40 border border-border shadow-2xl rounded-2xl p-6 sm:p-8">
                    {/* Header */}
                    <div className="mb-6 sm:mb-8 text-center">
                        <div className="flex justify-center mb-4">
                            <div className="w-12 h-12 sm:w-16 sm:h-16 rounded-full bg-primary/10 flex items-center justify-center">
                                <Lock className="size-6 sm:size-8 text-primary" />
                            </div>
                        </div>
                        <h1 className="text-2xl sm:text-3xl font-semibold bg-linear-to-r from-primary to-accent bg-clip-text text-transparent mb-1 sm:mb-2">
                            Reset Password
                        </h1>
                        <p className="text-muted-foreground text-sm sm:text-base">
                            Enter your new password below
                        </p>
                    </div>

                    {/* Form */}
                    <form
                        onSubmit={(e) => {
                            e.preventDefault()
                            form.handleSubmit()
                        }}
                        className="space-y-3 sm:space-y-4"
                        noValidate
                    >
                        <form.Field
                            name="password"
                            validators={{
                                onBlur: ({ value }) => {
                                    return value.length < 6
                                        ? 'Password must be at least 6 characters long'
                                        : undefined
                                },
                            }}
                            children={(field) => {
                                const isInvalid =
                                    field.state.meta.isTouched && !field.state.meta.isValid
                                return (
                                    <Field
                                        data-invalid={isInvalid}
                                        className="space-y-1 sm:space-y-2"
                                    >
                                        <FieldLabel
                                            htmlFor={field.name}
                                            className="text-xs sm:text-sm"
                                        >
                                            New Password
                                        </FieldLabel>
                                        <Input
                                            id={field.name}
                                            name={field.name}
                                            type="password"
                                            value={field.state.value}
                                            onBlur={field.handleBlur}
                                            onChange={(e) => field.handleChange(e.target.value)}
                                            aria-invalid={isInvalid}
                                            placeholder="••••••••"
                                            className="h-10 text-sm"
                                            required
                                        />
                                        {isInvalid && (
                                            <FieldError
                                                errors={[{ message: field.state.meta.errors[0] }]}
                                            />
                                        )}
                                    </Field>
                                )
                            }}
                        ></form.Field>

                        <form.Field
                            name="confirmPassword"
                            validators={{
                                onChangeListenTo: ['password'],
                                onChange: ({ value, fieldApi }) => {
                                    if (value !== (fieldApi.form.getFieldValue('password') ?? '')) {
                                        return 'Passwords do not match'
                                    }
                                    return undefined
                                },
                            }}
                            children={(field) => {
                                const passwordField = form.getFieldMeta('password')
                                const isInvalid =
                                    field.state.meta.isBlurred &&
                                    passwordField?.isBlurred &&
                                    !field.state.meta.isValid
                                return (
                                    <Field
                                        data-invalid={isInvalid}
                                        className="space-y-1 sm:space-y-2"
                                    >
                                        <FieldLabel
                                            htmlFor={field.name}
                                            className="text-xs sm:text-sm"
                                        >
                                            Confirm Password
                                        </FieldLabel>
                                        <Input
                                            id={field.name}
                                            name={field.name}
                                            type="password"
                                            value={field.state.value}
                                            onBlur={field.handleBlur}
                                            onChange={(e) => field.handleChange(e.target.value)}
                                            aria-invalid={isInvalid}
                                            placeholder="••••••••"
                                            className="h-10 text-sm"
                                            required
                                        />
                                        {isInvalid && (
                                            <FieldError
                                                errors={[
                                                    {
                                                        message: field.state.meta.errors[0],
                                                    },
                                                ]}
                                            />
                                        )}
                                    </Field>
                                )
                            }}
                        ></form.Field>

                        {/* Submit Button */}
                        <Button
                            type="submit"
                            disabled={isPending}
                            className="h-10 w-full mt-4 sm:mt-6 bg-linear-to-r from-primary to-accent hover:opacity-90 transform hover:scale-105 active:scale-95 shadow-lg"
                        >
                            {isPending ? 'Resetting...' : 'Reset Password'}
                        </Button>
                    </form>

                    {/* Footer Text */}
                    <p className="text-center text-xs text-muted-foreground mt-4 sm:mt-6">
                        Changed your mind?{' '}
                        <Button
                            onClick={continueToApp}
                            variant="link"
                            className="h-auto p-0 text-xs"
                        >
                            Continue to app
                        </Button>
                    </p>
                </div>
            </div>
        </div>
    )
}

// To reset Referrer header
function continueToApp() {
    window.location.href = '/'
}
