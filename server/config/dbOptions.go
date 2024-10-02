package config

type DBOptions struct {
	Host     string       `mapstructure:"host" yaml:"host"`
	DBName   string       `mapstructure:"db_name" yaml:"db_name"`
	User     string       `mapstructure:"user" yaml:"user"`
	Password string       `mapstructure:"password" yaml:"password"`
	Port     int          `mapstructure:"port" yaml:"port"`
	SSLMode  string       `mapstructure:"sslmode" yaml:"sslmode"`
	Type     string       `mapstructure:"type" yaml:"type"`
	Tables   SaltDBTables `mapstructure:"tables" yaml:"tables"`
}

type SaltDBTables struct {
	JIDs        string `mapstructure:"jids" yaml:"jids"`
	SaltCache   string `mapstructure:"salt_cache" yaml:"salt_cache"`
	SaltReturns string `mapstructure:"salt_returns" yaml:"salt_returns"`
	SaltEvents  string `mapstructure:"salt_events" yaml:"salt_events"`
}
