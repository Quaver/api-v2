package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetTeamMembers Retrieves users in the Quaver team along with contributors and donators
// Endpoint: /v2/user/team/members
func GetTeamMembers(c *gin.Context) *APIError {
	team, err := db.GetTeamMembers()

	if err != nil {
		return APIErrorServerError("Error getting team members", err)
	}

	c.JSON(http.StatusOK, gin.H{"team": team})
	return nil
}
