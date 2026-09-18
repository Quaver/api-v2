package handlers

import (
	"errors"
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/downloadlimit"
	"github.com/Quaver/api2/files"
	"github.com/Quaver/api2/tools"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
	"time"
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

	if isHeadRequest(c) {
		return setFileContentLength(c, path)
	}

	reservation, apiErr := reserveMapsetDownloadQuota(c, user.Id, path)

	if apiErr != nil {
		return apiErr
	}

	if err := db.InsertMapsetDownload(&db.MapsetDownload{
		UserId:    user.Id,
		MapsetId:  mapset.Id,
		Timestamp: time.Now().UnixMilli(),
	}); err != nil {
		if reservation != nil {
			if releaseErr := downloadlimit.Release(db.RedisCtx, db.Redis, reservation); releaseErr != nil {
				logrus.Errorf("Error releasing download quota for user %d: %v", user.Id, releaseErr)
			}
		}

		return APIErrorServerError("Error inserting mapset download into db", err)
	}

	c.FileAttachment(path, fmt.Sprintf("%v.qp", mapset.Id))
	return nil
}

// DownloadReplay Handles the downloading of a replay file
// Endpoint: GET /v2/download/replay/:id
func DownloadReplay(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	score, err := db.GetScoreById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error getting score by id", err)
	}

	if score == nil || (score.Failed && score.TournamentGameId == nil) {
		return APIErrorNotFound("Replay")
	}

	user, err := db.GetUserById(score.UserId)

	if err != nil {
		return APIErrorServerError("Error getting user in DB from score", err)
	}

	path, err := files.CacheReplay(id)

	if err != nil {
		return APIErrorNotFound("Replay")
	}

	tempPath := fmt.Sprintf("%v/replay-%v.qr", files.GetTempDirectory(), time.Now().UnixMilli())

	if err := tools.BuildReplay(user, score, path, tempPath); err != nil {
		return APIErrorServerError("Error building replay", err)
	}

	c.FileAttachment(tempPath, fmt.Sprintf("%v.qr", score.Id))

	if err := os.Remove(tempPath); err != nil {
		return nil
	}

	return nil
}

// DownloadMultiplayerMapset Serves the multiplayer mapset that was shared
// Endpoint: GET /v2/download/multiplayer/:id
func DownloadMultiplayerMapset(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	path := fmt.Sprintf("%v/multiplayer/%v.qp", files.GetTempDirectory(), id)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return APIErrorNotFound("Multiplayer Mapset")
	}

	if isHeadRequest(c) {
		return setFileContentLength(c, path)
	}

	c.FileAttachment(path, fmt.Sprintf("multiplayer_%v.qp", id))
	return nil
}

// Check if request is HEAD
func isHeadRequest(c *gin.Context) bool {
	return c.Request.Method == "HEAD"
}

func setFileContentLength(c *gin.Context, path string) *APIError {
	fileInfo, err := os.Stat(path)

	if err != nil {
		return APIErrorServerError("Error getting file information", err)
	}

	// Set Content-Length header
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	return nil
}

func reserveMapsetDownloadQuota(c *gin.Context, userID int, path string) (*downloadlimit.Reservation, *APIError) {
	fileInfo, err := os.Stat(path)

	if err != nil {
		return nil, APIErrorServerError("Error getting mapset file information", err)
	}

	byteCount, err := mapsetResponseByteCount(c, path, fileInfo)

	if err != nil {
		return nil, APIErrorServerError("Error determining mapset response size", err)
	}

	if byteCount == 0 {
		return nil, nil
	}

	reservation, allowed, err := downloadlimit.TryReserve(c.Request.Context(), db.Redis, userID, byteCount)

	if err != nil {
		return nil, APIErrorServerError("Error checking mapset download rate limit", err)
	}

	if !allowed {
		return nil, &APIError{
			Status:  http.StatusTooManyRequests,
			Message: "Download rate limit has been reached",
		}
	}

	return reservation, nil
}

// mapsetResponseByteCount asks net/http to evaluate the request as a HEAD
// response so quota accounting follows the same range and precondition rules as
// FileAttachment without reading the response body.
func mapsetResponseByteCount(c *gin.Context, path string, fileInfo os.FileInfo) (int64, error) {
	file, err := os.Open(path)

	if err != nil {
		return 0, err
	}

	defer file.Close()

	request := c.Request.Clone(c.Request.Context())
	request.Method = http.MethodHead
	response := &responseMetadataWriter{header: c.Writer.Header().Clone()}
	http.ServeContent(response, request, fileInfo.Name(), fileInfo.ModTime(), file)

	if response.status != http.StatusOK && response.status != http.StatusPartialContent {
		return 0, nil
	}

	contentLength := response.header.Get("Content-Length")

	if contentLength == "" && response.status == http.StatusOK {
		return fileInfo.Size(), nil
	}

	byteCount, err := strconv.ParseInt(contentLength, 10, 64)

	if err != nil || byteCount < 0 {
		return 0, fmt.Errorf("invalid response content length %q", contentLength)
	}

	return byteCount, nil
}

type responseMetadataWriter struct {
	header http.Header
	status int
}

func (w *responseMetadataWriter) Header() http.Header {
	return w.header
}

func (w *responseMetadataWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}

func (w *responseMetadataWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}

	return len(body), nil
}
