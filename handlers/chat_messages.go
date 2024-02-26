package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetChatHistory Retrieves the chat message history for a given channel
// Endpoint: GET /v2/chat/:channel/history?page=0
func GetChatHistory(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	channel := c.Param("channel")

	if channel == "admin" && !canUserAccessAdminChat(user) {
		return APIErrorForbidden("You do not have permission to access this chat channel history.")
	}

	var messages []*db.ChatMessage
	var dbError error

	if id, err := strconv.Atoi(channel); err == nil {
		messages, dbError = db.GetPrivateChatMessageHistory(user.Id, id)
	} else {
		messages, dbError = db.GetPublicChatMessageHistory(channel)
	}

	if dbError != nil {
		return APIErrorServerError("Error retrieving chat messages", dbError)
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
	return nil
}

// Returns if the user is eligible to access the #admin chat
func canUserAccessAdminChat(user *db.User) bool {
	return enums.HasUserGroup(user.UserGroups, enums.UserGroupModerator) ||
		enums.HasUserGroup(user.UserGroups, enums.UserGroupAdmin) ||
		enums.HasUserGroup(user.UserGroups, enums.UserGroupDeveloper) ||
		enums.HasUserGroup(user.UserGroups, enums.UserGroupSwan)
}
