package netapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	_ "github.com/PaulChristophel/agartha/server/httputil"

	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthDetails defines the details of the authentication response
type AuthDetails struct {
	Token       string      `json:"token" example:"6572118dfaaaa84363ebf491c98e669105f2b1db"`
	Expire      float64     `json:"expire" example:"1719387217.7491617"`
	Start       float64     `json:"start" example:"1719344017.7491612"`
	User        string      `json:"user" example:"megadude"`
	Eauth       string      `json:"eauth" example:"agartha"`
	Permissions Permissions `json:"perms" example:".*,@jobs,@runner,@wheel"`
}

// AuthResponse defines the structure of the response
type AuthResponse struct {
	Return []AuthDetails `json:"return"`
}

// Perms defines the permissions
type Permissions []string

// Credentials defines the structure of the request body
type Credentials struct {
	Username string `json:"username" example:"megadude"`
	Password string `json:"password" example:"Bearer foo.bar.baz"`
	Eauth    string `json:"eauth" example:"agartha"`
}

// class salt.netapi.rest_cherrypy.app.Login(*args, **kwargs)
//
//	@ID				Login.Post()
//	@Summary		Authenticate against Salt's eauth system.
//	@Description	Log in to receive a session token. Authenticate against Salt's eauth system. If credentials are not provided, the Authorization token will be used instead. https://docs.saltproject.io/en/latest/ref/netapi/all/salt.netapi.rest_cherrypy.html#salt.netapi.rest_cherrypy.app.Login.POST
//	@Tags			NetAPI
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	AuthResponse
//	@Failure		401		{object}	httputil.HTTPError401
//	@Failure		406		{object}	httputil.HTTPError406
//	@Failure		500		{object}	httputil.HTTPError500
//	@Param			Accept	header		Accept		false	"the desired response format"
//	@Param			req		body		Credentials	false	"Login Request"
//	@router			/api/v1/netapi/login [post]
//	@Security		Bearer
func DecodeTokenAndCreateCredentials() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the request body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body", "details": err.Error()})
			c.Abort()
			return
		}

		logger.GetLogger().Sugar().Debugf("Request Body: %s", string(bodyBytes))

		// Check if the body is already a Credentials object
		var existingCreds Credentials
		if err := json.Unmarshal(bodyBytes, &existingCreds); err == nil && existingCreds.Username != "" && existingCreds.Password != "" && existingCreds.Eauth != "" {
			// Body is already a Credentials object, do nothing
			// Restore the body to the request
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			c.Set("Authorization", "")
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing Authorization header"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["username"] == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		username := claims["username"].(string)

		creds := Credentials{
			Username: username,
			Password: authHeader,
			Eauth:    "agartha",
		}

		// Replace the request body with the credentials JSON
		body, err := json.Marshal(creds)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal credentials", "details": err.Error()})
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		c.Request.ContentLength = int64(len(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("Authorization", "")

		c.Next()
	}
}
