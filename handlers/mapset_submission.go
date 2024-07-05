package handlers

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/qua"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path"
	"slices"
	"strings"
)

var (
	acceptedFileExtensions = []string{".mp3", ".qua", ".ogg", ".png", ".jpg", ".jpeg", ".wav", ".lua"}

	acceptedMimeTypes = []string{
		"audio/mpeg", "audio/ogg", "audio/wav", "audio/wave", "audio/x-pn-wav",
		"audio/x-wav", "audio/vnd.wave", "image/png", "image/jpeg",
		"text/plain; charset=utf-8",
	}
)

const (
	__MACOSX string = "__MACOSX"
)

// HandleMapsetSubmission Handles the uploading/updating of a mapset archive (.qp) file
// Endpoint: POST /v2/mapset
func HandleMapsetSubmission(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	zipReader, apiErr := checkValidRequestMapset(c, user)

	if apiErr != nil {
		return apiErr
	}

	if apiErr := validateMapsetZipFiles(zipReader); apiErr != nil {
		return apiErr
	}

	quaFiles, apiErr := readQuaFilesFromZip(zipReader)

	if apiErr != nil {
		return apiErr
	}

	fmt.Println(quaFiles)
	c.JSON(http.StatusOK, gin.H{"message": "Your mapset has been successfully uploaded."})
	return nil
}

// Checks if the request contains a valid .qp file
func checkValidRequestMapset(c *gin.Context, user *db.User) (*zip.Reader, *APIError) {
	fileHeader, _ := c.FormFile("mapset")

	if fileHeader == nil {
		return nil, APIErrorBadRequest("You must provide a `mapset` file.")
	}

	if !strings.HasSuffix(fileHeader.Filename, ".qp") {
		return nil, APIErrorBadRequest("Your mapset file must be a valid .qp archive.")
	}

	file, err := fileHeader.Open()

	if err != nil {
		return nil, APIErrorServerError("Error opening file: ", err)
	}

	defer file.Close()

	var fileBytes = make([]byte, fileHeader.Size)

	if _, err = file.Read(fileBytes); err != nil {
		return nil, APIErrorServerError("Error reading file", err)
	}

	// Check File Size Limit
	fileSizeLimitMB := getMapsetFileSizeLimitMB(user)
	limitBytes := fileSizeLimitMB * 1_048_576

	if int64(len(fileBytes)) > limitBytes || fileHeader.Size > limitBytes {
		errMsg := fmt.Sprintf("The file you have uploaded must not exceed %v MB.", fileSizeLimitMB)
		return nil, APIErrorBadRequest(errMsg)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))

	if err != nil {
		logrus.Error("Error reading zip file: ", err)
		return nil, APIErrorBadRequest("The file you have provided was not a valid zip archive.")
	}

	return zipReader, nil
}

// Validates the mapset files that are in a zip archive.
func validateMapsetZipFiles(zip *zip.Reader) *APIError {
	hasAtleastOneQua := false

	for _, file := range zip.File {
		if strings.Contains(file.Name, __MACOSX) || strings.Contains(strings.ToLower(file.Name), "thumbs.db") {
			continue
		}

		extension := strings.ToLower(path.Ext(file.Name))
		invalidErr := APIErrorBadRequest(fmt.Sprintf("Your mapset contains an invalid file: %v", file.Name))

		if !slices.Contains(acceptedFileExtensions, extension) {
			return invalidErr
		}

		if err := validateMimetype(file); err != nil {
			logrus.Errorf("Error detecting mimetype of file %v - %v", file.Name, err)
			return invalidErr
		}

		if extension == ".qua" {
			hasAtleastOneQua = true
		}
	}

	if !hasAtleastOneQua {
		return APIErrorBadRequest("Your mapset archive must contain at least one .qua file.")
	}

	return nil
}

// Detects the mimetype of a file
func validateMimetype(file *zip.File) error {
	reader, err := file.Open()

	if err != nil {
		return err
	}

	defer reader.Close()

	extension := strings.ToLower(path.Ext(file.Name))

	switch extension {
	case ".jpeg", ".jpg", ".png":
		_, format, err := image.Decode(reader)

		if err != nil || format != "jpeg" && format != "png" {
			return err
		}
	default:
		mime, err := mimetype.DetectReader(reader)

		if err != nil {
			return err
		}

		if !slices.Contains(acceptedMimeTypes, mime.String()) {
			return errors.New(fmt.Sprintf("File %v has invalid mimetype: %v", file.Name, mime.String()))
		}
	}

	return nil
}

// Reads all .qua files from a mapset archive
func readQuaFilesFromZip(archive *zip.Reader) (map[*zip.File]*qua.Qua, *APIError) {
	quaFiles := map[*zip.File]*qua.Qua{}

	for _, file := range archive.File {
		if strings.Contains(file.Name, __MACOSX) || strings.ToLower(path.Ext(file.Name)) != ".qua" {
			continue
		}

		reader, err := file.Open()

		if err != nil {
			logrus.Error("Error opening file: ", file.Name, err)
			return nil, APIErrorBadRequest(fmt.Sprintf("Error reading file: %v", file.Name))
		}

		fileBytes, err := io.ReadAll(reader)
		reader.Close()

		if err != nil {
			logrus.Error("Error reading file: ", file.Name, err)
			return nil, APIErrorBadRequest(fmt.Sprintf("Error reading file: %v", file.Name))
		}

		quaFile, err := qua.Parse(fileBytes)

		if err != nil {
			logrus.Error("Error parsing qua file: ", file.Name, err)
			return nil, APIErrorBadRequest(fmt.Sprintf("Error reading file: %v", file.Name))
		}

		quaFiles[file] = quaFile
	}

	return quaFiles, nil
}

// Returns a mapsets file size limit in MB
func getMapsetFileSizeLimitMB(user *db.User) int64 {
	if enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		return 100
	}

	return 50
}
