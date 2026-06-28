import { Button } from '@/components/ui/button'
import { ArrowLeft } from 'lucide-react'
import { createFileRoute, Link, Outlet } from '@tanstack/react-router'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/settings')({
    component: Settings,
})

type Tab = {
    name: string
    to: string
}

function Settings() {
    const tabs: Tab[] = [
        {
            name: 'account',
            to: '/settings/account',
        },
    ]

    return (
        <div className="font-sans h-dvh flex flex-col bg-background text-foreground">
            <div className="sticky top-0 z-10 border-b border-border/40 bg-background/80 backdrop-blur-sm px-4 py-3 flex items-center gap-3">
                <Link to={'/'}>
                    <Button
                        variant="ghost"
                        size="icon-sm"
                        className="rounded-full hover:bg-primary/10"
                    >
                        <ArrowLeft className="w-4 h-4" />
                    </Button>
                </Link>
                <h1 className="text-xl font-semibold">Settings</h1>
            </div>

            <div className="border-b border-border/40 bg-background px-4 flex">
                {tabs.map((tab) => (
                    <Link key={tab.name} to={tab.to}>
                        {({ isActive }) => (
                            <button
                                className={cn(
                                    'relative px-4 py-3 text font-medium capitalize cursor-pointer transition-colors',
                                    isActive
                                        ? 'text-foreground'
                                        : 'text-muted-foreground hover:text-foreground',
                                )}
                            >
                                {tab.name}
                                {isActive && (
                                    <span className="absolute bottom-0 left-2 right-2 h-0.5 bg-primary rounded-t-full" />
                                )}
                            </button>
                        )}
                    </Link>
                ))}
            </div>

            <div className="flex-1 overflow-auto px-4 py-5">
                <div className="max-w-lg mx-auto space-y-4 h-full">
                    <Outlet />
                </div>
            </div>
        </div>
    )
}
