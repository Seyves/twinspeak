package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/twinspeak/backend/internal/googleauth"
)

const defaultPath = "/etc/twinspeak/config.yaml"

type ResendConfig struct {
	ApiKey    string `mapstructure:"api-key"`
	FromEmail string `mapstructure:"from-email"`
}

type Config struct {
	Host              string            `mapstructure:"host"`
	PublicUrl         string            `mapstructure:"public-url"` // for link generation
	HMACSecret        string            `mapstructure:"hmac-secret"`
	DBUrl             string            `mapstructure:"db-url"`
	Pipeline          string            `mapstructure:"pipeline"`
	GladiaKey         string            `mapstructure:"gladia-key"`
	FasterWhisperUrl  string            `mapstructure:"faster-whisper-url"`
	LibretranslateUrl string            `mapstructure:"libretranslate-url"`
	SchedulerInterval time.Duration     `mapstructure:"scheduler-interval"` // interval for subscription renewal scheduler
	Google            googleauth.Config `mapstructure:"google"`
	Resend            ResendConfig      `mapstructure:"resend"`
}

func Parse(path string, cfg any) error {
	viper.SetEnvPrefix("TWINSPEAK")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("scheduler-interval", "1h")

	bindEnvVars()

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			viper.SetConfigFile(path)
			if err := viper.ReadInConfig(); err != nil {
				return fmt.Errorf("error reading config file: %s", err.Error())
			}
		}
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error unmarshalling config: %s", err.Error())
	}

	return nil
}

func bindEnvVars() {
	viper.BindEnv("host", "TWINSPEAK_HOST")
	viper.BindEnv("public-url", "TWINSPEAK_PUBLIC_URL")
	viper.BindEnv("hmac-secret", "TWINSPEAK_HMAC_SECRET")
	viper.BindEnv("db-url", "TWINSPEAK_DB_URL")
	viper.BindEnv("pipeline", "TWINSPEAK_PIPELINE")
	viper.BindEnv("gladia-key", "TWINSPEAK_GLADIA_KEY")
	viper.BindEnv("faster-whisper-url", "TWINSPEAK_FASTER_WHISPER_URL")
	viper.BindEnv("libretranslate-url", "TWINSPEAK_LIBRETRANSLATE_URL")
	viper.BindEnv("scheduler-interval", "TWINSPEAK_SCHEDULER_INTERVAL")

	viper.BindEnv("google.client-id", "TWINSPEAK_GOOGLE_CLIENT_ID")
	viper.BindEnv("google.client-secret", "TWINSPEAK_GOOGLE_CLIENT_SECRET")
	viper.BindEnv("google.redirect-url", "TWINSPEAK_GOOGLE_REDIRECT_URL")

	viper.BindEnv("resend.api-key", "TWINSPEAK_RESEND_API_KEY")
	viper.BindEnv("resend.from-email", "TWINSPEAK_RESEND_FROM_EMAIL")
}

func ResolveConfigPath(flag string) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv("TWINSPEAK_CONFIG"); env != "" {
		return env
	}
	return defaultPath
}
