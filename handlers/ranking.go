package handlers

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/webhooks"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

// GetRankingQueueConfig Returns the vote/denial configuration for the ranking queue
func GetRankingQueueConfig(c *gin.Context) *APIError {
	c.JSON(http.StatusOK, gin.H{
		"votes_required":          config.Instance.RankingQueue.VotesRequired,
		"denials_required":        config.Instance.RankingQueue.DenialsRequired,
		"mapset_uploads_required": config.Instance.RankingQueue.MapsetUploadsRequired,
		"resubmission_days":       config.Instance.RankingQueue.ResubmissionDays,
	})

	return nil
}

// GetRankingQueue Retrieves the ranking queue for a given mode
// Endpoint: /v2/ranking/queue/mode/:mode
func GetRankingQueue(c *gin.Context) *APIError {
	mode, err := strconv.Atoi(c.Param("mode"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid mode")
	}

	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	rankingQueue, err := db.GetRankingQueue(enums.GameMode(mode), 20, page)

	if err != nil {
		return APIErrorServerError("Error retrieving ranking queue", err)
	}

	c.JSON(http.StatusOK, gin.H{"ranking_queue": rankingQueue})
	return nil
}

// SubmitMapsetToRankingQueue Submits a mapsets to the ranking queue
// Endpoint: POST /v2/ranking/queue/:id/submit
func SubmitMapsetToRankingQueue(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid mapset id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	mapset, err := db.GetMapsetById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving mapset from db", err)
	}

	if mapset == nil {
		return APIErrorNotFound("Mapset")
	}

	if mapset.CreatorID != user.Id {
		return APIErrorForbidden("You do not own this mapset and cannot submit it to the queue.")
	}

	if len(mapset.Maps) > 0 && mapset.Maps[0].RankedStatus == enums.RankedStatusRanked {
		return APIErrorForbidden("This mapset is already ranked.")
	}

	existingQueueMapset, err := db.GetRankingQueueMapset(mapset.Id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving existing mapset in the ranking queue", err)
	}

	uploadedMapsets, err := db.GetUserMapsets(user.Id)

	if err != nil {
		return APIErrorServerError("Error retrieving user uploaded mapsets from db", err)
	}

	uploadsRequired := config.Instance.RankingQueue.MapsetUploadsRequired

	if len(uploadedMapsets) < uploadsRequired {
		return APIErrorForbidden(fmt.Sprintf("You must have at least %v uploaded mapsets to submit to the queue", uploadsRequired))
	}

	isEligible, err := isEligibleToSubmitToRankingQueue(user.Id, uploadedMapsets)

	if err != nil {
		return APIErrorServerError("Error checking user queue submission eligibility", err)
	}

	if !isEligible {
		return APIErrorForbidden("You cannot submit any more mapsets to the queue at this time.")
	}

	if existingQueueMapset == nil {
		return addMapsetToRankingQueue(c, mapset)
	} else {
		return resubmitMapsetToRankingQueue(c, existingQueueMapset)
	}
}

// RemoveFromRankingQueue Allows a user to remove their mapset from the ranking queue (self deny)
// Endpoint: POST /v2/ranking/queue/:id/remove
func RemoveFromRankingQueue(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid mapset id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	queueMapset, err := db.GetRankingQueueMapset(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving ranking queue mapset", err)
	}

	if queueMapset == nil {
		return APIErrorNotFound("Mapset")
	}

	if queueMapset.Mapset.CreatorID != user.Id {
		return APIErrorForbidden("You do not own this mapset.")
	}

	if queueMapset.Mapset.Maps[0].RankedStatus == enums.RankedStatusRanked {
		return APIErrorForbidden("This mapset is already ranked.")
	}

	if queueMapset.Status == db.RankingQueueDenied || queueMapset.Status == db.RankingQueueBlacklisted {
		return APIErrorForbidden("This mapset has already been denied or blacklisted from the queue.")
	}

	denyAction := &db.MapsetRankingQueueComment{
		UserId:     user.Id,
		MapsetId:   queueMapset.MapsetId,
		ActionType: db.RankingQueueActionDeny,
		IsActive:   true,
		Comment:    "I have just removed my own mapset from the ranking queue.",
	}

	if err := denyAction.Insert(); err != nil {
		return APIErrorServerError("Error inserting new ranking queue denial", err)
	}

	if err := queueMapset.UpdateStatus(db.RankingQueueDenied); err != nil {
		return APIErrorServerError("Error updating ranking queue mapset status", err)
	}

	_ = webhooks.SendQueueWebhook(user, queueMapset.Mapset, db.RankingQueueActionDeny)
	c.JSON(http.StatusOK, gin.H{"message": "You have successfully removed your mapset from the ranking queue."})
	return nil
}

