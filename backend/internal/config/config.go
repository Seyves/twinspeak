package config

import (
	"fmt"
	"os"

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
	Google            googleauth.Config `mapstructure:"google"`
	Resend            ResendConfig      `mapstructure:"resend"`
}

func Parse(path string, cfg any) error {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error unmarshalling config: %s", err.Error())
	}
	return nil
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
