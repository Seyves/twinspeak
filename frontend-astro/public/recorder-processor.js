class RecorderProcessor extends AudioWorkletProcessor {
    process(inputs) {
        const input = inputs[0]

        if (!input || !input[0]) {
            return true
        }

        this.port.postMessage(input[0])

        return true
    }
}

registerProcessor('wav-recorder-processor', RecorderProcessor)
