package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewConfigUsesSecureSessionCookiesByDefault(t *testing.T) {
	require.True(t, NewConfig().HTTP.CookieSecure)
}

func validConfig() Config {
	config := *NewConfig()
	config.HTTP.Secret = "a-unique-secret-with-at-least-32-characters"
	config.Auth.Methods = []string{"local"}
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

func TestEffectiveAuthMethodsRequiresExplicitProvider(t *testing.T) {
	config := validConfig()
	config.Auth.Methods = nil

	_, err := config.EffectiveAuthMethods()
	require.ErrorContains(t, err, "must enable at least one authentication method")
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
	originalConfig := SylviteConfig
	t.Cleanup(func() { SylviteConfig = originalConfig })
	t.Setenv("SYLVITE_HTTP_READ_TIMEOUT", "17s")
	t.Setenv("SYLVITE_AUTH_METHODS", "local,ldap")

	require.NoError(t, InitConfig())
	require.Equal(t, 17*time.Second, SylviteConfig.HTTP.ReadTimeout)
	require.Equal(t, []string{"local", "ldap"}, SylviteConfig.Auth.Methods)
}

func TestLegacyEnvironmentOverridesConfigurationFile(t *testing.T) {
	original := SylviteConfig
	t.Cleanup(func() { SylviteConfig = original })
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("http:\n  read_timeout: 3s\nsalt:\n  auth: sylvite\n"), 0600))
	t.Setenv("SYLVITE_HTTP_READ_TIMEOUT", "")
	t.Setenv("SYLVITE_AUTH_METHODS", "")
	t.Setenv("SYLVITE_HTTP_COOKIE_SECURE", "")
	t.Setenv("SYLVITE_SALT_AUTH", "")
	t.Setenv("AGARTHA_HTTP_READ_TIMEOUT", "23s")
	t.Setenv("AGARTHA_AUTH_METHODS", "local,ldap")
	t.Setenv("AGARTHA_HTTP_COOKIE_SECURE", "false")
	t.Setenv("AGARTHA_SALT_AUTH", "agartha")

	require.NoError(t, initConfig([]string{dir}))
	require.Equal(t, 23*time.Second, SylviteConfig.HTTP.ReadTimeout)
	require.Equal(t, []string{"local", "ldap"}, SylviteConfig.Auth.Methods)
	require.False(t, SylviteConfig.HTTP.CookieSecure)
	require.Equal(t, "agartha", SylviteConfig.Salt.Auth)
}

func TestSylviteEnvironmentTakesPrecedenceOverLegacyEnvironment(t *testing.T) {
	original := SylviteConfig
	t.Cleanup(func() { SylviteConfig = original })
	t.Setenv("AGARTHA_HTTP_READ_TIMEOUT", "invalid-duration")
	t.Setenv("SYLVITE_HTTP_READ_TIMEOUT", "31s")
	t.Setenv("AGARTHA_HTTP_COOKIE_SECURE", "true")
	t.Setenv("SYLVITE_HTTP_COOKIE_SECURE", "false")

	require.NoError(t, initConfig([]string{t.TempDir()}))
	require.Equal(t, 31*time.Second, SylviteConfig.HTTP.ReadTimeout)
	require.False(t, SylviteConfig.HTTP.CookieSecure)
}

func TestConfigPathFallbackAndPrecedence(t *testing.T) {
	original := SylviteConfig
	t.Cleanup(func() { SylviteConfig = original })
	t.Setenv("SYLVITE_HTTP_PORT", "")
	t.Setenv("AGARTHA_HTTP_PORT", "")
	root := t.TempDir()
	t.Setenv("HOME", root)
	modern := filepath.Join(root, ".sylvite")
	legacy := filepath.Join(root, ".agartha")
	require.NoError(t, os.MkdirAll(legacy, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(legacy, "config.yaml"), []byte("http:\n  port: 8123\n"), 0600))
	paths := []string{"$HOME/.sylvite", "$HOME/.agartha"}
	require.NoError(t, initConfig(paths))
	require.EqualValues(t, 8123, SylviteConfig.HTTP.Port)

	require.NoError(t, os.MkdirAll(modern, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(modern, "config.yaml"), []byte("http:\n  port: 8456\n"), 0600))
	require.NoError(t, initConfig(paths))
	require.EqualValues(t, 8456, SylviteConfig.HTTP.Port)

	require.NoError(t, os.WriteFile(filepath.Join(modern, "config.yaml"), []byte("http: [invalid"), 0600))
	require.Error(t, initConfig(paths))
}

func TestLegacySaltPluginMatchesSylviteImplementation(t *testing.T) {
	modern, err := os.ReadFile("../../extras/auth/sylvite.py")
	require.NoError(t, err)
	legacy, err := os.ReadFile("../../extras/auth/agartha.py")
	require.NoError(t, err)
	// Both files must work when deployed alone, with Salt injecting globals into
	// each module independently. Keep the legacy copy identical except its name.
	require.Equal(t, strings.Replace(string(modern), `__virtualname__ = "sylvite"`, `__virtualname__ = "agartha"`, 1), string(legacy))
}
