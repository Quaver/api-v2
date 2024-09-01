package handlers

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/cache"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/files"
	"github.com/Quaver/api2/qua"
	"github.com/Quaver/api2/sliceutil"
	"github.com/Quaver/api2/tools"
	v1 "github.com/Quaver/api2/v1"
	"github.com/Quaver/api2/webhooks"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/oliamb/cutter"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
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

	errBannerFileNoExists       string = "could not create mapset banner (file no exists)"
	errAudioPreviewFileNoExists string = "could not create audio preview (file no exists)"
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

	if apiErr := validateQuaFiles(user, quaFiles); apiErr != nil {
		return apiErr
	}

	// If all qua files contain -1 for the mapset id and map id, then we're uploading a new set.
	isUploadingNewMapset := sliceutil.All(sliceutil.Values(quaFiles), func(q *qua.Qua) bool {
		return q.MapSetId == -1 && q.MapId == -1
	})

	if apiErr := checkDuplicateQuaData(quaFiles); apiErr != nil {
		return apiErr
	}

	var mapset *db.Mapset

	if isUploadingNewMapset {
		mapset, apiErr = uploadNewMapset(user, quaFiles)
	} else {
		mapset, apiErr = updateExistingMapset(user, quaFiles)
	}

	if apiErr != nil {
		return apiErr
	}

	archive, err := createMapsetArchive(zipReader, quaFiles)

	if err != nil {
		return APIErrorServerError("Failed to create mapset archive", err)
	}

	mapset.PackageMD5 = files.GetByteSliceMD5(archive)

	if err := db.UpdateMapsetPackageMD5(mapset.Id, mapset.PackageMD5); err != nil {
		return APIErrorServerError("Error updating mapset package md5", err)
	}

	if err := azure.Client.UploadFile("mapsets", fmt.Sprintf("%v.qp", mapset.Id), archive); err != nil {
		return APIErrorServerError("Failed to upload mapset archive to azure", err)
	}

	if err := db.IndexElasticSearchMapset(*mapset); err != nil {
		return APIErrorServerError("Error updating elastic search", err)
	}

	if err := v1.UpdateElasticSearchMapset(mapset.Id); err != nil {
		logrus.Error(err)
	}

	if apiErr := resolveMapsetInRankingQueue(user, mapset); apiErr != nil {
		return apiErr
	}

	go func() {
		if err := createMapsetBanner(zipReader, quaFiles); err != nil && err.Error() != errBannerFileNoExists {
			logrus.Warning("Error creating mapset banner: ", err)
		}

		if err := createAudioPreviewFromZip(zipReader, quaFiles); err != nil && err.Error() != errAudioPreviewFileNoExists {
			logrus.Warning("Error creating audio file: ", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Your mapset has been successfully uploaded.",
		"mapset":  mapset,
	})

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

// Goes through a map of qua files and makes sure they are valid
func validateQuaFiles(user *db.User, quaFiles map[*zip.File]*qua.Qua) *APIError {
	for _, quaFile := range quaFiles {
		if quaFile.Artist == "" || quaFile.Title == "" || quaFile.DifficultyName == "" || quaFile.Creator == "" {
			return APIErrorBadRequest("Your .qua files must contain filled in metadata.")
		}

		if quaFile.Creator != user.Username {
			return APIErrorBadRequest("The username in your .qua files must match your username.")
		}
	}

	return nil
}

// Handles the uploading of a brand new mapset
func uploadNewMapset(user *db.User, quaFiles map[*zip.File]*qua.Qua) (*db.Mapset, *APIError) {
	if apiErr := checkUserUploadEligibility(user); apiErr != nil {
		return nil, apiErr
	}

	referenceMap := sliceutil.Values(quaFiles)[0]

	mapset := &db.Mapset{
		CreatorID:       user.Id,
		CreatorUsername: user.Username,
		Artist:          referenceMap.Artist,
		Title:           referenceMap.Title,
		Source:          referenceMap.Source,
		Tags:            referenceMap.Tags,
	}

	if err := mapset.Insert(); err != nil {
		return nil, APIErrorServerError("Error inserting mapset into db", err)
	}

	for _, quaFile := range quaFiles {
		songMap, apiErr := InsertOrUpdateMap(user, mapset, quaFile)

		if apiErr != nil {
			return nil, apiErr
		}

		mapset.Maps = append(mapset.Maps, songMap)
	}

	if err := db.AddUserActivity(user.Id, db.UserActivityUploadedMapset, mapset.String(), mapset.Id); err != nil {
		return nil, APIErrorServerError("Error inserting user activity for uploading mapset", err)
	}

	if err := db.IncrementTotalMapsetCount(); err != nil {
		return nil, APIErrorServerError("Error increment total mapset count in redis", err)
	}

	return mapset, nil
}

// Handles the updating of an existing mapset
func updateExistingMapset(user *db.User, quaFiles map[*zip.File]*qua.Qua) (*db.Mapset, *APIError) {
	quaSlice := sliceutil.Values(quaFiles)

	if !sliceutil.All(quaSlice, func(q *qua.Qua) bool {
		return q.MapSetId == quaSlice[0].MapSetId
	}) {
		return nil, APIErrorBadRequest("Your .qua files have conflicting `MapSetId`s.")
	}

	mapset, err := db.GetMapsetById(quaSlice[0].MapSetId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, APIErrorServerError("Error retrieving mapset from database during update", err)
	}

	if mapset == nil || len(mapset.Maps) == 0 {
		return nil, APIErrorBadRequest("The mapset you are trying to update does not exist.")
	}

	if mapset.CreatorID != user.Id {
		return nil, APIErrorForbidden("You cannot update a mapset that you do not own.")
	}

	if mapset.Maps[0].RankedStatus == enums.RankedStatusRanked {
		return nil, APIErrorBadRequest("You cannot update an already ranked mapset.")
	}

	// Check to see if all non -1 files actually exist in the mapset.
	for _, quaFile := range quaFiles {
		if quaFile.MapId == -1 {
			continue
		}

		if !slices.ContainsFunc(mapset.Maps, func(mapQua *db.MapQua) bool {
			return quaFile.MapId == mapQua.Id
		}) {
			return nil, APIErrorBadRequest("One of your .qua files has a non -1 MapId that does not exist.")
		}
	}

	var newMaps []*db.MapQua

	// Insert new maps & update existing ones
	for _, quaFile := range quaFiles {
		songMap, apiErr := InsertOrUpdateMap(user, mapset, quaFile)

		if apiErr != nil {
			return nil, apiErr
		}

		newMaps = append(newMaps, songMap)
	}

	// Delete old maps that are no longer in the set.
	for _, mapQua := range mapset.Maps {
		if !slices.ContainsFunc(quaSlice, func(q *qua.Qua) bool {
			return q.MapId == mapQua.Id
		}) {
			if err := db.DeleteMap(mapQua.Id); err != nil {
				return nil, APIErrorServerError("Error deleting map from database", err)
			}
		}
	}

	mapset.CreatorUsername = user.Username
	mapset.Artist = quaSlice[0].Artist
	mapset.Title = quaSlice[0].Title
	mapset.Source = quaSlice[0].Source
	mapset.Tags = quaSlice[0].Tags
	mapset.User = nil
	mapset.Maps = newMaps

	if err = mapset.UpdateMetadata(); err != nil {
		return nil, APIErrorServerError("Error updating mapset metadata", err)
	}

	if err := db.AddUserActivity(user.Id, db.UserActivityUpdatedMapset, mapset.String(), mapset.Id); err != nil {
		return nil, APIErrorServerError("Error inserting user activity for updating mapset", err)
	}

	return mapset, nil
}

// InsertOrUpdateMap Inserts/Updates a map in the database
func InsertOrUpdateMap(user *db.User, mapset *db.Mapset, quaFile *qua.Qua) (*db.MapQua, *APIError) {
	songMap := &db.MapQua{
		MapsetId:             mapset.Id,
		CreatorId:            user.Id,
		CreatorUsername:      user.Username,
		GameMode:             quaFile.Mode,
		RankedStatus:         enums.RankedStatusUnranked,
		Artist:               quaFile.Artist,
		Title:                quaFile.Title,
		Source:               quaFile.Source,
		Tags:                 quaFile.Tags,
		Description:          quaFile.Description,
		DifficultyName:       quaFile.DifficultyName,
		Length:               quaFile.MapLength(),
		BPM:                  quaFile.CommonBPM(),
		CountHitObjectNormal: quaFile.CountHitObjectNormal(),
		CountHitObjectLong:   quaFile.CountHitObjectLong(),
		MaxCombo:             quaFile.MaxCombo(),
	}

	if quaFile.MapId != -1 {
		songMap.Id = quaFile.MapId
	}

	if err := db.SQL.Save(&songMap).Error; err != nil {
		return nil, APIErrorServerError("Error inserting map into db", err)
	}

	quaFile.ReplaceIds(mapset.Id, songMap.Id)
	songMap.MD5 = files.GetByteSliceMD5(quaFile.RawBytes)

	if err := db.UpdateMapMD5(songMap.Id, songMap.MD5); err != nil {
		return nil, APIErrorServerError("Error saving map in db", err)
	}

	filePath := fmt.Sprintf("%v/%v.qua", files.GetTempDirectory(), songMap.Id)

	if err := quaFile.Write(filePath); err != nil {
		return nil, APIErrorServerError("Error writing .qua file to disk", err)
	}

	if err := azure.Client.UploadFile("maps", quaFile.FileName(), quaFile.RawBytes); err != nil {
		return nil, APIErrorServerError("Error uploading .qua file to azure", err)
	}

	go func() {
		calcMapDifficulty(songMap, filePath)

		if err := os.Remove(filePath); err != nil {
			logrus.Error("Error removing file: ", filePath)
		}
	}()

	return songMap, nil
}

// Checks if a user is eligible to upload an existing mapset
func checkUserUploadEligibility(user *db.User) *APIError {
	mapsets, err := db.GetUserMonthlyUploadMapsets(user.Id)

	if err != nil {
		return APIErrorServerError("Error retrieving user monthly mapset uploads in db", err)
	}

	maxUploads := getUserMaxUploadsPerMonth(user)

	if len(mapsets) >= maxUploads {
		return APIErrorForbidden(fmt.Sprintf("You can only upload %v mapsets per month.", maxUploads))
	}

	return nil
}

// Checks if .qua files in a mapset have duplicate map ids or difficulty names
func checkDuplicateQuaData(quaFiles map[*zip.File]*qua.Qua) *APIError {
	var duplicateMapIds []int
	var duplicateDifficultyNames []string

	for _, quaFile := range quaFiles {
		if slices.Contains(duplicateMapIds, quaFile.MapId) {
			return APIErrorBadRequest("Your .qua files have duplicate `MapId`s.")
		}

		if slices.Contains(duplicateDifficultyNames, quaFile.DifficultyName) {
			return APIErrorBadRequest("Your .qua files have duplicate `DifficultyName`s.")
		}

		if quaFile.MapId != -1 {
			duplicateMapIds = append(duplicateMapIds, quaFile.MapId)
		}

		duplicateDifficultyNames = append(duplicateDifficultyNames, quaFile.DifficultyName)
	}

	return nil
}

// Returns a mapsets file size limit in MB
func getMapsetFileSizeLimitMB(user *db.User) int64 {
	if enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		return 100
	}

	return 50
}

