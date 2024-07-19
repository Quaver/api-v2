package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
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
		SenderId   int    `form:"sender_id" json:"sender_id" binding:"required"`
		ReceiverId int    `form:"receiver_id" json:"receiver_id" binding:"required"`
		Type       int    `form:"type" json:"type" binding:"required"`
		Data       string `form:"data" json:"data" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		fmt.Println(err)
		return APIErrorBadRequest("Invalid request body")
	}

	var parsed interface{}

	if err := json.Unmarshal([]byte(body.Data), &parsed); err != nil {
		return APIErrorBadRequest("Invalid JSON provided in `data` field")
	}

	notification := &db.UserNotification{
		SenderId:   body.SenderId,
		ReceiverId: body.ReceiverId,
		Type:       int8(body.Type),
		RawData:    body.Data,
	}

	if err := notification.Insert(); err != nil {
		return APIErrorServerError("Error inserting notification into db", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your notification has been successfully created."})
	return nil
}
