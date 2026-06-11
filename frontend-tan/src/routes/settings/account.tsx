import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { type CreditGrant } from '@/api/user'
import { useAtom } from 'jotai'
import { Calendar, LogOut, Zap } from 'lucide-react'
import { AnimatePresence } from 'motion/react'
import ErrorPage from '@/components/Error'
import { signOut as signOutReq } from '@/api/auth'
import { Button } from '@/components/ui/button'
import Loader from '@/components/ui/loader'
import { accountAtom } from '@/atoms/account'
import SectionCard from '@/components/SectionCard'

export const Route = createFileRoute('/settings/account')({
    component: Account,
})

function CreditGrantCard({ grant }: { grant: CreditGrant }) {
    const pct = grant.amount > 0 ? (grant.remainingAmount / grant.amount) * 100 : 0
    const isMonthly = grant.type === 'monthly'

    return (
        <div className="border border-border rounded-2xl p-4 space-y-4">
            <div className="flex items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                    <div
                        className={`w-9 h-9 rounded-xl flex items-center justify-center shrink-0 ${
                            isMonthly ? 'bg-blue-500/15' : 'bg-violet-500/15'
                        }`}
                    >
                        {isMonthly ? (
                            <Calendar className="w-5 h-5 text-blue-400" />
                        ) : (
                            <Zap className="w-5 h-5 text-violet-400" />
                        )}
                    </div>
                    <span className="text font-medium">{isMonthly ? 'Monthly' : 'Top-up'}</span>
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

function Account() {
    const navigate = useNavigate()
    const [{ data, isPending, isError, refetch }] = useAtom(accountAtom)

    async function signOut() {
        await signOutReq()
        navigate({ to: '/auth' })
    }

    return (
        <div className="relative h-full">
            <AnimatePresence>
                {(function () {
                    if (isPending) return <Loader key="loader" />

                    if (isError) return <ErrorPage key="error" onRetry={refetch} />

                    const [account, grants] = data

                    return (
                        <div key="content">
                            {/* Profile card */}
                            <div className="rounded-2xl border border-border/50 bg-card overflow-hidden mb-4">
                                <>
                                    <div className="px-4 pt-5 pb-4 flex items-center gap-4">
                                        <div className="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center shrink-0 overflow-hidden">
                                            {account.profilePicture ? (
                                                <img
                                                    src={account.profilePicture}
                                                    alt="Avatar"
                                                    className="w-full h-full object-cover"
                                                />
                                            ) : (
                                                <span className="text-xl font-semibold text-primary">
                                                    {getInitials(account.email)}
                                                </span>
                                            )}
                                        </div>
                                        <div className="flex-1 min-w-0">
                                            <p className="font-medium truncate">{account.email}</p>
                                            <span className="inline-flex items-center mt-1.5 py-0.5 text-sm rounded-full text-muted-foreground">
                                                Free Plan
                                            </span>
                                        </div>
                                    </div>
                                    <div className="px-4 pb-4 border-t border-border/50 pt-3">
                                        <Button
                                            variant="outline"
                                            size="lg"
                                            onClick={signOut}
                                            className="w-full text-destructive border-destructive/20 hover:bg-destructive/10 hover:text-destructive hover:border-destructive/40"
                                        >
                                            <LogOut className="w-3.5 h-3.5" />
                                            Sign out
                                        </Button>
                                    </div>
                                </>
                            </div>

                            {/* Credits */}
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
                        </div>
                    )
                })()}
            </AnimatePresence>
        </div>
    )
}
