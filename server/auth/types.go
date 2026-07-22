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
var casHTTPClient = &http.Client{Timeout: 10 * time.Second}

func SetOptions(secret []byte, ldap config.LDAPOptions, cas config.CASOptions) {
	jwtSecret = secret
	ldapOptions = ldap
	casOptions = cas
}
