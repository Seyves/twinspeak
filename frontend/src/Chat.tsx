import { ReactNode } from 'react'
import { cn } from './lib/utils';

export default function Chat(props: { children: ReactNode; className?: string }) {
    return <div className={cn('p-4 flex flex-col gap-2 min-h-0 overflow-auto', props.className)}>{props.children}</div>
}
