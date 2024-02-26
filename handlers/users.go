package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// SearchUsers Searches for users by username and returns them
// Endpoint: /v2/user/search/:name
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

// GetUser Gets a user by their id or username
// Endpoint: /v2/user/:query
func GetUser(c *gin.Context) *APIError {
	query := c.Param("id")

	if query == "" {
		return APIErrorBadRequest("You must supply a valid username or id.")
	}

	var user *db.User
	var dbError error

	if id, err := strconv.Atoi(query); err == nil {
		user, dbError = db.GetUserById(id)
	} else {
		user, dbError = db.GetUserByUsername(query)
	}

	switch dbError {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return APIErrorNotFound("User")
	default:
		return APIErrorServerError("Error retrieving user from database", dbError)
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
	return nil
}

// UpdateUserAboutMe Updates a user's about me
// Endpoint: PATCH /v2/user/aboutme
func UpdateUserAboutMe(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		return APIErrorForbidden("You must be a donator to update your about me.")
	}

	body := struct {
		AboutMe string `form:"about_me" json:"about_me"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if len(body.AboutMe) > 2000 {
		return APIErrorBadRequest("Your about me must not be longer than 2,000 characters.")
	}

	err := db.UpdateUserAboutMe(user.Id, body.AboutMe)

	if err != nil {
		return APIErrorServerError("Error updating user about me", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your about me has been successfully updated!"})
	return nil
}
