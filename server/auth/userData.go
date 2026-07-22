package auth

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/agartha"

	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
	"go.uber.org/zap"
)

type userData struct {
	Username          string
	FirstName         string
	LastName          string
	Email             string
	SamAccountName    string
	UserPrincipalName string
}

func auth(creds credentials, c *gin.Context) (userData, error) {
	var userData userData
	var err error
	if creds.Method != "local" && creds.Method != "ldap" && creds.Method != "cas" {
		return userData, fmt.Errorf("unsupported authentication method %q", creds.Method)
	}
	if _, enabled := enabledMethods[creds.Method]; !enabled {
		return userData, fmt.Errorf("authentication method %q is not enabled", creds.Method)
	}
	switch creds.Method {
	case "ldap":
		userData, err = authLDAP(creds.Username, creds.Password)
	case "cas":
		userData, err = authCAS(creds.Username, c)
	case "local":
		userData, err = authLocal(creds.Username, creds.Password)
	}

	return userData, err
}

func authLocal(username, password string) (userData, error) {
	var user model.AuthUser
	result := db.DB.Raw(`
		SELECT id, username, first_name, last_name, email, is_active
		FROM auth_user
		WHERE username = ?
		  AND is_active = TRUE
		  AND password = crypt(?, password)
		LIMIT 1
	`, username, password).Scan(&user)
	if result.Error != nil || result.RowsAffected != 1 {
		return userData{}, errors.New("invalid local credentials")
	}

	return userData{
		Username:          user.Username,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		Email:             user.Email,
		SamAccountName:    user.Username,
		UserPrincipalName: user.Username,
	}, nil
}

func authCAS(username string, c *gin.Context) (userData, error) {
	var log = logger.GetLogger()
	var userData userData

	ticket := c.Query("ticket")
	if ticket == "" {
		return userData, errors.New("no CAS ticket provided")
	}

	validateURL, err := url.Parse(casOptions.Server)
	if err != nil {
		return userData, fmt.Errorf("invalid CAS server URL: %w", err)
	}
	validateURL.Path, err = url.JoinPath(validateURL.Path, casOptions.ValidatePath)
	if err != nil {
		return userData, fmt.Errorf("invalid CAS validation path: %w", err)
	}
	query := validateURL.Query()
	query.Set("ticket", ticket)
	query.Set("service", casOptions.ServiceURL)
	validateURL.RawQuery = query.Encode()

	resp, err := casHTTPClient.Get(validateURL.String())
	if err != nil {
		return userData, fmt.Errorf("failed to validate CAS ticket: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// Optionally log the error
			log.Error("failed to close response body", zap.Error(cerr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return userData, errors.New("CAS ticket validation failed")
	}

	var serviceResponse CASServiceResponse
	decoder := xml.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	if err := decoder.Decode(&serviceResponse); err != nil {
		return userData, fmt.Errorf("failed to parse CAS validation response: %w", err)
	}
	if serviceResponse.AuthenticationSuccess == nil {
		return userData, errors.New("CAS ticket was not authenticated")
	}
	assertedUsername := strings.TrimSpace(serviceResponse.AuthenticationSuccess.User)
	if assertedUsername == "" {
		return userData, errors.New("CAS response did not include an authenticated user")
	}
	if username != "" && !strings.EqualFold(username, assertedUsername) {
		log.Warn("CAS asserted a different username than the login request", zap.String("requested_username", username), zap.String("asserted_username", assertedUsername))
	}

	userData.Username = assertedUsername
	userData.SamAccountName = assertedUsername
	userData.UserPrincipalName = assertedUsername

	return userData, nil
}

func authLDAP(username, password string) (userData, error) {
	var log = logger.GetLogger()
	var userData userData
	var ldap_server = ldapOptions.Server
	l, err := ldap.DialURL(ldap_server)
	if err != nil {
		return userData, fmt.Errorf("failed to connect: %w", err)
	}

	defer func() {
		if lerr := l.Close(); lerr != nil {
			// Optionally log the error
			log.Error("failed to close ldap connection", zap.Error(lerr))
		}
	}()

	accountName, userPrincipalName, err := normalizeLDAPUsername(username, ldapOptions.LDAPDomainDefault)
	if err != nil {
		return userData, err
	}
	userData.SamAccountName = accountName
	userData.UserPrincipalName = userPrincipalName

	if ldapOptions.StartTLS {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		if ldapURL, parseErr := url.Parse(ldap_server); parseErr == nil {
			if ldapURL.Scheme == "ldaps" {
				return userData, errors.New("ldap start_tls requires an ldap:// server URL, not ldaps://")
			}
			tlsConfig.ServerName = ldapURL.Hostname()
		}

		if err := l.StartTLS(tlsConfig); err != nil {
			return userData, fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// First bind with the user credentials to authenticate
	log.Debug("binding to ldap server", zap.String("server", ldap_server), zap.String("user", userData.UserPrincipalName))
	err = l.Bind(userData.UserPrincipalName, password)
	if err != nil {
		return userData, fmt.Errorf("failed to bind: %w", err)
	}

	// Rebind as a service account with permissions to search
	serviceUser := ldapOptions.User
	servicePassword := ldapOptions.Password
	err = l.Bind(serviceUser, servicePassword)
	if err != nil {
		return userData, fmt.Errorf("failed to bind as service user: %w", err)
	}

	// Authorization: check that the user satisfies the LDAP filter
	filter := fmt.Sprintf(ldapOptions.Filter, ldap.EscapeFilter(userData.SamAccountName))
	searchRequest := ldap.NewSearchRequest(
		ldapOptions.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", "sAMAccountName", "givenName", "sn", "mail"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return userData, fmt.Errorf("failed to execute search: %w", err)
	}
	if len(sr.Entries) != 1 {
		return userData, fmt.Errorf("user does not exist or too many entries returned")
	}

	entry := sr.Entries[0]
	authenticatedUsername := strings.TrimSpace(entry.GetAttributeValue("sAMAccountName"))
	if authenticatedUsername == "" {
		return userData, errors.New("LDAP entry did not include sAMAccountName")
	}
	userData.Username = authenticatedUsername
	userData.SamAccountName = authenticatedUsername
	userData.FirstName = entry.GetAttributeValue("givenName")
	userData.LastName = entry.GetAttributeValue("sn")
	userData.Email = entry.GetAttributeValue("mail")
	return userData, nil
}

func normalizeLDAPUsername(username, defaultDomain string) (string, string, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", "", errors.New("LDAP username is empty")
	}
	if strings.Count(username, "@") > 1 {
		return "", "", errors.New("LDAP username contains multiple @ separators")
	}

	accountName, domain, hasDomain := strings.Cut(username, "@")
	accountName = strings.TrimSpace(accountName)
	if accountName == "" {
		return "", "", errors.New("LDAP account name is empty")
	}
	if hasDomain {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			return "", "", errors.New("LDAP username domain is empty")
		}
		return accountName, accountName + "@" + domain, nil
	}

	defaultDomain = strings.TrimSpace(defaultDomain)
	if defaultDomain == "" {
		return "", "", errors.New("LDAP default domain is empty")
	}
	return accountName, accountName + "@" + defaultDomain, nil
}
