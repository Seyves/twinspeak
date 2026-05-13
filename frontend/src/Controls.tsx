import { RecordButton } from './components/ui/record-button'
import { cn } from './lib/utils'

export default function Controls(props: {
    className?: string
    isRecording: boolean
    start: () => void
    stop: () => void
}) {
    function onMicClick() {
        if (props.isRecording) {
            return props.stop()
        }
        return props.start()
    }

    return (
        <div
            className={cn(
                'flex justify-center items-center bg-background',
                props.className,
            )}
        >
            <RecordButton size="sm" onClick={onMicClick} isRecording={props.isRecording}/>
        </div>
    )
}
