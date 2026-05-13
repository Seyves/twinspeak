'use client'

import * as React from 'react'
import { motion } from 'motion/react'
import { cn } from '@/lib/utils'

type Props = Omit<React.ComponentProps<'div'>, 'children'> & {
    text: string
    filter?: boolean
    duration?: number

    /**
     * How many latest words stay "live"
     * and can be corrected/reanimated.
     */
    liveWindow?: number
}

type Word = {
    id: number
    value: string
    animated: boolean
}

export default function LiveTranscriptEffect({
    text,
    className,
    filter = true,
    duration = 0.22,
    liveWindow = 5,
    ...props
}: Props) {
    const idRef = React.useRef(0)

    const [stableWords, setStableWords] = React.useState<Word[]>([])

    const [liveWords, setLiveWords] = React.useState<Word[]>([])

    const prevWordsRef = React.useRef<string[]>([])

    React.useEffect(() => {
        const nextWords = text.trim().split(/\s+/).filter(Boolean)

        const prevWords = prevWordsRef.current

        /**
         * Split transcript:
         *
         * stable = everything except last N words
         * live   = last N words
         */
        const stablePart = nextWords.slice(0, Math.max(0, nextWords.length - liveWindow))

        const livePart = nextWords.slice(-liveWindow)

        /**
         * Freeze stable words
         */
        const prevStableLength = Math.max(0, prevWords.length - liveWindow)

        if (stablePart.length > prevStableLength) {
            const newlyStable = stablePart.slice(prevStableLength)

            setStableWords((current) => [
                ...current,
                ...newlyStable.map((word) => ({
                    id: idRef.current++,
                    value: word,
                    animated: false,
                })),
            ])
        }

        /**
         * Rebuild ONLY live window
         *
         * This allows ASR corrections
         * while keeping rest immutable.
         */
        const prevLive = prevWords.slice(-liveWindow)

        const rebuiltLive = livePart.map((word, index) => {
            const isNew = prevLive[index] !== word

            return {
                id: isNew ? idRef.current++ : (liveWords[index]?.id ?? idRef.current++),
                value: word,
                animated: isNew,
            }
        })

        setLiveWords(rebuiltLive)

        prevWordsRef.current = nextWords
    }, [text, liveWindow])

    return (
        <div
            className={cn('leading-relaxed whitespace-pre-wrap wrap-break-word', className)}
            {...props}
        >
            {/* Frozen transcript */}
            {stableWords.map((word) => (
                <React.Fragment key={word.id}>
                    <span>{word.value}</span>{' '}
                </React.Fragment>
            ))}

            {/* Live correction window */}
            {liveWords.map((word) =>
                word.animated ? (
                    <React.Fragment key={word.id}>
                        <motion.span
                            initial={{
                                opacity: 0,
                                y: 10,
                                filter: filter ? 'blur(8px)' : 'none',
                            }}
                            animate={{
                                opacity: 1,
                                y: 0,
                                filter: filter ? 'blur(0px)' : 'none',
                            }}
                            transition={{
                                duration,
                                ease: 'easeOut',
                            }}
                            className="inline"
                        >
                            {word.value}
                        </motion.span>{' '}
                    </React.Fragment>
                ) : (
                    <React.Fragment key={word.id}>
                        <span>{word.value}</span>{' '}
                    </React.Fragment>
                ),
            )}
        </div>
    )
}
