package handlers

import (
	"bytes"
	"fmt"
	"github.com/Quaver/api2/azure"
	"github.com/gin-gonic/gin"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
)

// UploadClanAvatar Handles the uploading of a clan avatar
// Endpoint: /v2/clan/avatar
func UploadClanAvatar(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	clan, apiErr := getClanAndCheckOwnership(user, *user.ClanId)

	if apiErr != nil {
		return apiErr
	}

	fileBytes, apiErr := validateUploadedClanImage(c)

	if apiErr != nil {
		return apiErr
	}

	if err := azure.Client.UploadFile("clan-avatars", fmt.Sprintf("%v.jpg", clan.Id), fileBytes); err != nil {
		return APIErrorServerError("Failed to upload clan avatar", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your clan avatar has been successfully uploaded."})
	return nil
}

// UploadClanBanner Handles the uploading of a clan banner
// Endpoint: /v2/clan/banner
func UploadClanBanner(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	clan, apiErr := getClanAndCheckOwnership(user, *user.ClanId)

	if apiErr != nil {
		return apiErr
	}

	fileBytes, apiErr := validateUploadedClanImage(c)

	if apiErr != nil {
		return apiErr
	}

	if err := azure.Client.UploadFile("clan-banners", fmt.Sprintf("%v.jpg", clan.Id), fileBytes); err != nil {
		return APIErrorServerError("Failed to upload clan banner", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your clan banner has been successfully uploaded."})
	return nil
}

// Validates and returns an uploaded clan image.
// - Must be a JPEG/PNG file
// - Must be 1MB or under
func validateUploadedClanImage(c *gin.Context) ([]byte, *APIError) {
	fileHeader, _ := c.FormFile("image")

	if fileHeader == nil {
		return nil, APIErrorBadRequest("You must provide a valid `image` file.")
	}

	file, err := fileHeader.Open()

	if err != nil {
		return nil, APIErrorServerError("Error opening file", err)
	}

	defer file.Close()

	// Read File
	var fileBytes = make([]byte, fileHeader.Size)

	if _, err = file.Read(fileBytes); err != nil {
		return nil, APIErrorServerError("Error reading file", err)
	}

	// Check File Size Limit
	const fileSizeLimit = 1048576

	if len(fileBytes) > fileSizeLimit || fileHeader.Size > fileSizeLimit {
		return nil, APIErrorBadRequest("The file you have uploaded is too large. You must not exceed 1MB.")
	}

	// Check image type
	_, format, err := image.Decode(bytes.NewReader(fileBytes))

	if err != nil || format != "jpeg" && format != "png" {
		return nil, APIErrorBadRequest("The image you have uploaded is not a valid jpeg file.")
	}

	return fileBytes, nil
}