// Adds a new mapset to the ranking queue
func addMapsetToRankingQueue(c *gin.Context, mapset *db.Mapset) *APIError {
	rankingQueueMapset := &db.RankingQueueMapset{
		MapsetId:        mapset.Id,
		Timestamp:       time.Now().UnixMilli(),
		DateLastUpdated: time.Now().UnixMilli(),
		Status:          db.RankingQueuePending,
	}

	if err := rankingQueueMapset.Insert(); err != nil {
		return APIErrorServerError("Error inserting mapset into ranking queue", err)
	}

	comment := &db.MapsetRankingQueueComment{
		UserId:     mapset.CreatorID,
		MapsetId:   mapset.Id,
		ActionType: db.RankingQueueActionComment,
		Comment:    "I have just submitted my mapset to the ranking queue!",
	}

	if err := comment.Insert(); err != nil {
		return APIErrorServerError("Error inserting ranking queue comment", err)
	}

	_ = webhooks.SendQueueSubmitWebhook(getAuthedUser(c), mapset)
	c.JSON(http.StatusOK, gin.H{"message": "Your mapset was successfully submitted to the queue"})
	return nil
}

// Resubmits a mapset to the ranking queue
func resubmitMapsetToRankingQueue(c *gin.Context, mapset *db.RankingQueueMapset) *APIError {
	switch mapset.Status {
	case db.RankingQueuePending, db.RankingQueueOnHold, db.RankingQueueResolved:
		return APIErrorForbidden("Your mapset is already pending for rank.")
	case db.RankingQueueBlacklisted:
		return APIErrorForbidden("Your mapset has been blacklisted from the ranking queue.")
	case db.RankingQueueRanked:
		return APIErrorForbidden("Your mapset is already ranked.")
	}

	// Only if mapset is denied will it get to this point.

	resubmitDays := config.Instance.RankingQueue.ResubmissionDays
	timeSinceDenied := time.Now().Sub(time.UnixMilli(mapset.DateLastUpdated))

	if timeSinceDenied.Hours() < 24*float64(resubmitDays) {
		return APIErrorForbidden(fmt.Sprintf("You can only resubmit your mapset for rank every %v days.", resubmitDays))
	}

	if err := mapset.UpdateVoteCount(0); err != nil {
		return APIErrorServerError("Error updating vote count for ranking queue mapset", err)
	}

	if err := mapset.UpdateStatus(db.RankingQueuePending); err != nil {
		return APIErrorServerError("Error updating ranking queue mapset status", err)
	}

	// Deactivate previous actions since they no longer count
	if err := db.DeactivateRankingQueueActions(mapset.MapsetId); err != nil {
		return APIErrorServerError("Error deactivating previous ranking queue actions", err)
	}

	comment := &db.MapsetRankingQueueComment{
		UserId:     mapset.Mapset.CreatorID,
		MapsetId:   mapset.Id,
		ActionType: db.RankingQueueActionComment,
		Comment:    "I have just resubmitted my mapset to the ranking queue!",
	}

	if err := comment.Insert(); err != nil {
		return APIErrorServerError("Error inserting ranking queue comment", err)
	}

	_ = webhooks.SendQueueSubmitWebhook(getAuthedUser(c), mapset.Mapset)
	c.JSON(http.StatusOK, gin.H{"message": "Your mapset has been resubmitted for rank"})
	return nil
}

// Returns if the user is eligible to submit to the ranking queue
func isEligibleToSubmitToRankingQueue(userId int, uploadedMapsets []*db.Mapset) (bool, error) {
	numRanked := 0

	for _, mapset := range uploadedMapsets {
		if len(mapset.Maps) > 0 && mapset.Maps[0].RankedStatus == enums.RankedStatusRanked {
			numRanked++
		}
	}

	maxSubmissions := 1

	if numRanked >= 5 {
		maxSubmissions = 3
	} else if numRanked >= 2 {
		maxSubmissions = 2
	}

	pendingMapsets, err := db.GetUserMapsetsInRankingQueue(userId)

	if err != nil {
		return false, err
	}

	return len(pendingMapsets) < maxSubmissions, nil
}
