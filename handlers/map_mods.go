package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetMapMods Gets mods for a given map
// Endpoint: GET /v2/maps/:id/mods
func GetMapMods(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	mods, err := db.GetMapMods(id)

	if err != nil {
		return APIErrorServerError("Error retrieving map mods from db", err)
	}

	c.JSON(http.StatusOK, gin.H{"mods": mods})
	return nil
}
