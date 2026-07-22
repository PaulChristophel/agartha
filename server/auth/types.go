package auth

import (
	"net/http"
	"time"

	"github.com/PaulChristophel/agartha/server/config"
	"github.com/gin-contrib/sessions"
)

var jwtSecret []byte
var session sessions.Session
var ldapOptions config.LDAPOptions
var casOptions config.CASOptions
var enabledMethods = map[string]struct{}{"local": {}}
var casHTTPClient = &http.Client{Timeout: 10 * time.Second}

func SetOptions(secret []byte, methods []string, ldap config.LDAPOptions, cas config.CASOptions) {
	jwtSecret = secret
	ldapOptions = ldap
	casOptions = cas
	enabledMethods = make(map[string]struct{}, len(methods))
	for _, method := range methods {
		enabledMethods[method] = struct{}{}
	}
}
