package speechpipeline

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/twinspeak/backend/internal/clients/fasterwhisper"
	"github.com/twinspeak/backend/internal/clients/libretranslate"
)

type WhisperPipeline struct {
	whisper        *fasterwhisper.Client
	libretranslate *libretranslate.Client
}

func (p *WhisperPipeline) Pipe(ctx context.Context, c LangConfig, in <-chan []byte, out chan<- Event) error {
	defer close(out)
	transcription, _, err := p.whisper.Transcribe(ctx, c.In, in)
	if err != nil {
		log.Errorf("Whisper transcribe: %s", err.Error())
		return err
	}
	out <- NewSpeechEvent(FinalTranscriptEvt, transcription)
	translation, err := p.libretranslate.Translate(ctx, c.In, c.Out, transcription)
	if err != nil {
		log.Errorf("Libretranslate translate: %s", err.Error())
		return err
	}
	out <- NewSpeechEvent(FinalTranslateEvt, translation)
	return nil
}

func NewWhisperPipeline(whisperPath string, librePath string) (*WhisperPipeline, error) {
	whisper, err := fasterwhisper.NewClient(whisperPath)
	if err != nil {
		return nil, fmt.Errorf("cannot create whisper client: %w", err)
	}
	libretranslate, err := libretranslate.NewClient(librePath)
	if err != nil {
		return nil, fmt.Errorf("cannot create libretranslate client: %w", err)
	}
	return &WhisperPipeline{
		whisper:        whisper,
		libretranslate: libretranslate,
	}, nil
}

func (g *WhisperPipeline) SupportedLanguages(ctx context.Context) map[string]string {
	return map[string]string{
		"sq": "Albanian",
		"ar": "Arabic",
		"az": "Azerbaijani",
		"bg": "Bulgarian",
		"ca": "Catalan",
		"zh": "Chinese",
		"cs": "Czech",
		"da": "Danish",
		"nl": "Dutch",
		"et": "Estonian",
		"fi": "Finnish",
		"fr": "French",
		"gl": "Galician",
		"de": "German",
		"el": "Greek",
		"he": "Hebrew",
		"hi": "Hindi",
		"hu": "Hungarian",
		"id": "Indonesian",
		"it": "Italian",
		"ja": "Japanese",
		"ko": "Korean",
		"lv": "Latvian",
		"lt": "Lithuanian",
		"ms": "Malay",
		"nb": "Norwegian",
		"fa": "Persian",
		"pl": "Polish",
		"pt": "Portuguese",
		"ro": "Romanian",
		"ru": "Russian",
		"sk": "Slovak",
		"sl": "Slovenian",
		"es": "Spanish",
		"sv": "Swedish",
		"tl": "Tagalog",
		"th": "Thai",
		"tr": "Turkish",
		"uk": "Ukrainian",
		"ur": "Urdu",
		"vi": "Vietnamese",
	}
}
