import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { ArrowLeft, LogOut } from 'lucide-react'
import { ThemeProvider, useTheme } from '@/components/theme-provider'
import { usePreferences } from '@/hooks/usePreferences'
import { getMe, type UserInfo } from '@/api/user'
import { signOut } from '@/api/auth'

function getInitials(email: string): string {
    return email.charAt(0).toUpperCase()
}

function SettingsContent() {
    const { setTheme, theme } = useTheme()
    const { preferences, updatePreferences } = usePreferences()
    const [userInfo, setUserInfo] = useState<UserInfo | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const fetchUserInfo = async () => {
            try {
                const info = await getMe()
                setUserInfo(info)
            } catch (e) {
                console.error('Failed to fetch user info:', e)
                setError('Failed to load account information')
            } finally {
                setLoading(false)
            }
        }
        fetchUserInfo()
    }, [])

    const handleLogout = async () => {
        try {
            await signOut()
        } catch (e) {
            console.error('Logout failed:', e)
        }
    }

    return (
        <div className="relative font-sans h-dvh overflow-auto flex flex-col cursor-default bg-background text-foreground">
            {/* Header */}
            <div className="sticky top-0 z-10 border-b border-border/50 bg-card/30 px-6 py-4 flex items-center gap-3">
                <a href="/" className="cursor-pointer">
                    <Button
                        variant="ghost"
                        size="icon-sm"
                        className="rounded-full hover:bg-primary/10 cursor-pointer"
                    >
                        <ArrowLeft className="w-4 h-4" />
                    </Button>
                </a>
                <h1 className="text-lg font-semibold">Settings</h1>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-auto px-6 py-6">
                <div className="max-w-2xl mx-auto space-y-6">
                    {/* Account Section */}
                    {!loading && userInfo && (
                        <div className="border border-border/50 bg-card/50 rounded-lg p-6 space-y-4">
                            <h2 className="text-base font-semibold">Account</h2>

                            {/* Avatar and Email */}
                            <div className="flex items-center gap-4">
                                <div className="w-16 h-16 rounded-full bg-accent/20 flex items-center justify-center flex-shrink-0">
                                    {userInfo.profilePicture ? (
                                        <img
                                            src={userInfo.profilePicture}
                                            alt="Avatar"
                                            className="w-full h-full rounded-full object-cover"
                                        />
                                    ) : (
                                        <span className="text-xl font-semibold text-accent">
                                            {getInitials(userInfo.email)}
                                        </span>
                                    )}
                                </div>
                                <div className="flex-1 min-w-0">
                                    <p className="text-sm text-muted-foreground mb-1">Email</p>
                                    <p className="text-foreground break-all">{userInfo.email}</p>
                                </div>
                            </div>

                            {/* Plan Badge */}
                            <div className="pt-2">
                                <div className="flex items-center gap-2">
                                    <span className="inline-block px-3 py-1 text-xs rounded-full bg-muted/30 text-muted-foreground">
                                        Free Plan <span className="ml-1 opacity-60">(TODO)</span>
                                    </span>
                                </div>
                            </div>

                            {/* Logout Button */}
                            <div className="pt-4 border-t border-border/50">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={handleLogout}
                                    className="w-full text-destructive hover:bg-destructive/10 hover:text-destructive cursor-pointer"
                                >
                                    <LogOut className="w-4 h-4 mr-2" />
                                    Log out
                                </Button>
                            </div>
                        </div>
                    )}

                    {loading && (
                        <div className="border border-border/50 bg-card/50 rounded-lg p-6">
                            <p className="text-muted-foreground">Loading account information...</p>
                        </div>
                    )}

                    {error && (
                        <div className="border border-destructive/30 bg-destructive/10 rounded-lg p-6">
                            <p className="text-destructive">{error}</p>
                        </div>
                    )}

                    {/* Preferences Section */}
                    <div className="border border-border/50 bg-card/50 rounded-lg p-6 space-y-6">
                        <h2 className="text-base font-semibold">Preferences</h2>

                        {/* Chat Font Size */}
                        <div className="space-y-3">
                            <label className="block text-sm font-medium">Chat message size</label>
                            <div className="flex gap-2">
                                {(['sm', 'md', 'lg'] as const).map((size) => (
                                    <Button
                                        key={size}
                                        variant={
                                            preferences.chatFontSize === size
                                                ? 'default'
                                                : 'outline'
                                        }
                                        size="sm"
                                        onClick={() => updatePreferences({ chatFontSize: size })}
                                        className="min-w-12 cursor-pointer"
                                    >
                                        {size === 'sm' ? 'S' : size === 'md' ? 'M' : 'L'}
                                    </Button>
                                ))}
                            </div>
                        </div>

                        {/* Theme Toggle */}
                        <div className="space-y-3">
                            <label className="block text-sm font-medium">Theme</label>
                            <div className="flex gap-2">
                                {(['dark', 'light', 'system'] as const).map((t) => (
                                    <Button
                                        key={t}
                                        variant={theme === t ? 'default' : 'outline'}
                                        size="sm"
                                        onClick={() => setTheme(t)}
                                        className="capitalize min-w-20 cursor-pointer"
                                    >
                                        {t}
                                    </Button>
                                ))}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default function Settings() {
    return (
        <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
            <SettingsContent />
        </ThemeProvider>
    )
}
