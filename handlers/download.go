package handlers

import (
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/files"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

// DownloadMapset Handles the downloading of an individual .qp file
// Endpoint: GET /v2/download/mapset/:id
func DownloadMapset(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	mapset, err := db.GetMapsetById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error getting mapset by id", err)
	}

	if mapset == nil || !mapset.IsVisible {
		return APIErrorNotFound("Mapset")
	}

	path, err := files.CacheMapset(mapset)

	if err != nil {
		logrus.Error("Error caching mapset", err)
		return APIErrorNotFound("Mapset")
	}

	c.FileAttachment(path, fmt.Sprintf("%v.qp", mapset.Id))
	return nil
}
