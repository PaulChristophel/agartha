package config

import "time"

type DBOptions struct {
	Host                string        `mapstructure:"host" yaml:"host"`
	DBName              string        `mapstructure:"db_name" yaml:"db_name"`
	User                string        `mapstructure:"user" yaml:"user"`
	Password            string        `mapstructure:"password" yaml:"password"`
	Port                int           `mapstructure:"port" yaml:"port"`
	SSLMode             string        `mapstructure:"sslmode" yaml:"sslmode"`
	Type                string        `mapstructure:"type" yaml:"type"`
	Tables              SaltDBTables  `mapstructure:"tables" yaml:"tables"`
	RetryAttempts       int           `mapstructure:"retry_attempts" yaml:"retry_attempts"`
	RetryInitialBackoff time.Duration `mapstructure:"retry_initial_backoff" yaml:"retry_initial_backoff"`
	RetryMaxBackoff     time.Duration `mapstructure:"retry_max_backoff" yaml:"retry_max_backoff"`
	RetryMultiplier     float64       `mapstructure:"retry_multiplier" yaml:"retry_multiplier"`
}

type SaltDBTables struct {
	JIDs        string `mapstructure:"jids" yaml:"jids"`
	SaltCache   string `mapstructure:"salt_cache" yaml:"salt_cache"`
	SaltKeys    string `mapstructure:"salt_keys" yaml:"salt_keys"`
	SaltReturns string `mapstructure:"salt_returns" yaml:"salt_returns"`
	SaltEvents  string `mapstructure:"salt_events" yaml:"salt_events"`
	UseJSONB    bool   `mapstructure:"use_jsonb" yaml:"use_jsonb"`
}
