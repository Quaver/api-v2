package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

// CreateHandler Creates a handler with automatic error handling
func CreateHandler(fn func(*gin.Context) *APIError) func(*gin.Context) {
	return func(c *gin.Context) {
		err := fn(c)

		if err == nil {
			return
		}

		if err.Error != nil {
			logrus.Errorf("%v - %v", err.Message, err.Error)
		}

		if err.Status == http.StatusInternalServerError {
			c.JSON(err.Status, gin.H{"error": "Internal Server Error"})
			return
		}

		c.JSON(err.Status, gin.H{"error": err.Message})
	}
}

// Returns an authenticated user from a context
func getAuthedUser(c *gin.Context) *db.User {
	user, exists := c.Get("user")

	if !exists {
		return nil
	}

	return user.(*db.User)
}

// Gets the ip address from the request
func getIpFromRequest(c *gin.Context) string {
	// Running under NGINX
	ip := c.GetHeader("X-Forwarded-For")

	if ip != "" {
		return ip
	}

	return "::1"
}
