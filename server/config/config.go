package config

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const envVarPrefix = "agartha"

type Config struct {
	HTTP HTTPOptions `mapstructure:"http" yaml:"http"`
	Auth AuthOptions `mapstructure:"auth" yaml:"auth"`
	LDAP LDAPOptions `mapstructure:"ldap" yaml:"ldap"`
	SAML SAMLOptions `mapstructure:"saml" yaml:"saml"`
	DB   DBOptions   `mapstructure:"db" yaml:"db"`
	Salt SaltOptions `mapstructure:"salt" yaml:"salt"`
	CAS  CASOptions  `mapstructure:"cas" yaml:"cas"`
}

func NewConfig() *Config {
	// return a new instance of Config with default values
	return &Config{
		HTTP: HTTPOptions{
			Host:              "",
			Port:              8080,
			ReadTimeout:       15 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
			ShutdownTimeout:   15 * time.Second,
		},
		LDAP: LDAPOptions{
			Server:            "ldap.example.com",
			User:              "user",
			Password:          "password",
			BaseDN:            "dc=example,dc=com",
			Filter:            "(objectClass=*)",
			LDAPDomainDefault: "example.com",
			StartTLS:          false,
		},
		DB: DBOptions{
			Host:                "localhost",
			DBName:              "agartha",
			User:                "agartha",
			Password:            "password",
			Port:                5432,
			SSLMode:             "disable",
			Type:                "postgres",
			RetryAttempts:       5,
			RetryInitialBackoff: time.Second,
			RetryMaxBackoff:     10 * time.Second,
			RetryMultiplier:     2,
			Tables: SaltDBTables{
				JIDs:        "jids",
				SaltCache:   "salt_cache",
				SaltKeys:    "salt_keys",
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

// ValidateForServe rejects unsafe placeholders and incomplete server/authentication
// settings before opening network listeners or database connections.
func (c Config) ValidateForServe() error {
	var errs []error

	if secret := strings.TrimSpace(c.HTTP.Secret); isPlaceholder(secret) || len(secret) < 32 {
		errs = append(errs, errors.New("http.secret must be a non-placeholder secret of at least 32 characters"))
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		errs = append(errs, errors.New("http.port must be between 1 and 65535"))
	}
	for name, value := range map[string]time.Duration{
		"http.read_timeout":        c.HTTP.ReadTimeout,
		"http.read_header_timeout": c.HTTP.ReadHeaderTimeout,
		"http.write_timeout":       c.HTTP.WriteTimeout,
		"http.idle_timeout":        c.HTTP.IdleTimeout,
		"http.shutdown_timeout":    c.HTTP.ShutdownTimeout,
	} {
		if value <= 0 {
			errs = append(errs, fmt.Errorf("%s must be greater than zero", name))
		}
	}
	if (c.HTTP.TLSCertFile == "") != (c.HTTP.TLSKeyFile == "") {
		errs = append(errs, errors.New("http.tls_cert_file and http.tls_key_file must be configured together"))
	}
	if c.HTTP.TLSClientCAFile != "" && c.HTTP.TLSCertFile == "" {
		errs = append(errs, errors.New("http.tls_client_ca_file requires server TLS configuration"))
	}
	for _, proxy := range c.HTTP.TrustedProxies {
		if net.ParseIP(proxy) != nil {
			continue
		}
		if _, _, err := net.ParseCIDR(proxy); err != nil {
			errs = append(errs, fmt.Errorf("http.trusted_proxies contains invalid IP address or CIDR %q", proxy))
		}
	}

	if err := c.ValidateDatabase(); err != nil {
		errs = append(errs, err)
	}
	methods, err := c.EffectiveAuthMethods()
	if err != nil {
		errs = append(errs, err)
	}
	if contains(methods, "ldap") {
		if err := validateLDAP(c.LDAP); err != nil {
			errs = append(errs, err)
		}
	}
	if contains(methods, "cas") {
		if err := validateCAS(c.CAS); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// EffectiveAuthMethods returns the configured authentication allowlist. When
// auth.methods is omitted, it derives the legacy provider list for compatibility.
func (c Config) EffectiveAuthMethods() ([]string, error) {
	methods := c.Auth.Methods
	if len(methods) == 0 {
		methods = []string{"local"}
		if ldapConfigured(c.LDAP) {
			methods = append(methods, "ldap")
		}
		if casConfigured(c.CAS) {
			methods = append(methods, "cas")
		}
	}

	seen := make(map[string]struct{}, len(methods))
	normalized := make([]string, 0, len(methods))
	for _, method := range methods {
		method = strings.ToLower(strings.TrimSpace(method))
		switch method {
		case "local", "ldap", "cas":
		default:
			return nil, fmt.Errorf("auth.methods contains unsupported method %q", method)
		}
		if _, exists := seen[method]; exists {
			continue
		}
		seen[method] = struct{}{}
		normalized = append(normalized, method)
	}
	if len(normalized) == 0 {
		return nil, errors.New("auth.methods must enable at least one authentication method")
	}
	return normalized, nil
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

// ValidateDatabase validates settings shared by the server and migration command.
func (c Config) ValidateDatabase() error {
	var errs []error
	for name, value := range map[string]string{
		"db.host":    c.DB.Host,
		"db.db_name": c.DB.DBName,
		"db.user":    c.DB.User,
	} {
		if strings.TrimSpace(value) == "" {
			errs = append(errs, fmt.Errorf("%s is required", name))
		}
	}
	if isPlaceholder(c.DB.Password) {
		errs = append(errs, errors.New("db.password must not be empty or a known placeholder"))
	}
	if c.DB.Port < 1 || c.DB.Port > 65535 {
		errs = append(errs, errors.New("db.port must be between 1 and 65535"))
	}
	if c.DB.RetryAttempts < 1 {
		errs = append(errs, errors.New("db.retry_attempts must be at least 1"))
	}
	if c.DB.RetryInitialBackoff <= 0 {
		errs = append(errs, errors.New("db.retry_initial_backoff must be greater than zero"))
	}
	if c.DB.RetryMaxBackoff < c.DB.RetryInitialBackoff {
		errs = append(errs, errors.New("db.retry_max_backoff must be at least db.retry_initial_backoff"))
	}
	if c.DB.RetryMultiplier < 1 {
		errs = append(errs, errors.New("db.retry_multiplier must be at least 1"))
	}
	return errors.Join(errs...)
}

func ldapConfigured(options LDAPOptions) bool {
	return strings.TrimSpace(options.Server) != "" && !strings.EqualFold(strings.TrimSpace(options.Server), "ldap.example.com")
}

func validateLDAP(options LDAPOptions) error {
	var errs []error
	parsed, err := url.Parse(options.Server)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "ldap" && parsed.Scheme != "ldaps") {
		errs = append(errs, errors.New("ldap.server must be an absolute ldap:// or ldaps:// URL"))
	} else if isExampleHost(parsed.Hostname()) {
		errs = append(errs, errors.New("ldap.server must not use an example.com placeholder host"))
	}
	for name, value := range map[string]string{
		"ldap.user":                options.User,
		"ldap.base_dn":             options.BaseDN,
		"ldap.filter":              options.Filter,
		"ldap.ldap_domain_default": options.LDAPDomainDefault,
	} {
		if strings.TrimSpace(value) == "" || isPlaceholder(value) {
			errs = append(errs, fmt.Errorf("%s must be configured when LDAP authentication is enabled", name))
		}
	}
	if isPlaceholder(options.Password) {
		errs = append(errs, errors.New("ldap.password must not be empty or a known placeholder"))
	}
	if !strings.Contains(options.Filter, "%s") {
		errs = append(errs, errors.New("ldap.filter must contain a %s username placeholder"))
	}
	if options.StartTLS && parsed != nil && parsed.Scheme == "ldaps" {
		errs = append(errs, errors.New("ldap.start_tls requires an ldap:// server URL, not ldaps://"))
	}
	return errors.Join(errs...)
}

func casConfigured(options CASOptions) bool {
	return strings.TrimSpace(options.Server) != "" && !strings.EqualFold(strings.TrimSpace(options.Server), "https://cas.example.com")
}

func validateCAS(options CASOptions) error {
	var errs []error
	for name, value := range map[string]string{
		"cas.server":        options.Server,
		"cas.service_url":   options.ServiceURL,
		"cas.validate_path": options.ValidatePath,
		"cas.login_path":    options.LoginPath,
		"cas.logout_path":   options.LogoutPath,
	} {
		if strings.TrimSpace(value) == "" {
			errs = append(errs, fmt.Errorf("%s must be configured when CAS authentication is enabled", name))
		}
	}
	for name, value := range map[string]string{"cas.server": options.Server, "cas.service_url": options.ServiceURL} {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			errs = append(errs, fmt.Errorf("%s must be an absolute URL", name))
		} else if isExampleHost(parsed.Hostname()) {
			errs = append(errs, fmt.Errorf("%s must not use an example.com placeholder host", name))
		}
	}
	return errors.Join(errs...)
}

func isPlaceholder(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" || strings.Contains(normalized, "replace_with") || strings.Contains(normalized, "replace-with") || strings.Contains(normalized, "example.com") || strings.Contains(normalized, "dc=example,") {
		return true
	}
	switch normalized {
	case "password", "foobar", "secret", "mysecret", "changeme", "change-me", "change_me", "example", "user":
		return true
	default:
		return false
	}
}

func isExampleHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	return host == "example.com" || strings.HasSuffix(host, ".example.com")
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

	if startTLS := v.Get("ldap.start_tls"); startTLS == nil {
		v.Set("ldap.start_tls", false)
	} else if startTLSString, ok := startTLS.(string); ok && strings.TrimSpace(startTLSString) == "" {
		v.Set("ldap.start_tls", false)
	}

	AgarthaConfig = &Config{}
	return v.Unmarshal(AgarthaConfig)
}
