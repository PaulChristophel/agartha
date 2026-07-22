package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func validConfig() Config {
	config := *NewConfig()
	config.HTTP.Secret = "a-unique-secret-with-at-least-32-characters"
	config.DB.Password = "a-unique-database-password"
	return config
}

func TestValidateForServeAcceptsCompleteConfiguration(t *testing.T) {
	config := validConfig()
	require.NoError(t, config.ValidateForServe())
}

func TestValidateForServeRejectsUnsafeCredentialsAndTimeouts(t *testing.T) {
	config := validConfig()
	config.HTTP.Secret = "mysecret"
	config.DB.Password = "password"
	config.HTTP.ReadTimeout = 0

	err := config.ValidateForServe()
	require.ErrorContains(t, err, "http.secret")
	require.ErrorContains(t, err, "db.password")
	require.ErrorContains(t, err, "http.read_timeout")
}

func TestValidateForServeRejectsIncompleteLDAPConfiguration(t *testing.T) {
	config := validConfig()
	config.Auth.Methods = []string{"local", "ldap"}
	config.LDAP.Server = "ldaps://directory.example.test:636"
	config.LDAP.Password = "password"
	config.LDAP.BaseDN = ""

	err := config.ValidateForServe()
	require.ErrorContains(t, err, "ldap.password")
	require.ErrorContains(t, err, "ldap.base_dn")
}

func TestValidateForServeRejectsIncompleteCASConfiguration(t *testing.T) {
	config := validConfig()
	config.Auth.Methods = []string{"cas"}
	config.CAS.Server = "https://login.example.test/cas"
	config.CAS.ServiceURL = ""

	err := config.ValidateForServe()
	require.ErrorContains(t, err, "cas.service_url")
}

func TestValidateForServeRejectsExampleAuthenticationEndpoints(t *testing.T) {
	config := validConfig()
	config.Auth.Methods = []string{"ldap", "cas"}
	config.LDAP.Server = "ldaps://ldap.example.com"
	config.LDAP.Password = "a-real-looking-service-password"
	config.LDAP.Filter = "(sAMAccountName=%s)"

	err := config.ValidateForServe()
	require.ErrorContains(t, err, "ldap.server must not use an example.com placeholder host")
	require.ErrorContains(t, err, "cas.server must not use an example.com placeholder host")
}

func TestEffectiveAuthMethodsRejectsUnsupportedProvider(t *testing.T) {
	config := validConfig()
	config.Auth.Methods = []string{"local", "oidc"}

	_, err := config.EffectiveAuthMethods()
	require.ErrorContains(t, err, "unsupported method")
}

func TestValidateDatabaseRejectsInvalidRetryPolicy(t *testing.T) {
	config := validConfig()
	config.DB.RetryAttempts = 0
	config.DB.RetryInitialBackoff = 2 * time.Second
	config.DB.RetryMaxBackoff = time.Second
	config.DB.RetryMultiplier = 0.5

	err := config.ValidateDatabase()
	require.ErrorContains(t, err, "db.retry_attempts")
	require.ErrorContains(t, err, "db.retry_max_backoff")
	require.ErrorContains(t, err, "db.retry_multiplier")
}

func TestValidateForServeRequiresCompleteTLSConfiguration(t *testing.T) {
	config := validConfig()
	config.HTTP.TLSCertFile = "/cert.pem"

	require.ErrorContains(t, config.ValidateForServe(), "tls_cert_file and http.tls_key_file")
}

func TestValidateForServeRejectsInvalidTrustedProxy(t *testing.T) {
	config := validConfig()
	config.HTTP.TrustedProxies = []string{"10.0.0.0/8", "not-a-network"}

	require.ErrorContains(t, config.ValidateForServe(), "http.trusted_proxies")
}

func TestInitConfigDecodesDurationAndAuthMethodEnvironmentValues(t *testing.T) {
	originalConfig := AgarthaConfig
	t.Cleanup(func() { AgarthaConfig = originalConfig })
	t.Setenv("AGARTHA_HTTP_READ_TIMEOUT", "17s")
	t.Setenv("AGARTHA_AUTH_METHODS", "local,ldap")

	require.NoError(t, InitConfig())
	require.Equal(t, 17*time.Second, AgarthaConfig.HTTP.ReadTimeout)
	require.Equal(t, []string{"local", "ldap"}, AgarthaConfig.Auth.Methods)
}
