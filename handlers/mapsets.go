package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return APIErrorNotFound("Mapset")
	default:
		return APIErrorServerError("Failed to get mapset from database", err)
	}

	c.JSON(http.StatusOK, gin.H{"mapset": mapset})
	return nil
}

// GetUserMapsets Gets a user's uploaded mapsets
// Endpoint: GET /v2/user/:id/mapsets
func GetUserMapsets(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	mapsets, err := db.GetUserMapsets(id)

	if err != nil {
		return APIErrorServerError("Failed to get mapsets from database", err)
	}

	c.JSON(http.StatusOK, gin.H{"mapsets": mapsets})
	return nil
}
