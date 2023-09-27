package handlers

import (
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
