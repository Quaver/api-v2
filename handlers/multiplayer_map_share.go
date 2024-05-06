package handlers

import (
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/files"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// UploadMultiplayerMapset Uploads a multiplayer map to be shared.
// Endpoint: POST /v2/download/multiplayer/:id/upload
func UploadMultiplayerMapset(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		return APIErrorForbidden("You do not have permission to access this resource")
	}

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid multiplayer game id")
	}

	body := struct {
		MapMD5     string `form:"map_md5" json:"map_md5" binding:"required"`
		PackageMD5 string `form:"package_md5" json:"package_md5" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if apiErr := validateUploadedMultiplayerMapset(c, id); apiErr != nil {
		return apiErr
	}

	mapShare := &db.MultiplayerMapShare{
		UserId:     user.Id,
		GameId:     id,
		MapMD5:     body.MapMD5,
		PackageMD5: body.PackageMD5,
		Timestamp:  time.Now().UnixMilli(),
	}

	if err := mapShare.Insert(); err != nil {
		return APIErrorServerError("Error inserting map share into DB", err)
	}

	if err := mapShare.PublishToRedis(); err != nil {
		return APIErrorServerError("Error publishing map share to redis", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your mapset file was successfully uploaded."})
	return nil
}

// Validates and caches multiplayer mapset
func validateUploadedMultiplayerMapset(c *gin.Context, gameId int) *APIError {
	fileHeader, _ := c.FormFile("mapset")

	if fileHeader == nil {
		return APIErrorBadRequest("You must provide a valid `mapset` file.")
	}

	if !strings.HasSuffix(fileHeader.Filename, ".qp") {
		return APIErrorBadRequest("You must provide a valid .qp file.")
	}

	file, err := fileHeader.Open()

	if err != nil {
		return APIErrorServerError("Error opening file", err)
	}

	defer file.Close()

	// Read File
	var fileBytes = make([]byte, fileHeader.Size)

	if _, err = file.Read(fileBytes); err != nil {
		return APIErrorServerError("Error reading file", err)
	}

	// Check File Size Limit
	const fileSizeLimit = 1048576 * 50

	if len(fileBytes) > fileSizeLimit || fileHeader.Size > fileSizeLimit {
		return APIErrorBadRequest("The file you have uploaded must not exceed 5MB.")
	}

	// TODO: Validate zip & containing files (IMPORTANT!)

	path := fmt.Sprintf("%v/multiplayer/%v.qp", files.GetTempDirectory(), gameId)

	if err := os.WriteFile(path, fileBytes, 0644); err != nil {
		return APIErrorServerError("Error writing file to temp directory", err)
	}

	return nil
}
