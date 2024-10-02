package config

type SaltOptions struct {
	Auth        string `mapstructure:"auth" yaml:"auth"`
	URL         string `mapstructure:"url" yaml:"url"`
	ExternalURL string `mapstructure:"external_url" yaml:"external_url"`
	Insecure    bool   `mapstructure:"insecure" yaml:"insecure"`
}
