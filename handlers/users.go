package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"time"
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

	value, err := strconv.Atoi(query)

	if err == nil && value <= math.MaxInt32 {
		user, dbError = db.GetUserById(value)
	} else if regexp.MustCompile(`\d`).MatchString(query) {
		user, dbError = db.GetUserBySteamId(query)
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

	if !user.Allowed && !canAuthedUserViewBannedUsers(c) {
		return APIErrorNotFound("User")
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
	return nil
}

// UpdateUserAboutMe Updates a user's about me
// Endpoint: POST /v2/user/profile/aboutme
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

// UnbanUser Unbans a user from the game
// Endpoint: POST /v2/user/:id/unban
func UnbanUser(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must supply a valid username or id.")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasPrivilege(user.Privileges, enums.PrivilegeBanUsers) {
		return APIErrorForbidden("You do not have permission to access this resource.")
	}

	targetUser, err := db.GetUserById(id)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return APIErrorNotFound("User not found")
	default:
		return APIErrorServerError("Error occurred while fetching user by id", err)
	}

	if targetUser.Allowed {
		return APIErrorBadRequest("This user is not banned.")
	}

	if err := db.UpdateUserAllowed(targetUser.Id, true); err != nil {
		return APIErrorServerError("Error changing user allowed status", err)
	}

	log := db.AdminActionLog{
		AuthorId:       user.Id,
		AuthorUsername: user.Username,
		TargetId:       targetUser.Id,
		TargetUsername: targetUser.Username,
		Action:         db.AdminActionBanned,
		Notes:          "User Banned",
		Timestamp:      time.Now().UnixMilli(),
	}

	if err := log.Insert(); err != nil {
		return APIErrorServerError("Error inserting admin action log", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User has been successfully unbanned."})
	return nil
}

func BanUser(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must supply a valid username or id.")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasPrivilege(user.Privileges, enums.PrivilegeBanUsers) {
		return APIErrorForbidden("You do not have permission to access this resource.")
	}

	targetUser, err := db.GetUserById(id)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return APIErrorNotFound("User not found")
	default:
		return APIErrorServerError("Error occurred while fetching user by id", err)
	}

	if !targetUser.Allowed {
		return APIErrorBadRequest("This user is already banned.")
	}

	if err := db.UpdateUserAllowed(targetUser.Id, false); err != nil {
		return APIErrorServerError("Error changing user allowed status", err)
	}

	if err := db.ReplaceUserFirstPlaces(targetUser.Id); err != nil {
		return APIErrorServerError("Error updating first place scores", err)
	}

	log := db.AdminActionLog{
		AuthorId:       user.Id,
		AuthorUsername: user.Username,
		TargetId:       targetUser.Id,
		TargetUsername: targetUser.Username,
		Action:         db.AdminActionUpdated,
		Notes:          "User Unbanned",
		Timestamp:      time.Now().UnixMilli(),
	}

	if err := log.Insert(); err != nil {
		return APIErrorServerError("Error inserting admin action log", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User has been successfully banned."})
	return nil
}

// UpdateUserDiscordId Updates a user's discord id
// Endpoint: POST /v2/user/:id/discord
func UpdateUserDiscordId(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must supply a valid username or id.")
	}

	body := struct {
		DiscordId *string `form:"discord_id" json:"discord_id" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasPrivilege(user.Privileges, enums.PrivilegeEditUsers) {
		return APIErrorForbidden("You do not have permission to access this resource.")
	}

	if err := db.UpdateUserDiscordId(id, body.DiscordId); err != nil {
		return APIErrorServerError("Error updating user discord id", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "The user's Discord id has been updated."})
	return nil
}
