package providers

import (
	"fmt"

	"github.com/twinspeak/backend/providers/libretranslate"
)

type Translater interface {
	Translate(inputLang string, outputLang string, text string) (string, error)
}

func NewTranslater(cfg ProviderConfig) (Translater, error) {
	available := []string{libretranslate.Name}

	switch cfg.Provider {
	case libretranslate.Name:
		client, err := libretranslate.NewClient(cfg.Url)
		if err != nil {
			return nil, fmt.Errorf("cannot create libretranslate client: %s", err.Error())
		}
		return client, nil
	default:
		return nil, fmt.Errorf("unsupported translation provider: %s, available: %v", cfg.Provider, available)
	}
}
