package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ResourceAccessRequest struct {
	ID int `json:"id"` // This assumes the ID will be provided as a JSON field.
}

// Ensures that the user authing has permissions to perform the specific action
func UniqueAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, ok := AuthenticatedUser(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User authorization context is missing"})
			c.Abort()
			return
		}

		// Try to extract the resource ID from the URL
		resourceID := c.Param("id")
		if resourceID == "" {
			// If no ID in URL, attempt to read from body
			var accessReq ResourceAccessRequest
			if err := c.ShouldBindJSON(&accessReq); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Resource ID is required"})
				c.Abort()
				return
			}
			resourceID = strconv.Itoa(accessReq.ID)
		}

		// Convert the resource ID to an integer
		id, err := strconv.Atoi(resourceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
			c.Abort()
			return
		}

		if currentUser.IsSuperuser {
			c.Next()
			return
		}

		if uint(id) != currentUser.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access to resource"})
			c.Abort()
			return
		}

		// If the user is authorized, continue with the request
		c.Next()
	}
}
