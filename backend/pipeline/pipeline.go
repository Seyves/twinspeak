package pipeline

import "context"

// Pipe method recieves raw binary (wav/pce 16000Hz 16 bit)
// and outputs speech events. Method pipe is sync, so it will return nil
// only if recording was stopped and other size finished processing.
// Returns error only if initialization went wrong.
// To stop recording `in` channel should be closed.
type SpeechPipeline interface {
	Pipe(ctx context.Context, c LangConfig, in <-chan []byte, out chan<- SpeechEvent) (duration int, err error)
	SupportedLanguages(ctx context.Context) (languages map[string]string)
}

func NewSpeechEvent(t SpeechEventType, p any) SpeechEvent {
	return SpeechEvent{
		Type:    t,
		Payload: p,
	}
}

type LangConfig struct {
	In  string
	Out string
}

type SpeechEventType int

const (
	LiveTranscriptEvt  SpeechEventType = iota // payload: string
	LiveTranslateEvt                          // payload: string
	FinalTranscriptEvt                        // payload: string
	FinalTranslateEvt                         // payload: string
	DurationEvt                               // payload: int (ms)
)

type SpeechEvent struct {
	Type    SpeechEventType `json:"type"`
	Payload any             `json:"payload"`
}
