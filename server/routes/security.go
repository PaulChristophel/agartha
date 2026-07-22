package routes

import "github.com/gin-gonic/gin"

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		headers := c.Writer.Header()
		headers.Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; connect-src 'self'; font-src 'self' data:; form-action 'self'; frame-ancestors 'none'; img-src 'self' data:; object-src 'none'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		headers.Set("Permissions-Policy", "camera=(), geolocation=(), microphone=()")
		headers.Set("Referrer-Policy", "no-referrer")
		headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		headers.Set("X-Content-Type-Options", "nosniff")
		headers.Set("X-Frame-Options", "DENY")
		c.Next()
	}
}
