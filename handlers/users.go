package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
)

// SearchUsers Searches for users by username and returns them
// Endpoint: /v2/users/search/:name
func SearchUsers(c *gin.Context) *APIError {
	name := c.Param("name")

	if name == "" {
		return APIErrorBadRequest("You must supply a valid name to search.")
	}

	users, err := db.SearchUsersByName(name)

	if err != nil {
		return APIErrorServerError("Error searching for users", err)
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
	return nil
}
