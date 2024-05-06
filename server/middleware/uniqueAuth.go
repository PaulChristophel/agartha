package middleware

import (
	"errors"
	"net/http"
	"strconv"

	model "github.com/PaulChristophel/agartha/server/model/agartha"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ResourceAccessRequest struct {
	ID int `json:"id"` // This assumes the ID will be provided as a JSON field.
}

// Ensures that the user authing has permissions to perform the specific action
func UniqueAuthRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the user ID from the context, set by AuthRequired middleware
		usernameInterface, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}
		username := usernameInterface.(string)

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

		// Retrieve the resource's owner from the database
		var authUser model.AuthUser
		if err := db.Where("id = ?", id).First(&authUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			c.Abort()
			return
		}

		// Check if the authenticated user is the owner of the resource
		if authUser.IsSuperuser {
			c.Next()
			return
		}

		if username != authUser.Username {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access to resource"})
			c.Abort()
			return
		}

		// If the user is authorized, continue with the request
		c.Next()
	}
}
