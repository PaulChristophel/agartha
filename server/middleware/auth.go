package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get JWT token from Authorization header
		authHeader := c.GetHeader("Authorization")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			httputil.NewError(c, http.StatusUnauthorized, "Authorization header format must be Bearer {token}.")
			c.Abort()
			return
		}
		authToken := parts[1]

		// parse and validate JWT token
		token, err := jwt.Parse(authToken, func(token *jwt.Token) (any, error) {
			// ensure the token method conforms to "SigningMethodHMAC"
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			httputil.NewError(c, http.StatusUnauthorized, "Invalid token.")
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Store the username in the context
			c.Set("username", claims["username"])
		} else {
			httputil.NewError(c, http.StatusUnauthorized, "Invalid token.")
			c.Abort()
			return
		}

		c.Next()
	}
}
