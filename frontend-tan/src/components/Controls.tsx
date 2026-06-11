import { useState } from 'react'
import { RecordButton } from '@/components/ui/record-button'
import { cn } from '@/lib/utils'

export default function Controls(props: {
    className?: string
    isRecording: boolean
    disabled?: boolean
    start: () => void
    stop: () => void
}) {
    const [isDisabled, setIsDisabled] = useState(false)

    function onMicClick() {
        setIsDisabled(true)
        setTimeout(() => {
            setIsDisabled(false)
        }, 500)
        if (props.isRecording) {
            return props.stop()
        }
        return props.start()
    }

    return (
        <div className={cn('flex justify-center items-center', props.className)}>
            <RecordButton
                disabled={isDisabled || props.disabled}
                size="sm"
                onClick={onMicClick}
                isRecording={props.isRecording}
            />
        </div>
    )
}
