package main

import (
	"flag"
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/parlyx/backend/providers"
	"github.com/parlyx/backend/server"
	"github.com/spf13/viper"
)

func main() {
	cfgPath := flag.String("c", "/etc/parlyx/config.yaml", "Path to configuration file")
	flag.Parse()

	var cfg Config
	err := ParseConfig(*cfgPath, &cfg)
	if err != nil {
		log.Errorf("Parsing config file: %s", err.Error())
		return
	}

	transcriber, err := providers.NewTranscriber(cfg.Transcription)
	if err != nil {
		log.Errorf("Creating transcriber: %s", err.Error())
		return
	}

	translator, err := providers.NewTranslater(cfg.Translation)
	if err != nil {
		log.Errorf("Creating translator: %s", err.Error())
		return
	}

	api := server.NewRestApi(transcriber, translator, cfg.Host)
	err = api.Start()
	if err != nil {
		log.Errorf("Starting server: %s", err.Error())
		return
	}
}

type Config struct {
	Host          string                   `mapstructure:"host"`
	Transcription providers.ProviderConfig `mapstructure:"transcription"`
	Translation   providers.ProviderConfig `mapstructure:"translation"`
}

func ParseConfig(path string, cfg any) error {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error unmarshalling config: %s", err.Error())
	}
	return nil
}
