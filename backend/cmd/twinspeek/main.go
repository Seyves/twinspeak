package main

import (
	"context"
	"flag"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twinspeak/backend/internal/auth"
	"github.com/twinspeak/backend/internal/billing"
	"github.com/twinspeak/backend/internal/config"
	"github.com/twinspeak/backend/internal/db"
	"github.com/twinspeak/backend/internal/email"
	"github.com/twinspeak/backend/internal/googleauth"
	"github.com/twinspeak/backend/internal/metrics"
	"github.com/twinspeak/backend/internal/preferences"
	"github.com/twinspeak/backend/internal/server"
	"github.com/twinspeak/backend/internal/speechpipeline"
	"github.com/twinspeak/backend/internal/users"
)

func main() {
	cfgPath := flag.String("c", "", "Path to configuration file (optional, falls back to env vars)")
	flag.Parse()

	var cfg config.Config
	err := config.Parse(*cfgPath, &cfg)
	if err != nil {
		log.Errorf("Parsing config: %s", err.Error())
		return
	}

	pool, err := pgxpool.New(context.Background(), cfg.DBUrl)
	if err != nil {
		log.Errorf("Connecting to DB: %s", err.Error())
		return
	}
	defer pool.Close()
	queries := db.New(pool)

	authm := auth.New(cfg.HMACSecret)
	googleauthm := googleauth.New(cfg.Google)
	billing := billing.New()

	emailm, err := email.New(cfg.Resend.ApiKey, cfg.Resend.FromEmail, cfg.PublicUrl)
	if err != nil {
		log.Errorf("Creating email module: %s", err.Error())
		return
	}

	metricss := metrics.New(pool, queries)
	preferencesm := &preferences.Module{}
	userss := users.New(pool, queries, authm, googleauthm, billing, emailm, preferencesm, metricss)

	var p speechpipeline.Pipeline
	switch cfg.Pipeline {
	case "gladia":
		p, err = speechpipeline.NewGladiaPipeline(cfg.GladiaKey)
		if err != nil {
			log.Errorf("Creating Gladia pipeline: %s", err.Error())
			return
		}
	case "whisper":
		p, err = speechpipeline.NewWhisperPipeline(cfg.FasterWhisperUrl, cfg.LibretranslateUrl)
		if err != nil {
			log.Errorf("Creating Whisper pipeline: %s", err.Error())
			return
		}
	}

	api := server.NewRestApi(cfg.Host, p, metricss, userss, emailm, pool, queries)
	err = api.Start()
	if err != nil {
		log.Errorf("Starting server: %s", err.Error())
		return
	}
}
