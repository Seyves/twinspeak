import { cn } from '@/lib/utils'
import { ReactNode } from 'react'

export default function ChatMessage(props: { children: ReactNode; type: 'incoming' | 'outgoing' }) {
    return (
        <div className={cn('flex', props.type === 'incoming' ? 'justify-start' : 'justify-end')}>
            <div
                className={cn(
                    'rounded-4xl inline-flex text-xl border p-4',
                    props.type === 'incoming'
                        ? 'rounded-tl-none border-(--color-border)'
                        : 'border-secondary bg-card rounded-tr-none',
                )}
            >
                {props.children}
            </div>
        </div>
    )
}
