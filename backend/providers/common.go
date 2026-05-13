package providers

type ProviderConfig struct {
	Provider string `mapstructure:"provider"`
	Url      string `mapstructure:"url"`
}
