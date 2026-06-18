import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { BrushCleaning, MoreVertical, Settings } from 'lucide-react'
import { useAtom } from 'jotai'
import { atomWithMutation, queryClientAtom } from 'jotai-tanstack-query'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import * as AccountApi from '@/api/account'

export const clearChatAtom = atomWithMutation((get) => ({
    mutationKey: ['clear-chat'],
    mutationFn: async () => {
        await AccountApi.clearChat()
    },
    onSuccess: () => {
        const queryClient = get(queryClientAtom)
        queryClient.invalidateQueries({ queryKey: ['messages'] })
    },
}))

export function OverflowMenu() {
    const [showConfirm, setShowConfirm] = useState(false)
    const [{ mutate: clearChat, isPending: isClearing }] = useAtom(clearChatAtom)

    const handleClearChat = () => {
        const promise = new Promise((resolve, reject) => {
            clearChat(undefined, {
                onSuccess: resolve,
                onError: reject,
            })
        })

        toast.promise(promise, {
            loading: 'Clearing chat...',
            success: 'Chat cleared successfully',
            error: 'Failed to clear chat',
            position: "top-right"
        })

        setShowConfirm(false)
    }

    return (
        <>
            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button
                        variant="ghost"
                        size="icon"
                        className="rounded-full hover:bg-primary/10"
                    >
                        <MoreVertical className="size-5 text-muted-foreground" />
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-auto" sideOffset={8}>
                    <DropdownMenuItem asChild>
                        <Link to="/settings" className="flex cursor-pointer items-center">
                            <Settings className="mr-2 size-5" />
                            <span className="text-base">Settings</span>
                        </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem
                        onClick={() => setShowConfirm(true)}
                        className="text-destructive focus:text-destructive"
                    >
                        <BrushCleaning className="mr-2 size-5" />
                        <span className="text-base">Clear chat</span>
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>

            <AlertDialog open={showConfirm} onOpenChange={setShowConfirm}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Clear chat history?</AlertDialogTitle>
                        <AlertDialogDescription>
                            This will hide all your current chat messages. New messages will still
                            appear normally. This action cannot be undone.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel disabled={isClearing}>Cancel</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={handleClearChat}
                            disabled={isClearing}
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        >
                            {isClearing ? 'Clearing...' : 'Clear chat'}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </>
    )
}
