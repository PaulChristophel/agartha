package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/agartha"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SaltCapability is an Agartha authorization boundary derived from the
// effective permissions returned by Salt eauth.
type SaltCapability string

const (
	ReadSaltData       SaltCapability = "read Salt data"
	ExecuteSaltCommand SaltCapability = "execute Salt commands"
)

// SaltPermissionRequired authorizes an operational Salt capability. Staff and
// superusers are explicit operational overrides; ordinary users are authorized
// from the effective Salt permissions cached after Salt login.
func SaltPermissionRequired(database *gorm.DB, capability SaltCapability) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := AuthenticatedUser(c)
		if !ok {
			httputil.NewError(c, http.StatusUnauthorized, "User authorization context is missing.")
			c.Abort()
			return
		}
		if user.IsSuperuser || user.IsStaff {
			c.Next()
			return
		}

		permissions, ok := loadSaltPermissions(c, database, user.ID)
		if !ok {
			return
		}

		allowed := hasSaltPermission(permissions)
		if capability == ExecuteSaltCommand {
			allowed = hasExecutableSaltPermission(permissions)
		}
		if !allowed {
			httputil.NewError(c, http.StatusForbidden, "Permission denied: cannot "+string(capability)+".")
			c.Abort()
			return
		}
		c.Next()
	}
}

// SaltPermissionForMethodRequired treats safe HTTP methods as reads and all
// other methods as Salt operations.
func SaltPermissionForMethodRequired(database *gorm.DB) gin.HandlerFunc {
	read := SaltPermissionRequired(database, ReadSaltData)
	execute := SaltPermissionRequired(database, ExecuteSaltCommand)
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			read(c)
			return
		}
		execute(c)
	}
}

// SaltWheelPermissionRequired requires Salt wheel authorization for a specific
// function such as key.accept. A bare @wheel grant authorizes every wheel
// function; scoped @wheel ACLs authorize only matching functions.
func SaltWheelPermissionRequired(database *gorm.DB, function string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := AuthenticatedUser(c)
		if !ok {
			httputil.NewError(c, http.StatusUnauthorized, "User authorization context is missing.")
			c.Abort()
			return
		}
		if user.IsSuperuser || user.IsStaff {
			c.Next()
			return
		}

		permissions, ok := loadSaltPermissions(c, database, user.ID)
		if !ok {
			return
		}
		if !hasWheelPermission(permissions, function) {
			httputil.NewError(c, http.StatusForbidden, "Permission denied: cannot manage minion keys.")
			c.Abort()
			return
		}
		c.Next()
	}
}

// SaltWheelAdministrationRequired protects raw salt_keys access, including
// master-key material. Scoped wheel grants are insufficient for this boundary.
func SaltWheelAdministrationRequired(database *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := AuthenticatedUser(c)
		if !ok {
			httputil.NewError(c, http.StatusUnauthorized, "User authorization context is missing.")
			c.Abort()
			return
		}
		if user.IsSuperuser || user.IsStaff {
			c.Next()
			return
		}

		permissions, ok := loadSaltPermissions(c, database, user.ID)
		if !ok {
			return
		}
		if !containsPermissionString(permissions, "@wheel") {
			httputil.NewError(c, http.StatusForbidden, "Permission denied: raw Salt key administration requires full @wheel access.")
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdministrationRequired reserves Agartha user and settings administration for
// superusers. Salt command permissions and is_staff do not cross this boundary.
func AdministrationRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := AuthenticatedUser(c)
		if !ok {
			httputil.NewError(c, http.StatusUnauthorized, "User authorization context is missing.")
			c.Abort()
			return
		}
		if !user.IsSuperuser {
			httputil.NewError(c, http.StatusForbidden, "Permission denied: Agartha administration requires superuser access.")
			c.Abort()
			return
		}
		c.Next()
	}
}

func loadSaltPermissions(c *gin.Context, database *gorm.DB, userID uint) (any, bool) {
	var settings model.UserSettings
	err := database.Select("salt_permissions").Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httputil.NewError(c, http.StatusForbidden, "Permission denied: no Salt permissions are available.")
		} else {
			httputil.NewError(c, http.StatusInternalServerError, "Unable to authorize Salt access.")
		}
		c.Abort()
		return nil, false
	}

	var permissions any
	if err := json.Unmarshal([]byte(settings.SaltPermissions), &permissions); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, "Unable to authorize Salt access.")
		c.Abort()
		return nil, false
	}
	return permissions, true
}

func hasSaltPermission(value any) bool {
	switch permission := value.(type) {
	case string:
		return strings.TrimSpace(permission) != ""
	case []any:
		for _, item := range permission {
			if hasSaltPermission(item) {
				return true
			}
		}
	case map[string]any:
		return len(permission) > 0
	}
	return false
}

func hasExecutableSaltPermission(value any) bool {
	switch permission := value.(type) {
	case string:
		permission = strings.TrimSpace(permission)
		return permission != "" && permission != "@jobs"
	case []any:
		for _, item := range permission {
			if hasExecutableSaltPermission(item) {
				return true
			}
		}
	case map[string]any:
		for scope, entries := range permission {
			if scope != "@jobs" && hasSaltPermission(entries) {
				return true
			}
		}
	}
	return false
}

func containsPermissionString(value any, wanted string) bool {
	switch permission := value.(type) {
	case string:
		return permission == wanted
	case []any:
		for _, item := range permission {
			if containsPermissionString(item, wanted) {
				return true
			}
		}
	}
	return false
}

func hasWheelPermission(value any, function string) bool {
	switch permission := value.(type) {
	case string:
		return permission == "@wheel"
	case []any:
		for _, item := range permission {
			if hasWheelPermission(item, function) {
				return true
			}
		}
	case map[string]any:
		entries, ok := permission["@wheel"]
		return ok && matchesSaltFunction(entries, function)
	}
	return false
}

func matchesSaltFunction(value any, function string) bool {
	switch permission := value.(type) {
	case string:
		expression, err := regexp.Compile("^(?:" + permission + ")$")
		return err == nil && expression.MatchString(function)
	case []any:
		for _, item := range permission {
			if matchesSaltFunction(item, function) {
				return true
			}
		}
	}
	return false
}
