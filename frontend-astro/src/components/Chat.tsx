import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

export default function Chat(props: {
    children: ReactNode
    ref?: React.Ref<HTMLDivElement>
    className?: string
}) {
    return (
        <div
            ref={props.ref}
            className={cn('px-6 py-4 flex flex-col gap-3 min-h-0 overflow-auto', props.className)}
        >
            {props.children}
        </div>
    )
}
