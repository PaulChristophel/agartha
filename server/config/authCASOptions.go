package config

type CASOptions struct {
	Server       string `mapstructure:"server" yaml:"server"`
	ServiceURL   string `mapstructure:"service_url" yaml:"service_url"`
	ValidatePath string `mapstructure:"validate_path" yaml:"validate_path"`
	LoginPath    string `mapstructure:"login_path" yaml:"login_path"`
	LogoutPath   string `mapstructure:"logout_path" yaml:"logout_path"`
}
