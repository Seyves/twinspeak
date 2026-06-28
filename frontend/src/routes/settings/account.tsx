import { createFileRoute, useNavigate } from '@tanstack/react-router'
import * as AccountApi from '@/api/account'
import * as AuthApi from '@/api/auth'
import { useAtom } from 'jotai'
import { Calendar, LogOut, Zap, Mail } from 'lucide-react'
import { AnimatePresence } from 'motion/react'
import ErrorPage from '@/components/Error'
import { Button } from '@/components/ui/button'
import Loader from '@/components/ui/loader'
import SectionCard from '@/components/SectionCard'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { atomWithQuery, atomWithMutation, queryClientAtom } from 'jotai-tanstack-query'
import { toast } from 'sonner'
import { localThemeAtom } from '@/components/theme-provider'
import { chatMessageSize, themes, type ChatMessageSize, type Theme } from '@/definitions/chat'

export const Route = createFileRoute('/settings/account')({
    component: Account,
})

export const accountAtom = atomWithQuery(() => ({
    queryKey: ['account'],
    queryFn: async () => await Promise.all([AccountApi.getAccount(), AccountApi.getCredits()]),
}))

const passwordResetAtom = atomWithMutation(() => ({
    mutationKey: ['password-reset-settings'],
    mutationFn: AuthApi.requestPasswordReset,
}))

export const preferencesAtom = atomWithQuery(() => ({
    queryKey: ['preferences'],
    queryFn: AccountApi.getPreferences,
}))

export const updatePreferencesAtom = atomWithMutation((get) => ({
    mutationKey: ['update-preferences'],
    mutationFn: async (prefs: AccountApi.Preferences) => {
        await AccountApi.updatePreferences(prefs)
        return prefs
    },
    onMutate: (data) => {
        const queryClient = get(queryClientAtom)
        queryClient.setQueryData(['preferences'], data)
    },
}))

function Account() {
    const navigate = useNavigate()
    const [_, setLocalTheme] = useAtom(localThemeAtom)
    const [account] = useAtom(accountAtom)
    const [prefs] = useAtom(preferencesAtom)
    const [{ mutate: setPrefs }] = useAtom(updatePreferencesAtom)
    const [{ mutateAsync: sendPasswordReset, isPending: isResetting }] = useAtom(passwordResetAtom)

    async function signOut() {
        await AuthApi.signOut()
        navigate({ to: '/auth' })
    }

    function handleSendPasswordReset(email: string) {
        toast.promise(() => sendPasswordReset(email), {
            loading: 'Sending reset email...',
            success: `Password reset email sent to ${email}`,
            error: 'Failed to send reset email. Please try again.',
            position: 'top-right',
        })
    }

    return (
        <div className="relative h-full">
            <AnimatePresence>
                {(function () {
                    if (account.isPending || prefs.isPending) return <Loader key="loader" />

                    if (account.isError || prefs.isError) {
                        return (
                            <ErrorPage
                                key="error"
                                onRetry={() => {
                                    account.refetch()
                                    prefs.refetch()
                                }}
                            />
                        )
                    }

                    const [me, grants] = account.data

                    return (
                        <div key="content">
                            <div className="rounded-2xl border border-border/50 bg-card overflow-hidden mb-4">
                                <>
                                    <div className="px-4 pt-5 pb-4 flex items-center gap-4">
                                        <div className="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center shrink-0 overflow-hidden">
                                            {me.profilePicture ? (
                                                <img
                                                    src={me.profilePicture}
                                                    alt="Avatar"
                                                    className="w-full h-full object-cover"
                                                />
                                            ) : (
                                                <span className="text-xl font-semibold text-primary">
                                                    {getInitials(me.email)}
                                                </span>
                                            )}
                                        </div>
                                        <div className="flex-1 min-w-0">
                                            <p className="font-medium truncate">{me.email}</p>
                                            <span className="inline-flex items-center mt-1.5 py-0.5 text-sm rounded-full text-muted-foreground">
                                                Free Plan
                                            </span>
                                        </div>
                                    </div>
                                    <div className="px-4 py-2 sm:py-4 gap-2 flex">
                                        <Button
                                            variant="outline"
                                            onClick={() => handleSendPasswordReset(me.email)}
                                            disabled={isResetting}
                                            className="basis-1/2 mb-2"
                                        >
                                            <Mail className="w-3.5 h-3.5" />
                                            Reset password
                                        </Button>
                                        <Button
                                            variant="outline"
                                            onClick={signOut}
                                            className="basis-1/2 text-destructive border-destructive/20 hover:bg-destructive/10 hover:text-destructive hover:border-destructive/40"
                                        >
                                            <LogOut className="w-3.5 h-3.5" />
                                            Sign out
                                        </Button>
                                    </div>
                                </>
                            </div>

                            <SectionCard label="Credits">
                                {grants.length === 0 ? (
                                    <p className="text-sm text-muted-foreground py-1">
                                        No active credit grants
                                    </p>
                                ) : (
                                    <div className="space-y-3">
                                        {grants.map((grant) => (
                                            <CreditGrantCard key={grant.id} grant={grant} />
                                        ))}
                                    </div>
                                )}
                            </SectionCard>
                            <SectionCard label="Appearance">
                                <div className="space-y-5">
                                    <div className="flex items-center justify-between gap-4">
                                        <label className="text font-medium">Message size</label>
                                        <Select
                                            value={prefs.data.chatMessageSize}
                                            onValueChange={(v) =>
                                                setPrefs({
                                                    ...prefs.data,
                                                    chatMessageSize: v as ChatMessageSize,
                                                })
                                            }
                                        >
                                            <SelectTrigger className="w-44 text-base">
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem
                                                    className="text-base"
                                                    value={chatMessageSize.sm}
                                                >
                                                    Small
                                                </SelectItem>
                                                <SelectItem
                                                    className="text-base"
                                                    value={chatMessageSize.md}
                                                >
                                                    Medium
                                                </SelectItem>
                                                <SelectItem
                                                    className="text-base"
                                                    value={chatMessageSize.lg}
                                                >
                                                    Large
                                                </SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>

                                    <div className="flex items-center justify-between gap-4">
                                        <label className="font-medium">Theme</label>
                                        <Select
                                            value={prefs.data.theme}
                                            onValueChange={(v) => {
                                                const theme = v as Theme
                                                setLocalTheme(theme)
                                                setPrefs({ ...prefs.data, theme })
                                            }}
                                        >
                                            <SelectTrigger className="w-44 text-base">
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem
                                                    className="text-base"
                                                    value={themes.dark}
                                                >
                                                    Dark
                                                </SelectItem>
                                                <SelectItem
                                                    className="text-base"
                                                    value={themes.light}
                                                >
                                                    Light
                                                </SelectItem>
                                                <SelectItem
                                                    className="text-base"
                                                    value={themes.system}
                                                >
                                                    System
                                                </SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>
                                </div>
                            </SectionCard>
                        </div>
                    )
                })()}
            </AnimatePresence>
        </div>
    )
}

