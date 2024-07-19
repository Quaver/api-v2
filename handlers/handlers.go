package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

// Gets a user by id and checks if they exist
func getUserById(id int, displayBannedUser bool) (*db.User, *APIError) {
	user, err := db.GetUserById(id)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, APIErrorNotFound("User")
		}

		return nil, APIErrorServerError("Error retrieving user from db", err)
	}

	if !user.Allowed && !displayBannedUser {
		return nil, APIErrorNotFound("User")
	}

	return user, nil
}

// Returns if the authed user can view banned users
func canAuthedUserViewBannedUsers(c *gin.Context) bool {
	user := getAuthedUser(c)

	if user == nil {
		return false
	}

	return enums.HasUserGroup(user.UserGroups, enums.UserGroupSwan) ||
		enums.HasUserGroup(user.UserGroups, enums.UserGroupDeveloper) ||
		enums.HasUserGroup(user.UserGroups, enums.UserGroupAdmin) ||
		enums.HasUserGroup(user.UserGroups, enums.UserGroupModerator) ||
		enums.HasUserGroup(user.UserGroups, enums.UserGroupBot)
}

// Returns if a user can access an admin only route
func canUserAccessAdminRoute(c *gin.Context) bool {
	user := getAuthedUser(c)

	if user == nil {
		return false
	}

	return enums.HasUserGroup(user.UserGroups, enums.UserGroupSwan) ||
		enums.HasUserGroup(user.UserGroups, enums.UserGroupDeveloper) ||
		enums.HasUserGroup(user.UserGroups, enums.UserGroupAdmin) ||
		enums.HasUserGroup(user.UserGroups, enums.UserGroupBot)
}
