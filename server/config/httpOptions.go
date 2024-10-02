package config

type HTTPOptions struct {
	Address           string `mapstructure:"address" yaml:"address"`
	Host              string `mapstructure:"host" yaml:"host"`
	Port              int    `mapstructure:"port" yaml:"port"`
	Secret            string `mapstructure:"secret" yaml:"secret"`
	ForgotPasswordURL string `mapstructure:"forgot_password_url" yaml:"forgot_password_url"`
	GetStartedURL     string `mapstructure:"get_started_url" yaml:"get_started_url"`
}
