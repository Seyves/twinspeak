'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'

interface RecordButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    isRecording: boolean
    size?: 'sm' | 'default' | 'lg'
}

const MicrophoneIcon = ({ className }: { className?: string }) => (
    <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className={className}
    >
        <path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z" />
        <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
        <line x1="12" x2="12" y1="19" y2="22" />
    </svg>
)

const RecordButton = React.forwardRef<HTMLButtonElement, RecordButtonProps>(
    (
        {
            className,
            isRecording,
            size = 'default',
            ...props
        },
        ref,
    ) => {
        const sizeClasses = {
            sm: 'size-20',
            default: 'size-28',
            lg: 'size-36',
        }

        const iconSizeClasses = {
            sm: 'size-6',
            default: 'size-10',
            lg: 'size-14',
        }

        return (
            <button
                ref={ref}
                type="button"
                className={cn(
                    'group relative flex items-center justify-center rounded-full',
                    'transition-all duration-300 ease-out',
                    'hover:scale-105',
                    'active:scale-95',
                    'focus:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950',
                    'disabled:opacity-50 disabled:pointer-events-none',
                    sizeClasses[size],
                    className,
                )}
                onClick={props.onClick}
                aria-label={isRecording ? 'Stop recording' : 'Start recording'}
                aria-pressed={isRecording}
                {...props}
            >
                {/* Outer glow effect */}
                <span
                    className={cn(
                        'absolute inset-0 rounded-full',
                        'bg-primary/20 blur-xl',
                        'animate-[glow-pulse_3s_ease-in-out_infinite]',
                        isRecording &&
                            'bg-primary/40 animate-[glow-pulse-recording_1s_ease-in-out_infinite]',
                    )}
                />

                {/* Glowing ring border */}
                <span
                    className={cn(
                        'absolute inset-0 rounded-full',
                        'border-2 border-primary',
                        'shadow-[0_0_15px_rgba(var(--primary),0.5),0_0_30px_rgba(var(--primary),0.3),inset_0_0_15px_rgba(var(--primary),0.1)]',
                        'animate-[ring-glow_3s_ease-in-out_infinite]',
                        isRecording &&
                            'border-primary shadow-[0_0_20px_rgba(34,211,238,0.7),0_0_40px_rgba(34,211,238,0.5),inset_0_0_20px_rgba(34,211,238,0.2)] animate-[ring-glow-recording_1s_ease-in-out_infinite]',
                    )}
                />

                {/* Ripple effect when recording */}
                {isRecording && (
                    <>
                        <span className="absolute inset-0 rounded-full border-2 border-primary/50 animate-[ripple_2s_ease-out_infinite]" />
                        <span className="absolute inset-0 rounded-full border-2 border-primary/30 animate-[ripple_2s_ease-out_infinite_0.6s]" />
                        <span className="absolute inset-0 rounded-full border-2 border-primary/20 animate-[ripple_2s_ease-out_infinite_1.2s]" />
                    </>
                )}

                {/* Dark inner circle */}
                <span
                    className={cn(
                        'relative z-10 flex items-center justify-center rounded-full',
                        'bg-card',
                        size === 'sm' ? 'size-[4.5rem]' : size === 'lg' ? 'size-[8rem]' : 'size-24',
                    )}
                >
                    {/* Microphone icon */}
                    <MicrophoneIcon
                        className={cn(
                            'text-primary transition-all duration-300',
                            iconSizeClasses[size],
                            isRecording &&
                                'text-primary animate-[icon-pulse_1s_ease-in-out_infinite]',
                        )}
                    />
                </span>
            </button>
        )
    },
)
RecordButton.displayName = 'RecordButton'

export { RecordButton }