function CreditGrantCard({ grant }: { grant: AccountApi.CreditGrant }) {
    const pct = grant.amount > 0 ? (grant.remainingAmount / grant.amount) * 100 : 0
    const isMonthly = grant.type === 'monthly'

    return (
        <div className="border border-border rounded-2xl p-4 space-y-4">
            <div className="flex items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                    <div
                        className={`w-8 h-8 rounded-xl flex items-center justify-center shrink-0 ${
                            isMonthly ? 'bg-blue-500/15' : 'bg-violet-500/15'
                        }`}
                    >
                        {isMonthly ? (
                            <Calendar className="w-4 h-4 text-blue-400" />
                        ) : (
                            <Zap className="w-4 h-4 text-violet-400" />
                        )}
                    </div>
                    <span className="text-sm font-medium">{isMonthly ? 'Monthly' : 'Top-up'}</span>
                </div>
                {grant.expiresAt && (
                    <span className="text-sm text-muted-foreground shrink-0">
                        Expires {formatDate(grant.expiresAt)}
                    </span>
                )}
            </div>

            <div className="space-y-2">
                <div className="flex justify-between text-sm text-muted-foreground">
                    <span>{formatSeconds(grant.remainingAmount)} remaining</span>
                    <span>{formatSeconds(grant.amount)} total</span>
                </div>
                <div className="h-2 bg-muted/30 rounded-full overflow-hidden">
                    <div
                        className={`h-full rounded-full transition-all duration-500 ${
                            isMonthly ? 'bg-blue-400' : 'bg-violet-400'
                        }`}
                        style={{ width: `${pct}%` }}
                    />
                </div>
            </div>
        </div>
    )
}

function formatSeconds(seconds: number): string {
    const minutes = Math.floor(seconds / 60)
    const secs = seconds % 60
    if (minutes === 0) return `${secs}s`
    if (secs === 0) return `${minutes}m`
    return `${minutes}m ${secs}s`
}

function formatDate(iso: string): string {
    return new Date(iso).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
    })
}

function getInitials(email: string): string {
    return email.charAt(0).toUpperCase()
}
