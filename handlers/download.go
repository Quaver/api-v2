package handlers

import (
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/files"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// DownloadQua Handles the downloading of an individual .qua file
// Endpoint: GET /v2/download/map/:id
func DownloadQua(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	mapQua, err := db.GetMapById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving map file from db", err)
	}

	if mapQua == nil {
		return APIErrorNotFound("Map")
	}

	path, err := files.CacheQuaFile(mapQua)

	if err != nil {
		return APIErrorServerError("Error caching .qua file", err)
	}

	c.FileAttachment(path, fmt.Sprintf("%v.qua", mapQua.Id))
	return nil
}
