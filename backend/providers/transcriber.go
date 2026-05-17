package providers

import (
	"fmt"
	"io"

	"github.com/twinspeak/backend/providers/faster-whisper"
)

type Transcriber interface {
	// Reader must be a multipart/form-data body
	Transcribe(lang string, multipartHeader string, contentLenght string, r io.Reader) (string, error)
}

func NewTranscriber(cfg ProviderConfig) (Transcriber, error) {
	available := []string{fasterwhisper.Name}

	switch cfg.Provider {
	case fasterwhisper.Name:
		client, err := fasterwhisper.NewClient(cfg.Url)
		if err != nil {
			return nil, fmt.Errorf("cannot create faster_whisper client: %s", err.Error())
		}
		return client, nil
	default:
		return nil, fmt.Errorf("unsupported transcription provider: %s, available: %v", cfg.Provider, available)
	}
}
