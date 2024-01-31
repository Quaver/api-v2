package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// GetMap Gets an individual map's data
// Endpoint: /v2/map/:id
func GetMap(c *gin.Context) *APIError {
	query := c.Param("id")

	if query == "" {
		return APIErrorBadRequest("You must supply a valid id or md5 hash.")
	}

	var qua *db.MapQua
	var dbError error

	if id, err := strconv.Atoi(query); err == nil {
		qua, dbError = db.GetMapById(id)
	} else {
		qua, dbError = db.GetMapByMD5(query)
	}

	switch dbError {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return APIErrorNotFound("Map")
	default:
		return APIErrorServerError("Error retrieving map from database", dbError)
	}

	c.JSON(http.StatusOK, gin.H{"map": qua})
	return nil
}
