package config

import (
	"bytes"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const envVarPrefix = "agartha"

type Config struct {
	HTTP HTTPOptions `mapstructure:"http" yaml:"http"`
	LDAP LDAPOptions `mapstructure:"ldap" yaml:"ldap"`
	SAML SAMLOptions `mapstructure:"saml" yaml:"saml"`
	DB   DBOptions   `mapstructure:"db" yaml:"db"`
	Salt SaltOptions `mapstructure:"salt" yaml:"salt"`
	CAS  CASOptions  `mapstructure:"cas" yaml:"cas"`
}

func NewConfig() *Config {
	secret, err := generateRandomString(32)
	if err != nil {
		secret = "secret"
	}
	// return a new instance of Config with default values
	return &Config{
		HTTP: HTTPOptions{
			Host:   "",
			Port:   8080,
			Secret: secret,
		},
		LDAP: LDAPOptions{
			Server:            "ldap.example.com",
			User:              "user",
			Password:          "password",
			BaseDN:            "dc=example,dc=com",
			Filter:            "(objectClass=*)",
			LDAPDomainDefault: "example.com",
		},
		DB: DBOptions{
			Host:     "localhost",
			DBName:   "agartha",
			User:     "agartha",
			Password: "password",
			Port:     5432,
			SSLMode:  "disable",
			Type:     "postgres",
			Tables: SaltDBTables{
				JIDs:        "jids",
				SaltCache:   "salt_cache",
				SaltReturns: "salt_returns",
				SaltEvents:  "salt_events",
				UseJSONB:    false,
			},
		},
		Salt: SaltOptions{
			URL:         "",
			ExternalURL: "",
			Auth:        "agartha",
			Insecure:    false,
		},
		SAML: SAMLOptions{
			MetadataURL: "",
			SessionCert: "",
			ServerCert:  "",
			ServerURL:   "",
			EntityID:    "",
			SessionKey:  "",
		},
		CAS: CASOptions{
			Server:       "https://cas.example.com",
			ServiceURL:   "http://agartha.example.com/cas",
			ValidatePath: "/serviceValidate",
			LoginPath:    "/login",
			LogoutPath:   "/logout",
		},
	}
}

var AgarthaConfig *Config

func InitConfig() error {
	v := viper.New()

	b, err := yaml.Marshal(NewConfig())
	if err != nil {
		return err
	}
	defaultConfig := bytes.NewReader(b)
	v.SetConfigType("yaml")
	if err := v.MergeConfig(defaultConfig); err != nil {
		return err
	}

	v.SetConfigName("config")          // name of config file (without extension)
	v.AddConfigPath("/etc/agartha/")   // path to look for the config file in
	v.AddConfigPath("$HOME/.agartha/") // call multiple times to add many search paths
	v.AddConfigPath(".")               // optionally look for config in the working directory

	if err := v.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	v.AutomaticEnv()
	v.SetEnvPrefix(envVarPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	AgarthaConfig = &Config{}
	return v.Unmarshal(AgarthaConfig)
}
