package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetUserApplications Returns a users active applications
// Endpoint: GET /v2/developers/applications
func GetUserApplications(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	applications, err := db.GetUserActiveApplications(user.Id)

	if err != nil {
		return APIErrorServerError("Error retrieving applications from the db", err)
	}
	
	for _, app := range applications {
		app.ClientSecret = ""
	}

	c.JSON(http.StatusOK, gin.H{"applications": applications})
	return nil
}