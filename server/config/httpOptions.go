package config

import "time"

type HTTPOptions struct {
	Address           string `mapstructure:"address" yaml:"address"`
	Host              string `mapstructure:"host" yaml:"host"`
	Port              int    `mapstructure:"port" yaml:"port"`
	Secret            string `mapstructure:"secret" yaml:"secret"`
	ForgotPasswordURL string `mapstructure:"forgot_password_url" yaml:"forgot_password_url"`
	GetStartedURL     string `mapstructure:"get_started_url" yaml:"get_started_url"`

	TLSCertFile     string `mapstructure:"tls_cert_file" yaml:"tls_cert_file"`
	TLSKeyFile      string `mapstructure:"tls_key_file" yaml:"tls_key_file"`
	TLSClientCAFile string `mapstructure:"tls_client_ca_file" yaml:"tls_client_ca_file"`

	TrustedProxies    []string      `mapstructure:"trusted_proxies" yaml:"trusted_proxies"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout" yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout" yaml:"idle_timeout"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown_timeout" yaml:"shutdown_timeout"`
}
