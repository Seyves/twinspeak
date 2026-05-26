class GladiaProcessor extends AudioWorkletProcessor {
    constructor() {
        super()
        // Buffer size: 16000 samples/sec * 0.1 sec = 1600 samples (~100ms chunks)
        // This balances latency with efficiency
        this.bufferSize = 1600
        this.buffer = new Float32Array(this.bufferSize)
        this.bufferIndex = 0
    }

    process(inputs) {
        const input = inputs[0]

        if (!input || !input[0]) {
            return true
        }

        const channelData = input[0]

        for (let i = 0; i < channelData.length; i++) {
            this.buffer[this.bufferIndex] = channelData[i]
            this.bufferIndex++

            // Send buffer when full
            if (this.bufferIndex >= this.bufferSize) {
                // Send a copy to avoid issues with buffer reuse
                this.port.postMessage(this.buffer.slice(0, this.bufferIndex))
                this.bufferIndex = 0
            }
        }

        return true
    }
}

registerProcessor('gladia-processor', GladiaProcessor)
