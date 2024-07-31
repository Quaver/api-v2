package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetRankingSupervisorActions Retrieves ranking supervisor actions between a given time frame
// Endpoint: GET /ranking/queue/supervisors/actions?start=&end=
func GetRankingSupervisorActions(c *gin.Context) *APIError {
	body := struct {
		Start int64 `form:"start" json:"start"`
		End   int64 `form:"end" json:"end"`
	}{}

	if err := c.ShouldBindQuery(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	supervisors, err := db.GetRankingSupervisors(false)

	if err != nil {
		return APIErrorServerError("Error retrieving ranking supervisors from db", err)
	}

	type supervisorActionData struct {
		User    *db.User                        `json:"user"`
		Actions []*db.MapsetRankingQueueComment `json:"actions"`
	}

	supervisorActions := make([]supervisorActionData, 0)

	for _, supervisor := range supervisors {
		actions, err := db.GetUserRankingQueueComments(supervisor.Id, body.Start, body.End)

		if err != nil {
			return APIErrorServerError("Error retrieving supervisor comments from db", err)
		}

		supervisorActions = append(supervisorActions, supervisorActionData{
			User:    supervisor,
			Actions: actions,
		})
	}

	c.JSON(http.StatusOK, gin.H{"supervisor_actions": supervisorActions})
	return nil
}
