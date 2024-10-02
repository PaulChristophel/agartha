package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
)

type userData struct {
	FirstName         string
	LastName          string
	Email             string
	SamAccountName    string
	UserPrincipalName string
}

func auth(creds credentials, c *gin.Context) (userData, error) {
	var userData userData
	var err error
	switch creds.Method {
	case "ldap":
		userData, err = authLDAP(creds.Username, creds.Password)
	case "cas":
		userData, err = authCAS(creds.Username, c)
	default:
		userData, err = authLocal(creds.Username, creds.Password)
	}

	return userData, err
}

func authLocal(username, password string) (userData, error) {
	// demonstration. Update this to actually check the database
	if username == "localuser" && password == "localpassword" {
		return userData{
			FirstName:         "Local",
			LastName:          "User",
			Email:             "local.user@example.com",
			SamAccountName:    username,
			UserPrincipalName: username,
		}, nil
	}
	return userData{}, errors.New("invalid local credentials")
}

func authCAS(username string, c *gin.Context) (userData, error) {
	var userData userData

	ticket := c.Query("ticket")
	if ticket == "" {
		redirectURL := fmt.Sprintf("%s/login?service=%s", casOptions.Server, casOptions.ServiceURL)
		c.Redirect(http.StatusFound, redirectURL)
		return userData, errors.New("no CAS ticket provided")
	}

	validateURL := fmt.Sprintf("%s/serviceValidate?ticket=%s&service=%s", casOptions.Server, ticket, casOptions.ServiceURL)
	resp, err := http.Get(validateURL)
	if err != nil {
		return userData, fmt.Errorf("failed to validate CAS ticket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return userData, errors.New("CAS ticket validation failed")
	}

	userData.FirstName = "First"
	userData.LastName = "Last"
	userData.Email = "email@example.com"
	userData.SamAccountName = username
	userData.UserPrincipalName = username

	return userData, nil
}

func authLDAP(username, password string) (userData, error) {
	var userData userData
	var ldap_server = ldapOptions.Server
	l, err := ldap.DialURL(ldap_server)
	if err != nil {
		return userData, fmt.Errorf("failed to connect: %w", err)
	}
	defer l.Close()

	ldapDomain := ldapOptions.LDAPDomainDefault
	if !strings.HasSuffix(username, ldapDomain) {
		userData.SamAccountName = username
		userData.UserPrincipalName = username + "@" + ldapDomain
	} else {
		userData.UserPrincipalName = username
		userData.SamAccountName = strings.Split(username, "@")[1]
	}

	// First bind with the user credentials to authenticate
	log.Printf("Binding to %s as %s", ldap_server, userData.UserPrincipalName)
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
	filter := fmt.Sprintf(ldapOptions.Filter, userData.SamAccountName)
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
	userData.FirstName = entry.GetAttributeValue("givenName")
	userData.LastName = entry.GetAttributeValue("sn")
	userData.Email = entry.GetAttributeValue("mail")
	return userData, nil
}
