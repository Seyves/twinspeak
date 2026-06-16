'use client'

import { motion } from 'motion/react'
import * as React from 'react'
import { cn } from '@/lib/utils'

type TextGenerateEffectProps = Omit<React.ComponentProps<'div'>, 'children'> & {
    words: string
    disable?: boolean
    duration?: number
}

function TextGenerateEffect({
    ref,
    words,
    className,
    disable = false,
    duration = 0.2,
    ...props
}: TextGenerateEffectProps) {
    const wordsArray = React.useMemo(() => words.split(' '), [words])

    return (
        <div
            className={cn('font-bold inline', className)}
            data-slot="text-generate-effect"
            {...(props as any)}
        >
            <motion.div className="inline">
                {(function () {
                    if (disable) return <span>{wordsArray.join(' ')}</span>
                    return wordsArray.map((word, idx) => {
                        return (
                            <motion.span
                                key={`${word}-${idx}`}
                                className="will-change-transform will-change-opacity will-change-filter"
                                transition={{ duration: duration }}
                                initial={{
                                    opacity: 0,
                                    filter: 'blur(10px)',
                                }}
                                animate={{
                                    opacity: 1,
                                    filter: 'blur(0px)',
                                }}
                            >
                                {word}{' '}
                            </motion.span>
                        )
                    })
                })()}
            </motion.div>
        </div>
    )
}

export { TextGenerateEffect, type TextGenerateEffectProps }
export default TextGenerateEffect
