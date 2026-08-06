package auth

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/PaulChristophel/agartha/server/config"
	"github.com/gin-contrib/sessions"
	"github.com/go-ldap/ldap/v3"
)

type ldapConnection interface {
	Bind(username, password string) error
	Close() error
	Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error)
	StartTLS(config *tls.Config) error
}

var jwtSecret []byte
var session sessions.Session
var ldapOptions config.LDAPOptions
var casOptions config.CASOptions
var enabledMethods = map[string]struct{}{}
var casHTTPClient = &http.Client{Timeout: 10 * time.Second}
var ldapDialURL = func(server string) (ldapConnection, error) {
	return ldap.DialURL(server)
}

func SetOptions(secret []byte, methods []string, ldap config.LDAPOptions, cas config.CASOptions) {
	jwtSecret = secret
	ldapOptions = ldap
	casOptions = cas
	enabledMethods = make(map[string]struct{}, len(methods))
	for _, method := range methods {
		enabledMethods[method] = struct{}{}
	}
}
