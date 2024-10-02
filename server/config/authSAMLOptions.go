package config

type SAMLOptions struct {
	MetadataURL string `mapstructure:"metadata_url" yaml:"metadata_url"`
	SessionCert string `mapstructure:"session_certificate" yaml:"session_certificate"`
	SessionKey  string `mapstructure:"session_key" yaml:"session_key"`
	ServerCert  string `mapstructure:"server_certificate" yaml:"server_certificate"`
	ServerURL   string `mapstructure:"server_url" yaml:"server_url"`
	EntityID    string `mapstructure:"entity_id" yaml:"entity_id"`
}
