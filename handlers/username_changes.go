package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

// ChangeUserUsername Changes a user's username
// Endpoint: POST /v2/user/profile/username
func ChangeUserUsername(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		return APIErrorForbidden("You must be a donator to change your username.")
	}

	body := struct {
		Username string `form:"username" json:"username" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	isUsernameFlagged, err := isTextFlagged(body.Username)

	if err != nil {
		logrus.Error("Error checking if username is flagged", err)
	}

	if isUsernameFlagged {
		return APIErrorBadRequest("The username you have chosen has been flagged as inappropriate.")
	}

	changed, reason, err := db.ChangeUserUsername(user.Id, user.Username, body.Username)

	if err != nil {
		return APIErrorServerError("Error changing username", err)
	}

	if !changed {
		return APIErrorBadRequest(reason)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your username has been successfully changed."})
	return nil
}
