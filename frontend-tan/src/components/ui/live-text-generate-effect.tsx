'use client'

import { motion, stagger, useAnimate } from 'motion/react'
import * as React from 'react'
import { cn } from '@/lib/utils'

type TextGenerateEffectProps = Omit<React.ComponentProps<'div'>, 'children'> & {
    words: string
    disable?: boolean
    filter?: boolean
    duration?: number
    staggerDelay?: number
}

function TextGenerateEffect({
    ref,
    words,
    className,
    filter = true,
    duration = 0.2,
    staggerDelay = 0.1,
    ...props
}: TextGenerateEffectProps) {
    const localRef = React.useRef<HTMLDivElement>(null)

    React.useImperativeHandle(ref as any, () => localRef.current as HTMLDivElement)

    const [scope, animate] = useAnimate()

    const wordsArray = React.useMemo(() => words.split(' '), [words])

    // сколько слов уже было отрендерено ранее
    const prevLengthRef = React.useRef(0)

    React.useEffect(() => {
        if (!scope.current) return

        const prevLength = prevLengthRef.current
        const newWordsCount = wordsArray.length - prevLength

        // если новых слов нет — ничего не делаем
        if (newWordsCount <= 0) {
            prevLengthRef.current = wordsArray.length
            return
        }

        // берём только новые элементы
        const spans = scope.current.querySelectorAll('span')
        const newSpans = Array.from(spans).slice(prevLength)

        animate(
            newSpans,
            {
                opacity: 1,
                filter: filter ? 'blur(0px)' : 'none',
            },
            {
                duration,
                delay: stagger(staggerDelay),
            },
        )

        prevLengthRef.current = wordsArray.length
    }, [animate, duration, filter, scope, staggerDelay, wordsArray])

    return (
        <div
            className={cn('font-bold inline', className)}
            data-slot="text-generate-effect"
            ref={localRef}
            {...(props as any)}
        >
            <motion.div className="inline" ref={scope}>
                {wordsArray.map((word, idx) => {
                    const isNew = idx >= prevLengthRef.current

                    return (
                        <motion.span
                            key={`${word}-${idx}`}
                            className="will-change-transform will-change-opacity will-change-filter"
                            initial={
                                props.disable
                                    ? false
                                    : {
                                          opacity: isNew ? 0 : 1,
                                          filter: filter
                                              ? isNew
                                                  ? 'blur(10px)'
                                                  : 'blur(0px)'
                                              : 'none',
                                      }
                            }
                        >
                            {word}{' '}
                        </motion.span>
                    )
                })}
            </motion.div>
        </div>
    )
}

export { TextGenerateEffect, type TextGenerateEffectProps }
export default TextGenerateEffect
