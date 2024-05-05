package handlers

import (
	"fmt"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/files"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// UploadMultiplayerMapset Uploads a multiplayer map to be shared.
// Endpoint: POST /v2/download/multiplayer/:id
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

	if apiErr := validateUploadedMultiplayerMapset(c, id); apiErr != nil {
		return apiErr
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your mapset file was successfully uploaded and shared."})
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

	path := fmt.Sprintf("%v/%v.qp", files.GetTempDirectory(), gameId)

	if err := os.WriteFile(path, fileBytes, 0644); err != nil {
		return APIErrorServerError("Error writing file to temp directory", err)
	}

	return nil
}
