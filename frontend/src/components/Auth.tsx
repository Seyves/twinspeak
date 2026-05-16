import React, { useState } from 'react'
import { AnimatedBackground } from './ui/animated-background'

type AuthMode = 'signin' | 'signup'

export function Auth() {
    const [mode, setMode] = useState<AuthMode>('signin')
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [confirmPassword, setConfirmPassword] = useState('')

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()
        // TODO: Implement actual auth functionality
        if (mode === 'signin') {
            console.log('TODO: Sign in with', email, password)
        } else {
            console.log('TODO: Sign up with', email, password, confirmPassword)
        }
    }

    const handleGoogleAuth = () => {
        // TODO: add state param 
        const query = new URLSearchParams()
        query.set(
            'client_id',
            '229914232766-n45dpg5dtohqqbv5k7hnsi2khd5egbb8.apps.googleusercontent.com',
        )
        query.set('redirect_uri', 'https://localhost:5173/google-oauth')
        query.set('response_type', 'code')
        query.set('scope', 'openid')
        query.set('access_type', 'online')
        window.location.replace(`https://accounts.google.com/o/oauth2/auth?${query.toString()}`)
    }

    return (
        <div className="relative w-full h-screen flex flex-col items-center justify-center overflow-y-auto">
            <AnimatedBackground />

            {/* Content - scrollable on small screens */}
            <div className="relative z-10 w-full max-w-md px-4 py-4 sm:py-8 my-auto">
                <div className="backdrop-blur-xl bg-card/40 border border-border/50 rounded-2xl p-6 sm:p-8 shadow-2xl">
                    {/* Header */}
                    <div className="mb-6 sm:mb-8 text-center">
                        <h1 className="text-2xl sm:text-3xl font-semibold bg-gradient-to-r from-primary to-accent bg-clip-text text-transparent mb-1 sm:mb-2">
                            TwinSpeak
                        </h1>
                        <p className="text-muted-foreground text-xs sm:text-sm">
                            Real-time multilanguage conversation
                        </p>
                    </div>

                    {/* Mode Toggle */}
                    <div className="flex gap-2 mb-6 sm:mb-8 bg-muted p-1 rounded-lg">
                        <button
                            onClick={() => setMode('signin')}
                            className={`flex-1 py-2 px-4 rounded-md font-medium text-sm transition-all duration-200 ${
                                mode === 'signin'
                                    ? 'bg-primary text-primary-foreground shadow-lg'
                                    : 'text-muted-foreground hover:text-foreground'
                            }`}
                        >
                            Sign In
                        </button>
                        <button
                            onClick={() => setMode('signup')}
                            className={`flex-1 py-2 px-4 rounded-md font-medium text-sm transition-all duration-200 ${
                                mode === 'signup'
                                    ? 'bg-primary text-primary-foreground shadow-lg'
                                    : 'text-muted-foreground hover:text-foreground'
                            }`}
                        >
                            Sign Up
                        </button>
                    </div>

                    {/* Form */}
                    <form onSubmit={handleSubmit} className="space-y-3 sm:space-y-4">
                        {/* Email Input */}
                        <div>
                            <label
                                htmlFor="email"
                                className="block text-xs sm:text-sm font-medium mb-1 sm:mb-2"
                            >
                                Email
                            </label>
                            <input
                                id="email"
                                type="email"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                placeholder="you@example.com"
                                className="w-full px-3 sm:px-4 py-2 text-sm bg-input border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
                                required
                            />
                        </div>

                        {/* Password Input */}
                        <div>
                            <label
                                htmlFor="password"
                                className="block text-xs sm:text-sm font-medium mb-1 sm:mb-2"
                            >
                                Password
                            </label>
                            <input
                                id="password"
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                placeholder="••••••••"
                                className="w-full px-3 sm:px-4 py-2 text-sm bg-input border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
                                required
                            />
                        </div>

                        {/* Confirm Password (Sign Up Only) */}
                        {mode === 'signup' && (
                            <div>
                                <label
                                    htmlFor="confirmPassword"
                                    className="block text-xs sm:text-sm font-medium mb-1 sm:mb-2"
                                >
                                    Confirm Password
                                </label>
                                <input
                                    id="confirmPassword"
                                    type="password"
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
                                    placeholder="••••••••"
                                    className="w-full px-3 sm:px-4 py-2 text-sm bg-input border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
                                    required
                                />
                            </div>
                        )}

                        {/* Submit Button */}
                        <button
                            type="submit"
                            className="w-full mt-4 sm:mt-6 py-2 px-4 text-sm sm:text-base bg-gradient-to-r from-primary to-accent hover:opacity-90 text-primary-foreground font-semibold rounded-lg transition-all duration-200 transform hover:scale-105 active:scale-95 shadow-lg"
                        >
                            {mode === 'signin' ? 'Sign In' : 'Create Account'}
                        </button>
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
                    <button
                        onClick={handleGoogleAuth}
                        className="w-full py-2 px-4 text-sm sm:text-base border border-border/50 rounded-lg hover:bg-muted/50 transition-all duration-200 flex items-center justify-center gap-2 sm:gap-3 font-medium"
                    >
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
                    </button>

                    {/* Footer Text */}
                    <p className="text-center text-xs text-muted-foreground mt-4 sm:mt-6">
                        {mode === 'signin' ? (
                            <>
                                Don't have an account?{' '}
                                <button
                                    onClick={() => setMode('signup')}
                                    className="text-primary hover:text-accent font-medium transition-colors"
                                >
                                    Sign up
                                </button>
                            </>
                        ) : (
                            <>
                                Already have an account?{' '}
                                <button
                                    onClick={() => setMode('signin')}
                                    className="text-primary hover:text-accent font-medium transition-colors"
                                >
                                    Sign in
                                </button>
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
