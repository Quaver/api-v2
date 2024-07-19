package handlers

import (
	"encoding/json"
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// CreateUserNotification Inserts a user notification into the db
// Endpoint: POST /v2/notifications
func CreateUserNotification(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !canUserAccessAdminRoute(c) {
		return APIErrorForbidden("You do not have permission to access this endpoint.")
	}

	body := struct {
		SenderId   int                         `form:"sender_id" json:"sender_id" binding:"required"`
		ReceiverId int                         `form:"receiver_id" json:"receiver_id" binding:"required"`
		Type       db.UserNotificationType     `form:"type" json:"type" binding:"required"`
		Category   db.UserNotificationCategory `form:"category" json:"category" binding:"required"`
		Data       string                      `form:"data" json:"data" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	var parsed interface{}

	if err := json.Unmarshal([]byte(body.Data), &parsed); err != nil {
		return APIErrorBadRequest("Invalid JSON provided in `data` field")
	}

	notification := &db.UserNotification{
		SenderId:   body.SenderId,
		ReceiverId: body.ReceiverId,
		Type:       body.Type,
		Category:   body.Category,
		RawData:    body.Data,
	}

	if err := notification.Insert(); err != nil {
		return APIErrorServerError("Error inserting notification into db", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your notification has been successfully created."})
	return nil
}

// GetUserNotifications Gets a user's notification
// Endpoint /v2/notifications
func GetUserNotifications(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		Categories []db.UserNotificationCategory `form:"categories" json:"categories"`
		Unread     bool                          `form:"unread" json:"unread"`
		Page       int                           `form:"page" json:"page"`
	}{}

	if err := c.ShouldBindQuery(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	notifications, err := db.GetNotifications(user.Id, body.Unread, body.Page, 20, body.Categories...)

	if err != nil {
		return APIErrorServerError("Error retrieving notifications", err)
	}

	filteredCount, err := db.GetNotificationCount(user.Id, body.Unread, body.Categories...)

	if err != nil {
		return APIErrorServerError("Error getting notification filtered notification count", err)
	}

	unreadCount, err := db.GetTotalUnreadNotifications(user.Id)

	if err != nil {
		return APIErrorServerError("Error getting user unread notification count", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_count_unread":   unreadCount,
		"total_count_filtered": filteredCount,
		"notifications":        notifications,
	})

	return nil
}

// MarkUserNotificationAsRead Marks an individual notification as read
// Endpoint: POST /v2/notifications/:id/read
func MarkUserNotificationAsRead(c *gin.Context) *APIError {
	if apiErr := updateNotificationReadStatus(c, true); apiErr != nil {
		return apiErr
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your notification has been marked as read."})
	return nil
}

// MarkUserNotificationAsUnread Marks an individual notification as unread
// Endpoint: POST /v2/notifications/:id/unread
func MarkUserNotificationAsUnread(c *gin.Context) *APIError {
	if apiErr := updateNotificationReadStatus(c, false); apiErr != nil {
		return apiErr
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your notification has been marked as unread."})
	return nil
}

// Performs authentication/validation & Sets the notification read status
func updateNotificationReadStatus(c *gin.Context, isRead bool) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	notification, err := db.GetNotificationById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving notification from db", err)
	}

	if notification == nil {
		return APIErrorNotFound("Notification")
	}

	if notification.ReceiverId != user.Id {
		return APIErrorForbidden("You are not the receiver of this notification.")
	}

	if err := notification.UpdateReadStatus(isRead); err != nil {
		return APIErrorServerError("Error updating notification read status", err)
	}

	return nil
}
