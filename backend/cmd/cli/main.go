package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/twinspeak/backend/internal/auth"
	"github.com/twinspeak/backend/internal/billing"
	"github.com/twinspeak/backend/internal/config"
	"github.com/twinspeak/backend/internal/db"
	"github.com/twinspeak/backend/internal/email"
	"github.com/twinspeak/backend/internal/googleauth"
	"github.com/twinspeak/backend/internal/metrics"
	"github.com/twinspeak/backend/internal/preferences"
	"github.com/twinspeak/backend/internal/service"
)

type deps struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	service *service.Service
}

func initDeps(cfgPath string) (*deps, error) {
	var cfg config.Config
	err := config.Parse(config.ResolveConfigPath(cfgPath), &cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to db: %w", err)
	}

	queries := db.New(pool)
	authModule := auth.New(cfg.HMACSecret)
	googleauthModule := googleauth.New(cfg.Google)
	billingModule := billing.New()
	emailModule, err := email.New(cfg.Resend.ApiKey, cfg.Resend.FromEmail, cfg.PublicUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot create email module: %w", err)
	}
	metricsModule := metrics.New(pool, queries)
	preferencesModule := preferences.New()
	mainService := service.New(pool, queries, authModule, googleauthModule, billingModule, emailModule, preferencesModule, metricsModule)

	return &deps{
		pool:    pool,
		queries: queries,
		service: mainService,
	}, nil
}

var cfgPath string

func main() {
	rootCmd := &cobra.Command{
		Use:   "twinspeak-cli",
		Short: "Twinspeak helper CLI",
	}

	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "", "config file path (default: $TWINSPEAK_CONFIG or /etc/twinspeak/config.yaml)")

	seedCmd := &cobra.Command{
		Use:   "seed",
		Short: "Seed database with predefined templates",
	}
	seedCmd.AddCommand(seedUserCmd())

	rootCmd.AddCommand(seedCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// --- seed templates ---

func seedUserCmd() *cobra.Command {
	var (
		activeTopups, expiredTopups []int
		usedSub                     bool
	)

	cmd := &cobra.Command{
		Use:   "user <email> <password>",
		Short: "Seed a user",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			d, err := initDeps(cfgPath)
			if err != nil {
				return err
			}
			defer d.pool.Close()

			email, password := args[0], args[1]
			now := time.Now()

			accessToken, _, err := d.service.SignUp(ctx, now, email, password)
			if err != nil {
				return err
			}

			userId, err := d.service.ValidateAccessToken(ctx, now, accessToken.Value)
			if err != nil {
				return err
			}

			// Manually verifying email
			err = d.queries.VerifyUserEmail(ctx, userId)
			if err != nil {
				return err
			}

			if usedSub {
				// mock speeches until subscription credits runs out
				for range 1000 {
					err := d.service.StartSpeech(ctx, now, userId)
					if err != nil {
						if errors.Is(err, billing.ErrInsufficientCredits) {
							break
						}
						return err
					}
					err = d.service.EndSpeech(context.Background(), now, db.InsertSpeechParams{
						UserID:        userId,
						InLang:        "en",
						OutLang:       "fr",
						Transcription: "Hey how are you doing?",
						Translation:   "Salut, comment vas-tu?",
						ChatSide:      db.ChatSideBottom,
						StartedAt:     now,
						EndedAt:       now,
					})
					if err != nil {
						return err
					}
				}
			}

			for _, amount := range activeTopups {
				if err := d.service.BuyTopup(ctx, now, userId, int32(amount)); err != nil {
					return err
				}
			}

			// expired: pass now-2months so expiry lands 1 month in the past
			expiredNow := now.AddDate(0, -2, 0)
			for _, amount := range expiredTopups {
				if err := d.service.BuyTopup(ctx, expiredNow, userId, int32(amount)); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().IntSliceVar(&activeTopups, "active-topup", nil, "add active topup with given amount (repeatable)")
	cmd.Flags().IntSliceVar(&expiredTopups, "expired-topup", nil, "add expired topup with given amount (repeatable)")
	cmd.Flags().BoolVar(&usedSub, "used-sub", false, "make subscription used")

	return cmd
}
