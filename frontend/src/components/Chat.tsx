import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'
import { motion } from 'motion/react'

export default function Chat(props: {
    children: ReactNode
    ref?: React.Ref<HTMLDivElement>
    className?: string
}) {
    return (
        <motion.div
            // initial={{opacity:0}}
            // animate={{opacity:1}}
            ref={props.ref}
            id='chat'
            className={cn('px-6 py-4 flex flex-col gap-3 min-h-0 overflow-auto', props.className)}
        >
            {props.children}
        </motion.div>
    )
}
