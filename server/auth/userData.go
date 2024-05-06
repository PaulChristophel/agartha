package auth

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/spf13/viper"
)

type UserData struct {
	FirstName         string
	LastName          string
	Email             string
	SamAccountName    string
	UserPrincipalName string
}

func authenticate(username, password string) (UserData, error) {
	var userData UserData
	var ldap_server = viper.GetString("ldap_server")
	l, err := ldap.DialURL(ldap_server)
	if err != nil {
		return userData, fmt.Errorf("failed to connect: %w", err)
	}
	defer l.Close()

	ldapDomain := viper.GetString("ldap_domain_default")
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
	serviceUser := viper.GetString("ldap_user")
	servicePassword := viper.GetString("ldap_password")
	err = l.Bind(serviceUser, servicePassword)
	if err != nil {
		return userData, fmt.Errorf("failed to bind as service user: %w", err)
	}

	// Authorization: check that the user satisfies the LDAP filter
	filter := fmt.Sprintf(viper.GetString("ldap_filter"), userData.SamAccountName)
	searchRequest := ldap.NewSearchRequest(
		viper.GetString("ldap_base_dn"), // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", "sAMAccountName", "givenName", "sn", "mail"}, // A list attributes to retrieve
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
