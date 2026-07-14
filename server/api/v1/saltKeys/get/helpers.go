package saltKeys

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

// splitBankAndKey splits a catch-all Gin route into a Salt bank and key.
func splitBankAndKey(rawBank, rawRest string) (string, string, error) {
	rawRest = strings.TrimPrefix(rawRest, "/")

	bank, err := url.PathUnescape(rawBank)
	if err != nil {
		return "", "", err
	}
	rest, err := url.PathUnescape(rawRest)
	if err != nil {
		return "", "", err
	}
	if bank == "" || rest == "" {
		return "", "", nil
	}

	parts := strings.Split(rest, "/")
	key := parts[len(parts)-1]
	if key == "" {
		return "", "", nil
	}
	if len(parts) > 1 {
		bankExtra := strings.Join(parts[:len(parts)-1], "/")
		if bankExtra != "" {
			if strings.HasSuffix(bank, "/") {
				bank = bank + bankExtra
			} else {
				bank = bank + "/" + bankExtra
			}
		}
	}

	return bank, key, nil
}

// keyState extracts the key state from a salt_keys JSON payload.
func keyState(data any) string {
	keyData, ok := data.(map[string]any)
	if !ok {
		return ""
	}
	state, ok := keyData["state"].(string)
	if !ok {
		return ""
	}
	return state
}
