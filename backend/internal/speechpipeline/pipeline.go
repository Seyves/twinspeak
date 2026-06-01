package speechpipeline

import "context"

// Pipe method recieves raw binary (wav/pce 16000Hz 16 bit)
// and outputs speech events. Method pipe is sync, so it will return nil
// only if recording was stopped and other size finished processing.
// Returns error only if initialization went wrong.
// To stop recording `in` channel should be closed.
type Pipeline interface {
	Pipe(ctx context.Context, c LangConfig, in <-chan []byte, out chan<- Event) error
	SupportedLanguages(ctx context.Context) (languages map[string]string)
}

func NewSpeechEvent(t EventType, p any) Event {
	return Event{
		Type:    t,
		Payload: p,
	}
}

type LangConfig struct {
	In  string
	Out string
}

type EventType int

const (
	LiveTranscriptEvt  EventType = iota // payload: string
	LiveTranslateEvt                    // payload: string
	FinalTranscriptEvt                  // payload: string
	FinalTranslateEvt                   // payload: string
	ErrorEvt                            // payload: string
)

type Event struct {
	Type    EventType `json:"type"`
	Payload any       `json:"payload"`
}
