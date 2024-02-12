package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetMapsetById Retrieves a mapset from the database by its id
// Endpoint: GET /v2/mapset/:id
func GetMapsetById(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	mapset, err := db.GetMapsetById(id)

	if err != nil {
		return APIErrorServerError("Failed to get mapset from database", err)
	}

	c.JSON(http.StatusOK, gin.H{"mapset": mapset})
	return nil
}
