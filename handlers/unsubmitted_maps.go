package handlers

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/qua"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

// UploadUnsubmittedMap Uploads an unsubmitted map to azure for donator leaderboards.
// Endpoint: POST /v2/map
func UploadUnsubmittedMap(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		return APIErrorForbidden("You must be a donator to access this resource.")
	}

	body := struct {
		MD5            string `form:"md5" json:"md5" binding:"required"`
		AlternativeMD5 string `form:"alternative_md5" json:"alternative_md5" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	// Check if map already exists
	foundMap, err := db.GetMapByMD5AndAlternative(body.MD5, body.AlternativeMD5)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving map from database", err)
	}

	if foundMap != nil {
		return APIErrorBadRequest("This map has already been uploaded to the server.")
	}

	fileHeader, _ := c.FormFile("map")

	if fileHeader == nil {
		return APIErrorBadRequest("You must provide a valid `map` .qua file.")
	}

	file, err := fileHeader.Open()

	if err != nil {
		return APIErrorServerError("Error opening file", err)
	}

	defer file.Close()

	var fileBytes = make([]byte, fileHeader.Size)

	if _, err = file.Read(fileBytes); err != nil {
		return APIErrorServerError("Error reading file", err)
	}

	quaFile, err := qua.Parse(fileBytes)

	if err != nil {
		return APIErrorBadRequest("The file you provided was not a valid .qua")
	}

	// Insert map into DB
	dbMap := &db.MapQua{
		MapsetId:        -1,
		MD5:             body.MD5,
		AlternativeMD5:  body.AlternativeMD5,
		CreatorId:       -1,
		CreatorUsername: quaFile.Creator,
		GameMode:        quaFile.Mode,
		RankedStatus:    enums.RankedStatusNotSubmitted,
		Artist:          quaFile.Artist,
		Title:           quaFile.Title,
		Source:          quaFile.Source,
		Tags:            quaFile.Tags,
		Description:     quaFile.Description,
		DifficultyName:  quaFile.DifficultyName,
	}

	if err := db.InsertMap(dbMap); err != nil {
		return APIErrorServerError("Error inserting map into db", err)
	}

	// Compress File
	var gzipBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipBuffer)

	if _, err = gzipWriter.Write(fileBytes); err != nil {
		_ = gzipWriter.Close()
		return APIErrorServerError("Failed to gzip compress file", err)
	}

	_ = gzipWriter.Close()

	// Upload File
	if err := azure.Client.UploadFile("maps", fmt.Sprintf("%v.qua", dbMap.Id), gzipBuffer.Bytes()); err != nil {
		return APIErrorServerError("Failed to upload unsubmitted map", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your map has been successfully uploaded"})
	return nil
}
