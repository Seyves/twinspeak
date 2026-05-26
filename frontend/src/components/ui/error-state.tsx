'use client'

import { cn } from '@/lib/utils'
import { AlertCircle } from 'lucide-react'
import type { ReactNode } from 'react'

export default function ErrorState(props: { children: ReactNode; className?: string }) {
    return (
        <div className="w-full max-w-sm space-y-2">
            <div className={cn('flex items-center text-destructive', props.className)}>
                <AlertCircle className="h-[1em] w-[1em] mr-[0.5em] shrink-0" />
                {props.children}
            </div>
        </div>
    )
}
