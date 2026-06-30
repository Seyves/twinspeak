package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	"github.com/twinspeak/backend/internal/scheduler"
	"github.com/twinspeak/backend/internal/server"
	"github.com/twinspeak/backend/internal/service"
	"github.com/twinspeak/backend/internal/speechpipeline"
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

	authModule := auth.New(cfg.HMACSecret)
	googleauthModule := googleauth.New(cfg.Google)
	billingModule := billing.New()

	emailModule, err := email.New(cfg.Resend.ApiKey, cfg.Resend.FromEmail, cfg.PublicUrl)
	if err != nil {
		log.Errorf("Creating email module: %s", err.Error())
		return
	}

	metricsModule := metrics.New(pool, queries)
	preferencesModule := preferences.New()
	mainService := service.New(pool, queries, authModule, googleauthModule, billingModule, emailModule, preferencesModule, metricsModule)

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

	sched := scheduler.New(mainService, cfg.SchedulerInterval)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched.Start(ctx, true)
	log.Infof("Subscription renewal scheduler started with interval: %s", cfg.SchedulerInterval)

	api := server.NewRestApi(cfg.Host, p, mainService, pool, queries)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- api.Start()
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-serverErrors:
		log.Errorf("Server error: %s", err.Error())
	case sig := <-signalChan:
		log.Infof("Received signal %v, initiating graceful shutdown", sig)
	}

	log.Info("Shutting down gracefully...")

	if err := sched.Stop(30 * time.Second); err != nil {
		log.Warnf("Scheduler shutdown warning: %s", err.Error())
	}

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := api.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server shutdown error: %s", err.Error())
	}

	pool.Close()

	log.Info("Shutdown complete")
}
