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
