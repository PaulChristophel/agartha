package config

type AuthOptions struct {
	Methods []string `mapstructure:"methods" yaml:"methods"`
}
