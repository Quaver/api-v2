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
	Id              int                `gorm:"column:id; PRIMARY_KEY" json:"id"`
	MapsetId        int                `gorm:"column:mapset_id" json:"mapset_id"`
	Timestamp       int64              `gorm:"column:timestamp" json:"-"`
	CreatedAtJSON   time.Time          `gorm:"-:all" json:"created_at"` // Same value as Timestamp
	DateLastUpdated int64              `gorm:"column:date_last_updated" json:"-"`
	LastUpdatedJSON time.Time          `gorm:"-:all" json:"last_updated"` // Same value as DateLastUpdated
	Status          RankingQueueStatus `gorm:"column:status" json:"status"`
	NeedsAttention  bool               `gorm:"column:needs_attention" json:"-"`
	Votes           int                `gorm:"column:votes" json:"votes"`
	Mapset          *Mapset            `gorm:"foreignKey:MapsetId; references:Id" json:"mapset"`
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
	mapset.Votes = votes

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
		Joins("LEFT JOIN maps ON maps.mapset_id = Mapset.id").
		Where("(status = ? OR status = ? OR status = ?) "+
			"AND maps.game_mode = ?",
			RankingQueuePending, RankingQueueOnHold, RankingQueueResolved,
			mode).
		Order(fmt.Sprintf("votes DESC, "+
			"status = %v DESC, status = %v DESC, status = %v DESC, "+
			"date_last_updated DESC", RankingQueueResolved, RankingQueuePending, RankingQueueOnHold)).
		Limit(limit).
		Offset(page * limit).
		Find(&mapsets)

	if result.Error != nil {
		return nil, result.Error
	}

	return mapsets, nil
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