// Gets the maximum amount of mapsets a user can upload per month
func getUserMaxUploadsPerMonth(user *db.User) int {
	if enums.HasUserGroup(user.UserGroups, enums.UserGroupSwan) ||
		enums.HasUserGroup(user.UserGroups, enums.UserGroupDeveloper) ||
		enums.HasUserGroup(user.UserGroups, enums.UserGroupContributor) {
		return math.MaxInt32
	} else if enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		return 20
	} else {
		return 10
	}
}

// Calculates a map's difficulty rating
func calcMapDifficulty(songMap *db.MapQua, filePath string) {
	calc, err := tools.RunDifficultyCalculator(filePath, 0)

	if err != nil {
		logrus.Error("Error calculating difficulty for map: ", err)
		return
	}

	if err := db.UpdateMapDifficultyRating(songMap.Id, calc.Difficulty.OverallDifficulty); err != nil {
		logrus.Error("Error updating map difficulty rating in DB: ", err)
		return
	}
}

// Creates a mapset archive file (.qp)
func createMapsetArchive(zipReader *zip.Reader, quaFiles map[*zip.File]*qua.Qua) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add all files that aren't .qua from the original package into this one.
	for _, zipFile := range zipReader.File {
		if strings.Contains(zipFile.Name, __MACOSX) ||
			strings.Contains(strings.ToLower(zipFile.Name), "thumbs.db") ||
			path.Ext(zipFile.Name) == ".qua" {
			continue
		}

		newFile, err := zipWriter.Create(zipFile.Name)

		if err != nil {
			return nil, err
		}

		reader, err := zipFile.Open()

		if err != nil {
			return nil, err
		}

		fileBytes, err := io.ReadAll(reader)

		if err != nil {
			return nil, err
		}

		reader.Close()

		if _, err = newFile.Write(fileBytes); err != nil {
			return nil, err
		}
	}

	// Now add .qua files to the archive.
	for _, quaFile := range quaFiles {
		newFile, err := zipWriter.Create(quaFile.FileName())

		if err != nil {
			return nil, err
		}

		if _, err = newFile.Write(quaFile.RawBytes); err != nil {
			return nil, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Creates an auto-cropped mapset banner and uploads it to azure
func createMapsetBanner(zip *zip.Reader, quaFiles map[*zip.File]*qua.Qua) error {
	// Loop through each qua file & try to find a matching background file in the archive.
	// Both need to exist in order for a banner to get created.
	for _, quaFile := range quaFiles {
		for _, zipFile := range zip.File {
			if strings.Contains(zipFile.Name, __MACOSX) || path.Base(zipFile.Name) != quaFile.BackgroundFile {
				continue
			}

			reader, err := zipFile.Open()

			if err != nil {
				return err
			}

			img, _, err := image.Decode(reader)

			if err != nil {
				return err
			}

			reader.Close()

			cropped, err := cutter.Crop(img, cutter.Config{
				Width:   900,
				Height:  250,
				Anchor:  image.Point{X: 0, Y: 0},
				Mode:    cutter.Centered,
				Options: cutter.Copy,
			})

			if err != nil {
				return err
			}

			buf := new(bytes.Buffer)

			if err := jpeg.Encode(buf, cropped, nil); err != nil {
				return err
			}

			fileName := fmt.Sprintf("%v_banner.jpg", quaFile.MapSetId)

			_ = cache.RemoveCacheServerMapsetBanner(quaFile.MapSetId)

			if err := azure.Client.UploadFile("banners", fileName, buf.Bytes()); err != nil {
				return err
			}

			return nil
		}
	}

	return errors.New(errBannerFileNoExists)
}

// Creates an auto-cropped mapset banner and uploads it to azure
func createAudioPreviewFromZip(zip *zip.Reader, quaFiles map[*zip.File]*qua.Qua) error {
	// Loop through each qua file & try to find a matching audio file.
	// Both need to exist in order for a banner to get created.
	for _, quaFile := range quaFiles {
		for _, zipFile := range zip.File {
			if strings.Contains(zipFile.Name, __MACOSX) || path.Base(zipFile.Name) != quaFile.AudioFile {
				continue
			}

			reader, err := zipFile.Open()

			if err != nil {
				return nil
			}

			audioFilePath, err := filepath.Abs(fmt.Sprintf("%v/%v%v", files.GetTempDirectory(),
				time.Now().UnixMilli(), path.Ext(zipFile.Name)))

			if err != nil {
				return err
			}

			outFile, err := os.Create(audioFilePath)

			if err != nil {
				reader.Close()
				return err
			}

			_, err = io.Copy(outFile, reader)

			if err != nil {
				reader.Close()
				_ = outFile.Close()
				return err
			}

			_ = outFile.Close()

			outputPath, err := filepath.Abs(fmt.Sprintf("%v/%v-preview.mp3", files.GetTempDirectory(),
				time.Now().UnixMilli()))

			if err != nil {
				return err
			}

			previewTime := float32(quaFile.SongPreviewTime) / float32(1000)

			if err := createAudioPreviewFromFile(audioFilePath, outputPath, previewTime); err != nil {
				return err
			}

			fileBytes, err := os.ReadFile(outputPath)

			if err != nil {
				return err
			}

			_ = cache.RemoveCacheServerAudioPreview(quaFile.MapSetId)

			if err := azure.Client.UploadFile("audio-previews",
				fmt.Sprintf("%v.mp3", quaFile.MapSetId), fileBytes); err != nil {
				return err
			}

			if err := os.Remove(audioFilePath); err != nil {
				logrus.Error("Error removing original file: ", err)
				return nil
			}

			if err := os.Remove(outputPath); err != nil {
				logrus.Error("Error removing original file: ", err)
				return nil
			}

			return err
		}
	}

	return errors.New(errAudioPreviewFileNoExists)
}

// Uses FFMPEG to create an audio preview from a file path
func createAudioPreviewFromFile(filePath string, outputPath string, previewTimeSeconds float32) error {
	originalOutputPath := outputPath

	if path.Ext(filePath) == ".ogg" {
		outputPath = strings.Replace(outputPath, ".mp3", ".ogg", -1)
	}

	trimCmd := exec.Command("ffmpeg",
		"-i",
		fmt.Sprintf("%v", filePath),
		"-ss",
		fmt.Sprintf("%2f", previewTimeSeconds),
		"-to",
		fmt.Sprintf("%v", previewTimeSeconds+10),
		"-c",
		"copy",
		fmt.Sprintf("%v", outputPath))

	var trimStdout bytes.Buffer
	var trimStderr bytes.Buffer
	trimCmd.Stdout = &trimStdout
	trimCmd.Stderr = &trimStderr

	if err := trimCmd.Run(); err != nil {
		return fmt.Errorf("%v\n\n```%v```", err, trimStderr.String())
	}

	if path.Ext(filePath) != ".ogg" {
		return nil
	}

	// Convert .ogg to mp3
	convertCmd := exec.Command("ffmpeg",
		"-i",
		fmt.Sprintf("%v", outputPath),
		originalOutputPath)

	var convertStdout bytes.Buffer
	var convertStderr bytes.Buffer
	convertCmd.Stdout = &convertStdout
	convertCmd.Stderr = &convertStderr

	if err := convertCmd.Run(); err != nil {
		return fmt.Errorf("%v\n\n```%v```", err, convertStderr.String())
	}

	if err := os.Remove(outputPath); err != nil {
		return err
	}

	return nil
}

// Sets a mapset's ranking queue status to resolved. This is usually done when a user updates their map
// while on hold.
func resolveMapsetInRankingQueue(user *db.User, mapset *db.Mapset) *APIError {
	rankingQueueMapset, err := db.GetRankingQueueMapset(mapset.Id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving mapset in the ranking queue", err)
	}

	if rankingQueueMapset == nil {
		return nil
	}

	if rankingQueueMapset.Status != db.RankingQueueOnHold {
		return nil
	}

	ok, err := downloadAndRunAutomod(mapset)

	if err != nil {
		return APIErrorServerError("Error running automod", err)
	}

	if !ok {
		return nil
	}

	if err := db.DeactivateRankingQueueActions(mapset.Id); err != nil {
		return APIErrorServerError("Error deactivating ranking queue actions", err)
	}

	resolvedAction := &db.MapsetRankingQueueComment{
		UserId:     mapset.CreatorID,
		MapsetId:   mapset.Id,
		ActionType: db.RankingQueueActionResolved,
		IsActive:   true,
		Comment:    "I have just updated my mapset, and its status has been changed back to Resolved.",
	}

	if err := resolvedAction.Insert(); err != nil {
		return APIErrorServerError("Error inserting new ranking queue on hold action.", err)
	}

	if err := db.NewMapsetActionNotification(mapset, resolvedAction).Insert(); err != nil {
		return APIErrorServerError("Error inserting resolve notification", err)
	}

	if err := rankingQueueMapset.UpdateStatus(db.RankingQueueResolved); err != nil {
		return APIErrorServerError("Error updating ranking queue mapset status", err)
	}

	_ = webhooks.SendQueueWebhook(user, mapset, db.RankingQueueActionResolved)
	return nil
}
