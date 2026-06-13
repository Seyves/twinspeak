import { AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface ErrorStateProps {
    message?: string
    onRetry?: () => void
}

export default function ErrorPage(props: ErrorStateProps) {
    return (
        <div
            key="error"
            className="relative w-full h-full flex flex-col items-center justify-center overflow-hidden rounded-3xl p-4"
        >
            <div className="flex justify-center">
                <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl border border-destructive/20 bg-destructive/10">
                    <AlertTriangle className="h-6 w-6 text-destructive" />
                </div>
            </div>

            <h3 className="text-center font-semibold tracking-tight">Something went wrong</h3>

            <p className="text-sm mt-2 text-center text-muted-foreground">
                {props.message || 'We hit an unexpected error while processing your request.'}
            </p>

            <div className="mt-6 flex justify-center gap-3">
                <Button onClick={props.onRetry}>Retry</Button>
            </div>
        </div>
    )
}
