import { useState } from 'react'
import { AnimatedBackground } from '@/components/animated-background'
import { redirectToGoogleAuth, signIn, signUp } from '@/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { Field, FieldError, FieldLabel } from '#/components/ui/field'
import { useForm } from '@tanstack/react-form'
import { toast } from 'sonner'
import { HTTPError } from 'ky'
import { createServerFn } from '@tanstack/react-start'
import { getCookie } from '@tanstack/react-start/server'

export const checkAuthServerFn = createServerFn().handler(async () => {
    const refreshToken = getCookie('refresh_token')
    return { session: Boolean(refreshToken) }
})

export const Route = createFileRoute('/auth/')({
    component: Auth,
    beforeLoad: async () => {
        const auth = await checkAuthServerFn()
        if (auth.session) throw Route.redirect({ to: '/' })
    },
})

type AuthMode = 'signin' | 'signup'

function Auth() {
    const navigate = useNavigate()
    const [mode, setMode] = useState<AuthMode>('signin')

    const form = useForm({
        defaultValues: {
            email: '',
            password: '',
            confirmPassword: '',
        },
        onSubmit: async ({ value }) => {
            if (mode === 'signin') {
                toast.promise(() => signIn(value.email, value.password), {
                    loading: 'Working on it...',
                    success: () => {
                        navigate({ to: '/' })
                        return 'You are successfully logged in!'
                    },
                    error: (e) => {
                        if (e instanceof HTTPError && e.response.status === 401) {
                            return 'Email or password is wrong'
                        } else {
                            return 'Something went wrong'
                        }
                    },
                    position: 'top-right',
                    richColors: true,
                })
            } else {
                await signUp(value.email, value.password)
                navigate({ to: '/' })
            }
        },
    })

    const handleGoogleAuth = () => {
        redirectToGoogleAuth()
    }

    return (
        <div className="relative w-full h-screen flex flex-col items-center justify-center overflow-y-auto">
            <AnimatedBackground />

            {/* Content - scrollable on small screens */}
            <div className="relative z-10 w-full max-w-md px-4 py-4 sm:py-8 my-auto">
                <div className="backdrop-blur-xl bg-card/40 border border-border/50 rounded-2xl p-6 sm:p-8 shadow-2xl">
                    {/* Header */}
                    <div className="mb-6 sm:mb-8 text-center">
                        <h1 className="text-2xl sm:text-3xl font-semibold bg-linear-to-r from-primary to-accent bg-clip-text text-transparent mb-1 sm:mb-2">
                            TwinSpeak
                        </h1>
                        <p className="text-muted-foreground text-sm sm:text-base">
                            Real-time multilanguage conversation
                        </p>
                    </div>

                    {/* Mode Toggle */}
                    <div className="flex gap-2 mb-6 sm:mb-8 bg-muted p-1 rounded-lg">
                        <Button
                            onClick={() => setMode('signin')}
                            variant={mode === 'signin' ? 'default' : 'ghost'}
                            className="flex-1"
                        >
                            Sign In
                        </Button>
                        <Button
                            onClick={() => setMode('signup')}
                            variant={mode === 'signup' ? 'default' : 'ghost'}
                            className="flex-1"
                        >
                            Sign Up
                        </Button>
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
                            name="email"
                            validators={{
                                onBlur: ({ value }) => {
                                    return !value.includes('@')
                                        ? 'Please provide valid email address'
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
                                            Email
                                        </FieldLabel>
                                        <Input
                                            id={field.name}
                                            name={field.name}
                                            type="email"
                                            value={field.state.value}
                                            onBlur={field.handleBlur}
                                            onChange={(e) => field.handleChange(e.target.value)}
                                            aria-invalid={isInvalid}
                                            placeholder="you@example.com"
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
                            name="password"
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
                                            Password
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

                        {mode === 'signup' && (
                            <form.Field
                                name="confirmPassword"
                                validators={{
                                    onBlurListenTo: ['password'],
                                    onBlur: ({ value, fieldApi }) => {
                                        if (
                                            mode === 'signup' &&
                                            value !==
                                                (fieldApi.form.getFieldValue('password') ?? '')
                                        ) {
                                            return 'Passwords do not match'
                                        }
                                        return undefined
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
                                                Confirm password
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
                        )}

                        {/* Submit Button */}
                        <Button
                            type="submit"
                            className="h-10 w-full mt-4 sm:mt-6 bg-linear-to-r from-primary to-accent hover:opacity-90 transform hover:scale-105 active:scale-95 shadow-lg"
                        >
                            {mode === 'signin' ? 'Sign In' : 'Create Account'}
                        </Button>
                    </form>

                    {/* Divider */}
                    <div className="relative my-4 sm:my-6">
                        <div className="absolute inset-0 flex items-center">
                            <div className="w-full border-t border-border/50"></div>
                        </div>
                        <div className="relative flex justify-center text-xs sm:text-sm">
                            <span className="px-2 bg-card/40 text-muted-foreground">
                                Or continue with
                            </span>
                        </div>
                    </div>

                    {/* Google OAuth Button */}
                    <Button onClick={handleGoogleAuth} variant="outline" className="w-full h-10">
                        <svg
                            className="w-4 h-4 sm:w-5 sm:h-5"
                            viewBox="0 0 24 24"
                            fill="currentColor"
                        >
                            <path
                                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                                fill="#4285F4"
                            />
                            <path
                                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                                fill="#34A853"
                            />
                            <path
                                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                                fill="#FBBC05"
                            />
                            <path
                                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                                fill="#EA4335"
                            />
                        </svg>
                        <span className="hidden sm:inline">
                            {mode === 'signin' ? 'Sign in' : 'Sign up'} with Google
                        </span>
                        <span className="sm:hidden">
                            {mode === 'signin' ? 'Sign in' : 'Sign up'}
                        </span>
                    </Button>

                    {/* Footer Text */}
                    <p className="text-center text-xs text-muted-foreground mt-4 sm:mt-6">
                        {mode === 'signin' ? (
                            <>
                                Don't have an account?{' '}
                                <Button
                                    onClick={() => setMode('signup')}
                                    variant="link"
                                    className="h-auto p-0 text-xs"
                                >
                                    Sign up
                                </Button>
                            </>
                        ) : (
                            <>
                                Already have an account?{' '}
                                <Button
                                    onClick={() => setMode('signin')}
                                    variant="link"
                                    className="h-auto p-0 text-xs"
                                >
                                    Sign in
                                </Button>
                            </>
                        )}
                    </p>
                </div>

                {/* Bottom Info */}
                <p className="text-center text-xs text-muted-foreground mt-4 sm:mt-8 px-2">
                    By continuing, you agree to our Terms of Service and Privacy Policy
                </p>
            </div>
        </div>
    )
}
