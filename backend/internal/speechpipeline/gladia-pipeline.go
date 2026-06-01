package speechpipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twinspeak/backend/internal/clients/gladia"
	"golang.org/x/sync/errgroup"
)

type GladiaPipeline struct {
	gladia gladia.Client
}

func (g *GladiaPipeline) Pipe(ctx context.Context, c LangConfig, in <-chan []byte, out chan<- Event) error {
	defer close(out)
	receiver := make(chan gladia.WSEvent)

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		err := g.gladia.LiveSession(ctx, c.In, c.Out, in, receiver)
		if err != nil {
			return fmt.Errorf("processing gladia live session: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case evt, ok := <-receiver:
				if !ok {
					return nil
				}
				switch evt.Type {
				case gladia.EventError:
					return fmt.Errorf("error event occured: %s", evt)
				case gladia.EventTranscript:
					var data gladia.TransribeData
					err := json.Unmarshal(evt.Data, &data)
					if err != nil {
						return fmt.Errorf("cannot unmarshal transcript data: %w", err)
					}
					if !data.IsFinal {
						continue
					}
					out <- NewSpeechEvent(LiveTranscriptEvt, data.Utterance.Text)
				case gladia.EventTranslation:
					var data gladia.TranslateData
					err := json.Unmarshal(evt.Data, &data)
					if err != nil {
						return fmt.Errorf("cannot unmarshal translate data: %w", err)
					}
					out <- NewSpeechEvent(LiveTranslateEvt, data.TranslatedUtterance.Text)
				case gladia.EventPostTranscript:
					continue
				case gladia.EventFinalTranscript:
					var data gladia.FTranscriptData
					err := json.Unmarshal(evt.Data, &data)
					if err != nil {
						return fmt.Errorf("cannot unmarshal final transcript data: %w", err)
					}
					out <- NewSpeechEvent(FinalTranscriptEvt, data.Transcription.FullTranscript)
					if len(data.Translation.Results) > 0 {
						out <- NewSpeechEvent(FinalTranslateEvt, data.Translation.Results[0].FullTranscript)
					} else {
						out <- NewSpeechEvent(FinalTranslateEvt, "")
					}
				}
			}
		}
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (g *GladiaPipeline) SupportedLanguages(ctx context.Context) map[string]string {
	return map[string]string{
		"af":  "Afrikaans",
		"sq":  "Albanian",
		"am":  "Amharic",
		"ar":  "Arabic",
		"hy":  "Armenian",
		"as":  "Assamese",
		"az":  "Azerbaijani",
		"ba":  "Bashkir",
		"eu":  "Basque",
		"be":  "Belarusian",
		"bn":  "Bengali",
		"bs":  "Bosnian",
		"br":  "Breton",
		"bg":  "Bulgarian",
		"ca":  "Catalan",
		"zh":  "Chinese",
		"hr":  "Croatian",
		"cs":  "Czech",
		"da":  "Danish",
		"nl":  "Dutch",
		"en":  "English",
		"et":  "Estonian",
		"fo":  "Faroese",
		"fi":  "Finnish",
		"fr":  "French",
		"gl":  "Galician",
		"ka":  "Georgian",
		"de":  "German",
		"el":  "Greek",
		"gu":  "Gujarati",
		"ht":  "Haitian Creole",
		"ha":  "Hausa",
		"haw": "Hawaiian",
		"he":  "Hebrew",
		"hi":  "Hindi",
		"hu":  "Hungarian",
		"is":  "Icelandic",
		"id":  "Indonesian",
		"it":  "Italian",
		"ja":  "Japanese",
		"jw":  "Javanese",
		"kn":  "Kannada",
		"kk":  "Kazakh",
		"km":  "Khmer",
		"ko":  "Korean",
		"lo":  "Lao",
		"la":  "Latin",
		"lv":  "Latvian",
		"ln":  "Lingala",
		"lt":  "Lithuanian",
		"lb":  "Luxembourgish",
		"mk":  "Macedonian",
		"mg":  "Malagasy",
		"ms":  "Malay",
		"ml":  "Malayalam",
		"mt":  "Maltese",
		"mi":  "Maori",
		"mr":  "Marathi",
		"mn":  "Mongolian",
		"my":  "Myanmar",
		"ne":  "Nepali",
		"no":  "Norwegian",
		"nn":  "Nynorsk",
		"oc":  "Occitan",
		"ps":  "Pashto",
		"fa":  "Persian",
		"pl":  "Polish",
		"pt":  "Portuguese",
		"pa":  "Punjabi",
		"ro":  "Romanian",
		"ru":  "Russian",
		"sa":  "Sanskrit",
		"sr":  "Serbian",
		"sn":  "Shona",
		"sd":  "Sindhi",
		"si":  "Sinhala",
		"sk":  "Slovak",
		"sl":  "Slovenian",
		"so":  "Somali",
		"es":  "Spanish",
		"su":  "Sundanese",
		"sw":  "Swahili",
		"sv":  "Swedish",
		"tl":  "Tagalog",
		"tg":  "Tajik",
		"ta":  "Tamil",
		"tt":  "Tatar",
		"te":  "Telugu",
		"th":  "Thai",
		"bo":  "Tibetan",
		"tr":  "Turkish",
		"tk":  "Turkmen",
		"uk":  "Ukrainian",
		"ur":  "Urdu",
		"uz":  "Uzbek",
		"vi":  "Vietnamese",
		"cy":  "Welsh",
		"wo":  "Wolof",
		"yi":  "Yiddish",
		"yo":  "Yoruba",
	}
}

func NewGladiaPipeline(apiKey string) (*GladiaPipeline, error) {
	return &GladiaPipeline{
		gladia: *gladia.NewClient(apiKey),
	}, nil
}
