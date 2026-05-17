'use client'

import { cn } from '@/lib/utils'
import { AlertCircle } from 'lucide-react'
import type { ReactNode } from 'react'

export default function ErrorState(props: { children: ReactNode; className?: string }) {
    return (
        <div className="w-full max-w-sm space-y-2">
            <div className={cn('flex items-center gap-2 text-destructive', props.className)}>
                <AlertCircle className="h-4 w-4" />
                {props.children}
            </div>
        </div>
    )
}
