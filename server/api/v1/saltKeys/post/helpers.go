package saltKeys

import (
	"errors"
	"net/http"

	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	keysBank   = "pki/master/keys"
	deniedBank = "pki/master/denied_keys"
)

var errSaltKeysUnavailable = errors.New("salt_keys table is unavailable")

// ensureSaltKeysTable verifies that the configured salt_keys table exists.
func ensureSaltKeysTable(conn *gorm.DB) error {
	if !conn.Migrator().HasTable(table) {
		return errSaltKeysUnavailable
	}
	return nil
}

// writeSaltKeysError writes a consistent HTTP error for salt_keys failures.
func writeSaltKeysError(c *gin.Context, err error) {
	if errors.Is(err, errSaltKeysUnavailable) {
		httputil.NewError(c, http.StatusNotFound, "salt_keys table is unavailable")
		return
	}
	httputil.NewError(c, http.StatusInternalServerError, err.Error())
}

// uniqueStrings returns non-empty string values without duplicates.
func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	return unique
}
