package gladia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fasthttp/websocket"
	"golang.org/x/sync/errgroup"
)

const (
	endpoint   = "https://api.gladia.io/v2/live"
	stopRecMsg = `{"type":"stop_recording"}`

	EventTranscript      = "transcript"
	EventTranslation     = "translation"
	EventPostTranscript  = "post_transcript"
	EventFinalTranscript = "post_final_transcript"
	EventError           = "error"
)

type Client struct {
	apiKey string
	client *http.Client
}

// Full docs is avalable at https://docs.gladia.io/api-reference/v2/live/init
type LiveSessionRequest struct {
	Encoding   string `json:"encoding,omitempty"`
	BitDepth   int    `json:"bit_depth,omitempty"`
	SampleRate int    `json:"sample_rate,omitempty"`

	LanguageConfig     *LanguageConfig     `json:"language_config,omitempty"`
	RealtimeProcessing *RealtimeProcessing `json:"realtime_processing,omitempty"`
	MessagesConfig     *MessagesConfig     `json:"messages_config,omitempty"`
}

type LanguageConfig struct {
	Languages     []string `json:"languages,omitempty"`
	CodeSwitching *bool    `json:"code_switching,omitempty"`
}

type RealtimeProcessing struct {
	Translation       *bool              `json:"translation,omitempty"`
	TranslationConfig *TranslationConfig `json:"translation_config,omitempty"`
}

type TranslationConfig struct {
	TargetLanguages []string `json:"target_languages"`
}

type MessagesConfig struct {
	ReceivePartialTranscripts       *bool `json:"receive_partial_transcripts,omitempty"`
	ReceiveFinalTranscripts         *bool `json:"receive_final_transcripts,omitempty"`
	ReceiveSpeechEvents             *bool `json:"receive_speech_events,omitempty"`
	ReceivePreProcessingEvents      *bool `json:"receive_pre_processing_events,omitempty"`
	ReceiveRealtimeProcessingEvents *bool `json:"receive_realtime_processing_events,omitempty"`
	ReceivePostProcessingEvents     *bool `json:"receive_post_processing_events,omitempty"`
	ReceiveAcknowledgments          *bool `json:"receive_acknowledgments,omitempty"`
	ReceiveErrors                   *bool `json:"receive_errors,omitempty"`
	ReceiveLifecycleEvents          *bool `json:"receive_lifecycle_events,omitempty"`
}

func Bool(v bool) *bool {
	return &v
}

type LiveSessionResponse struct {
	Id        string `json:"id"`
	CreatedAt string `json:"string"`
	Url       string `json:"url"`
}

type WSEvent struct {
	SessionId string          `json:"session_id"`
	CreatedAt string          `json:"created_at"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
}

type FTranscriptData struct {
	Metadata      FTranscriptMeta `json:"metadata"`
	Transcription FTranscriptItem `json:"transcription"`
	Translation   FtTranslateItem `json:"translation"`
}

type FTranscriptMeta struct {
	AudioDuration float32 `json:"audio_duration"`
}

type FTranscriptItem struct {
	FullTranscript string `json:"full_transcript"`
}

type FtTranslateItem struct {
	Results []FtTranslateResult `json:"results"`
}

type FtTranslateResult struct {
	FullTranscript string `json:"full_transcript"`
}

type Utterance struct {
	Text string `json:"text"`
}

type TransribeData struct {
	IsFinal   bool      `json:"is_final"`
	Utterance Utterance `json:"utterance"`
}

type TranslateData struct {
	TranslatedUtterance Utterance `json:"translated_utterance"`
}

var expectedWSClosures = []int{
	websocket.CloseNormalClosure,
	websocket.CloseGoingAway,
	websocket.CloseNoStatusReceived,
	websocket.CloseAbnormalClosure,
}

func (g *Client) LiveSession(ctx context.Context, inLang string, outLang string, in <-chan []byte, out chan<- WSEvent) error {
	defer close(out)

	payload := LiveSessionRequest{
		Encoding:   "wav/pcm",
		SampleRate: 16000,
		LanguageConfig: &LanguageConfig{
			Languages: []string{inLang},
		},
		RealtimeProcessing: &RealtimeProcessing{
			Translation: Bool(true),
			TranslationConfig: &TranslationConfig{
				TargetLanguages: []string{outLang},
			},
		},
		MessagesConfig: &MessagesConfig{
			ReceivePartialTranscripts:       Bool(false),
			ReceiveFinalTranscripts:         Bool(true),
			ReceiveSpeechEvents:             Bool(false),
			ReceivePreProcessingEvents:      Bool(false),
			ReceiveRealtimeProcessingEvents: Bool(true),
			ReceivePostProcessingEvents:     Bool(true),
			ReceiveAcknowledgments:          Bool(false),
			ReceiveErrors:                   Bool(true),
			ReceiveLifecycleEvents:          Bool(false),
		},
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Add("x-gladia-key", g.apiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("error status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var parsedBody LiveSessionResponse
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		return fmt.Errorf("cannot unmarshal resp: %w", err)
	}

	gladiaC, _, err := websocket.DefaultDialer.Dial(parsedBody.Url, nil)
	if err != nil {
		return fmt.Errorf("cannot start gladia ws session: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		defer gladiaC.Close()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				_, message, err := gladiaC.ReadMessage()
				if err != nil {
					if websocket.IsCloseError(err, expectedWSClosures...) {
						return nil
					} else {
						return fmt.Errorf("reading gladia message: %w", err)
					}
				}
				var parsedEvt WSEvent
				err = json.Unmarshal(message, &parsedEvt)
				if err != nil {
					return fmt.Errorf("cannot unmarshad gladia message: %w", err)
				}
				out <- parsedEvt
			}
		}
	})

	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				gladiaC.Close()
				return ctx.Err()
			case data, ok := <-in:
				if !ok {
					err = gladiaC.WriteMessage(websocket.TextMessage, []byte(stopRecMsg))
					if err != nil {
						gladiaC.Close()
						return fmt.Errorf("writing gladia stop message: %w", err)
					}
					return nil
				}
				err = gladiaC.WriteMessage(websocket.BinaryMessage, data)
				if err != nil {
					gladiaC.Close()
					return fmt.Errorf("writing gladia message: %w", err)
				}
			}
		}
	})

	if err = eg.Wait(); err != nil {
		return err
	}

	return nil
}

func NewClient(apiKey string) *Client {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &Client{
		apiKey: apiKey,
		client: client,
	}
}
