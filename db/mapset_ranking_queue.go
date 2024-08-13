package db

import (
	"fmt"
	"github.com/Quaver/api2/enums"
	"gorm.io/gorm"
	"time"
)

type RankingQueueStatus int8

const (
	RankingQueuePending RankingQueueStatus = iota
	RankingQueueDenied
	RankingQueueBlacklisted
	RankingQueueOnHold
	RankingQueueResolved
	RankingQueueRanked
)

type RankingQueueMapset struct {
	Id              int                          `gorm:"column:id; PRIMARY_KEY" json:"id"`
	MapsetId        int                          `gorm:"column:mapset_id" json:"mapset_id"`
	Timestamp       int64                        `gorm:"column:timestamp" json:"-"`
	CreatedAtJSON   time.Time                    `gorm:"-:all" json:"created_at"` // Same value as Timestamp
	DateLastUpdated int64                        `gorm:"column:date_last_updated" json:"-"`
	LastUpdatedJSON time.Time                    `gorm:"-:all" json:"last_updated"` // Same value as DateLastUpdated
	Status          RankingQueueStatus           `gorm:"column:status" json:"status"`
	NeedsAttention  bool                         `gorm:"column:needs_attention" json:"-"`
	VoteCount       int                          `gorm:"-:all" json:"-"`
	Mapset          *Mapset                      `gorm:"foreignKey:MapsetId; references:Id" json:"mapset"`
	Votes           []*MapsetRankingQueueComment `gorm:"-:all" json:"votes,omitempty"`
	Denies          []*MapsetRankingQueueComment `gorm:"-:all" json:"denies,omitempty"`

	Comments []*MapsetRankingQueueComment `gorm:"foreignKey:MapsetId; references:MapsetId" json:"-"`
}

func (*RankingQueueMapset) TableName() string {
	return "mapset_ranking_queue"
}

func (mapset *RankingQueueMapset) AfterFind(*gorm.DB) (err error) {
	mapset.CreatedAtJSON = time.UnixMilli(mapset.Timestamp)
	mapset.LastUpdatedJSON = time.UnixMilli(mapset.DateLastUpdated)
	return nil
}

// Insert Inserts a mapset into the ranking queue
func (mapset *RankingQueueMapset) Insert() error {
	if err := SQL.Create(&mapset).Error; err != nil {
		return err
	}

	return nil
}

// UpdateStatus Updates the status of a ranking queue mapset
func (mapset *RankingQueueMapset) UpdateStatus(status RankingQueueStatus) error {
	mapset.Status = status

	result := SQL.Model(&RankingQueueMapset{}).
		Where("id = ?", mapset.Id).
		Update("status", status).
		Update("date_last_updated", time.Now().UnixMilli())

	return result.Error
}

// UpdateVoteCount Updates the vote count of a ranking queue mapset
func (mapset *RankingQueueMapset) UpdateVoteCount(votes int) error {
	mapset.VoteCount = votes

	result := SQL.Model(&RankingQueueMapset{}).
		Where("id = ?", mapset.Id).
		Update("votes", votes).
		Update("date_last_updated", time.Now().UnixMilli())

	return result.Error
}

// GetRankingQueue Retrieves the ranking queue for a given game mode
func GetRankingQueue(mode enums.GameMode, limit int, page int) ([]*RankingQueueMapset, error) {
	var mapsets = make([]*RankingQueueMapset, 0)

	result := SQL.
		Joins("Mapset").
		Preload("Mapset.Maps").
		Preload("Comments").
		Preload("Comments.User").
		Joins("LEFT JOIN maps ON maps.mapset_id = Mapset.id").
		Where("(status = ? OR status = ? OR status = ?) "+
			"AND maps.game_mode = ?",
			RankingQueuePending, RankingQueueOnHold, RankingQueueResolved,
			mode).
		Group("maps.mapset_id").
		Order(fmt.Sprintf("votes DESC, "+
			"status = %v DESC, status = %v DESC, status = %v DESC, "+
			"date_last_updated DESC", RankingQueueResolved, RankingQueuePending, RankingQueueOnHold)).
		Limit(limit).
		Offset(page * limit).
		Find(&mapsets)

	if result.Error != nil {
		return nil, result.Error
	}

	for _, mapset := range mapsets {
		mapset.Votes = make([]*MapsetRankingQueueComment, 0)
		mapset.Denies = make([]*MapsetRankingQueueComment, 0)

		for _, comment := range mapset.Comments {
			if !comment.IsActive {
				continue
			}

			switch comment.ActionType {
			case RankingQueueActionVote:
				mapset.Votes = append(mapset.Votes, comment)
			case RankingQueueActionDeny:
				mapset.Denies = append(mapset.Denies, comment)
			}
		}
	}

	return mapsets, nil
}

// GetRankingQueueCount Gets the total amount of maps in the ranking queue
func GetRankingQueueCount(mode enums.GameMode) (int, error) {
	var count int

	result := SQL.Raw("SELECT COUNT(DISTINCT mapset_ranking_queue.id) from mapset_ranking_queue "+
		"LEFT JOIN mapsets ON mapset_ranking_queue.mapset_id = mapsets.id "+
		"LEFT JOIN maps ON maps.mapset_id = mapsets.id "+
		"WHERE (status = ? OR status = ? OR status = ?) AND maps.game_mode = ? ",
		RankingQueuePending, RankingQueueOnHold, RankingQueueResolved, mode).
		Scan(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// GetRankingQueueMapset Retrieves a ranking queue mapset for a given mapset id
func GetRankingQueueMapset(mapsetId int) (*RankingQueueMapset, error) {
	var mapset *RankingQueueMapset

	result := SQL.
		Joins("Mapset").
		Preload("Mapset.User").
		Preload("Mapset.Maps").
		Where("mapset_id = ?", mapsetId).
		First(&mapset)

	if result.Error != nil {
		return nil, result.Error
	}

	return mapset, nil
}

// GetUserMapsetsInRankingQueue Retrieves the mapsets the user has in the ranking queue
func GetUserMapsetsInRankingQueue(userId int) ([]*RankingQueueMapset, error) {
	var mapsets = make([]*RankingQueueMapset, 0)

	result := SQL.
		Joins("Mapset").
		Preload("Mapset.Maps").
		Joins("LEFT JOIN maps ON maps.mapset_id = Mapset.id").
		Where("(status = ? OR status = ? OR status = ?) AND Mapset.creator_id = ?",
			RankingQueuePending, RankingQueueOnHold, RankingQueueResolved, userId).
		Find(&mapsets)

	if result.Error != nil {
		return nil, result.Error
	}

	return mapsets, nil
}
