import { cn } from '@/lib/utils'
import { motion } from 'motion/react'
import type { ReactNode } from 'react'

export default function ChatMessage(props: {
    children: ReactNode
    type: 'incoming' | 'outgoing'
    size?: 'sm' | 'md' | 'lg'
    className?: string
}) {
    const sizeClass = {
        sm: 'text-base px-4 py-3',
        md: 'text-lg px-4 py-3',
        lg: 'text-xl px-4 py-4',
    }[props.size ?? 'md']

    return (
        <div
            className={cn(
                'flex gap-3',
                props.type === 'incoming' ? 'justify-start' : 'justify-end',
            )}
        >
            <div
                className={cn(
                    'inline-flex max-w-xs lg:max-w-md rounded-4xl leading-relaxed transition-colors duration-700',
                    sizeClass,
                    props.type === 'incoming'
                        ? 'border border-input text-foreground rounded-tl-sm'
                        : 'bg-accent text-foreground rounded-tr-sm border-primary/20',
                    props.className
                )}
            >
                <motion.div
                    initial={false}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.2 }}
                >
                    {props.children}
                </motion.div>
            </div>
        </div>
    )
}
