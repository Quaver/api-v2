package handlers

import (
	"net/http"
	"time"

	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/webhooks"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AddCrashLog Adds a new crash log into the database
// Endpoint: /v2/log/crash
func AddCrashLog(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		Runtime string  `form:"runtime" json:"runtime" binding:"required"`
		Network string  `form:"network" json:"network" binding:"required"`
		Version *string `form:"version" json:"version"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	crashLog := &db.CrashLog{
		UserId:    user.Id,
		Timestamp: time.Now().UnixMilli(),
		Runtime:   body.Runtime,
		Network:   body.Network,
	}

	if err := db.InsertCrashLog(crashLog); err != nil {
		return APIErrorServerError("Error inserting crash log", err)
	}

	if err := webhooks.SendCrashLogWebhook(user, crashLog, body.Version); err != nil {
		logrus.Error("Error sending crash log webhook: ", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your crash log has been successfully submitted."})
	return nil
}
