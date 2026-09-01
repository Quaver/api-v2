package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/stringutil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"
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
	} else if regexp.MustCompile(`^\d+$`).MatchString(query) {
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

// GetUserAboutMe Retrieves a user's about me / userpage
// Endpoint: GET /v2/user/:id/aboutme
func GetUserAboutMe(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must supply a valid username or id.")
	}

	user, apiErr := getUserById(id, canAuthedUserViewBannedUsers(c))

	if apiErr != nil {
		return apiErr
	}

	if !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) || user.UserPageDisabled {
		c.JSON(http.StatusOK, gin.H{"about_me": nil})
		return nil
	}

	c.JSON(http.StatusOK, gin.H{"about_me": user.UserPage})
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

	if len(body.AboutMe) > 5000 {
		return APIErrorBadRequest("Your about me must not be longer than 5,000 characters.")
	}

	body.AboutMe = stringutil.SanitizeHTML(body.AboutMe)

	err := db.UpdateUserAboutMe(user.Id, body.AboutMe)

	if err != nil {
		return APIErrorServerError("Error updating user about me", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your about me has been successfully updated!"})
	return nil
}

const maxUserInformationValueLength = 100

// parseUserInformation parses a complete user information update payload.
func parseUserInformation(body io.Reader) (db.UserInformation, error) {
	information := db.UserInformation{
		NotifyMapsetActions: true,
		DefaultMode:         enums.GameModeKeys4,
	}

	var raw json.RawMessage
	decoder := json.NewDecoder(body)

	if err := decoder.Decode(&raw); err != nil {
		return db.UserInformation{}, err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return db.UserInformation{}, fmt.Errorf("request body must contain a single JSON object")
	}

	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '{' {
		return db.UserInformation{}, fmt.Errorf("request body must be a JSON object")
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return db.UserInformation{}, err
	}

	for _, value := range fields {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return db.UserInformation{}, fmt.Errorf("user information fields cannot be null")
		}
	}

	decoder = json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&information); err != nil {
		return db.UserInformation{}, err
	}

	for _, value := range []string{
		information.Discord,
		information.Twitter,
		information.Twitch,
		information.Youtube,
	} {
		if utf8.RuneCountInString(value) > maxUserInformationValueLength {
			return db.UserInformation{}, fmt.Errorf("user information values cannot exceed 100 characters")
		}
	}

	if information.DefaultMode != enums.GameModeKeys4 && information.DefaultMode != enums.GameModeKeys7 {
		return db.UserInformation{}, fmt.Errorf("default mode must be 1 or 2")
	}

	return information, nil
}

// UpdateUserInformation Updates the authenticated user's information.
// Endpoint: POST /v2/user/profile/information
func UpdateUserInformation(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	information, err := parseUserInformation(c.Request.Body)
	if err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if err := db.UpdateUserInformation(user.Id, information); err != nil {
		return APIErrorServerError("Error updating user information", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your user information has been successfully updated."})
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

	if targetUser.ClanId != nil {
		clan, err := db.GetClanById(*targetUser.ClanId)

		if err != nil {
			return APIErrorServerError(fmt.Sprintf("Error fetching clan for user: #%v", targetUser.Id), err)
		}

		if err := db.PerformFullClanRecalculation(clan); err != nil {
			return APIErrorServerError("Error performing full recalc on clan", err)
		}
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

	if err := db.RemoveUserFromLeaderboards(targetUser); err != nil {
		return APIErrorServerError("Error removing user from leaderboards", err)
	}

	if targetUser.ClanId != nil {
		clan, err := db.GetClanById(*targetUser.ClanId)

		if err != nil {
			return APIErrorServerError(fmt.Sprintf("Error fetching clan for user: #%v", targetUser.Id), err)
		}

		if err := db.PerformFullClanRecalculation(clan); err != nil {
			return APIErrorServerError("Error performing full recalc on clan after ban", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "User has been successfully banned."})
	return nil
}

// UpdateUserDiscordId Updates a user's discord id
// Endpoint: POST /v2/user/:id/discord
func UpdateUserDiscordId(c *gin.Context) *APIError {
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

	if err := db.UpdateUserDiscordId(user.Id, body.DiscordId); err != nil {
		return APIErrorServerError("Error updating user discord id", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "The user's Discord id has been updated."})
	return nil
}

// UpdateUserAccentColor Updates a user's accent color
// Endpoint: POST /v2/user/:id/accent
func UpdateUserAccentColor(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		AccentColor string `form:"accent_color" json:"accent_color" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if !stringutil.IsValidHexCode(body.AccentColor) {
		return APIErrorBadRequest("You must provide a valid accent color.")
	}

	if !user.AccentColorCustomizable {
		return APIErrorForbidden("You must purchase accent color customizables to access this endpoint.")
	}
	if err := db.UpdateUserAccentColor(user.Id, &body.AccentColor); err != nil {
		return APIErrorServerError("Error updating user discord id", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your accent color has been updated."})
	return nil
}
