package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"github.com/twinspeak/backend/auth"
	"github.com/twinspeak/backend/db"
	"github.com/twinspeak/backend/metrics"
	"github.com/twinspeak/backend/pipeline"
	"github.com/twinspeak/backend/server"
)

func main() {
	cfgPath := flag.String("c", "/etc/twinspeak/config.yaml", "Path to configuration file")
	flag.Parse()

	var cfg Config
	err := ParseConfig(*cfgPath, &cfg)
	if err != nil {
		log.Errorf("Parsing config file: %s", err.Error())
		return
	}

	pool, err := pgxpool.New(context.Background(), cfg.DBUrl)
	if err != nil {
		log.Errorf("Connecting to DB: %s", err.Error())
		return
	}
	defer pool.Close()
	queries := db.New(pool)

	authm := auth.NewAuth(pool, queries, cfg.HMACSecret)
	googleOauth := auth.NewGoogleOauth(cfg.Google, queries, cfg.HMACSecret)

	var p pipeline.SpeechPipeline
	switch cfg.Pipeline {
	case "gladia":
		p, err = pipeline.NewGladiaPipeline(cfg.GladiaKey)
		if err != nil {
			log.Errorf("Creating Gladia pipeline: %s", err.Error())
			return
		}
	case "whisper":
		p, err = pipeline.NewWhisperPipeline(cfg.FasterWhisperUrl, cfg.LibretranslateUrl)
		if err != nil {
			log.Errorf("Creating Whisper pipeline: %s", err.Error())
			return
		}
	}

	metricsm := metrics.NewMetrics(pool, queries)

	api := server.NewRestApi(cfg.Host, googleOauth, authm, p, metricsm)
	err = api.Start()
	if err != nil {
		log.Errorf("Starting server: %s", err.Error())
		return
	}
}

type Config struct {
	Host              string                 `mapstructure:"host"`
	HMACSecret        string                 `mapstructure:"hmac-secret"`
	DBUrl             string                 `mapstructure:"db-url"`
	Pipeline          string                 `mapstructure:"pipeline"`
	GladiaKey         string                 `mapstructure:"gladia-key"`
	FasterWhisperUrl  string                 `mapstructure:"faster-whisper-url"`
	LibretranslateUrl string                 `mapstructure:"libretranslate-url"`
	Google            auth.GoogleOauthConfig `mapstructure:"google"`
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
