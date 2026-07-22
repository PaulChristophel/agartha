package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/agartha"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const authUserContextKey = "auth_user"

func AuthRequired(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get JWT token from Authorization header
		authHeader := c.GetHeader("Authorization")
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || parts[0] != "Bearer" {
			httputil.NewError(c, http.StatusUnauthorized, "Authorization header format must be Bearer {token}.")
			c.Abort()
			return
		}
		authToken := parts[1]

		// parse and validate JWT token
		token, err := jwt.Parse(
			authToken,
			func(token *jwt.Token) (any, error) { return jwtSecret, nil },
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		)

		if err != nil {
			httputil.NewError(c, http.StatusUnauthorized, "Invalid token.")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			httputil.NewError(c, http.StatusUnauthorized, "Invalid token.")
			c.Abort()
			return
		}

		username, ok := claims["username"].(string)
		if !ok || username == "" {
			httputil.NewError(c, http.StatusUnauthorized, "Invalid token claims.")
			c.Abort()
			return
		}
		userID, err := claimUserID(claims["user_id"])
		if err != nil {
			httputil.NewError(c, http.StatusUnauthorized, "Invalid token claims.")
			c.Abort()
			return
		}

		c.Set("username", username)
		c.Set("user_id", userID)

		c.Next()
	}
}

func claimUserID(value any) (uint, error) {
	userID, ok := value.(float64)
	if !ok || userID < 1 || userID != float64(uint(userID)) {
		return 0, fmt.Errorf("invalid user_id claim")
	}
	return uint(userID), nil
}

// ActiveUserRequired ensures a valid JWT still belongs to an active database user.
func ActiveUserRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, idOK := c.Get("user_id")
		username, usernameOK := c.Get("username")
		if !idOK || !usernameOK {
			httputil.NewError(c, http.StatusUnauthorized, "Invalid token claims.")
			c.Abort()
			return
		}

		var user model.AuthUser
		err := db.Where("id = ? AND username = ? AND is_active = ?", userID, username, true).First(&user).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				httputil.NewError(c, http.StatusUnauthorized, "User is inactive or no longer exists.")
			} else {
				httputil.NewError(c, http.StatusInternalServerError, "Unable to authorize user.")
			}
			c.Abort()
			return
		}

		c.Set(authUserContextKey, user)
		c.Next()
	}
}

// AuthenticatedUser returns the active user loaded by ActiveUserRequired.
func AuthenticatedUser(c *gin.Context) (model.AuthUser, bool) {
	value, exists := c.Get(authUserContextKey)
	if !exists {
		return model.AuthUser{}, false
	}
	user, ok := value.(model.AuthUser)
	return user, ok
}
