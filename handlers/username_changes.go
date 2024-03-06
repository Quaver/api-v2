package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetCanUserChangeUsername Returns if a user is eligible to change their username
// Endpoint: GET /v2/user/profile/username/eligible
func GetCanUserChangeUsername(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		return APIErrorForbidden("You must be a donator to change your username.")
	}

	eligible, nextChangeTime, err := db.CanUserChangeUsername(user.Id)

	if err != nil {
		return APIErrorServerError("Error checking if user can change username", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"is_eligible":      eligible,
		"next_change_time": nextChangeTime,
	})
	return nil
}

// IsUsernameAvailable Returns if a username is available for a user to take
// Endpoint: GET /v2/user/profile/username/available?name=
func IsUsernameAvailable(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	name := c.Query("name")

	if name == "" {
		return APIErrorBadRequest("You need to provide a `name` query parameter.")
	}

	available, err := db.IsUsernameAvailable(user.Id, name)

	if err != nil {
		return APIErrorServerError("Error checking if username is available", err)
	}

	c.JSON(http.StatusOK, gin.H{"username_available": available})
	return nil
}
