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

// GetRankingQueue Retrieves the ranking queue for a given game mode
func GetRankingQueue(mode enums.GameMode, limit int, page int) ([]*RankingQueueMapset, error) {
	var mapsets []*RankingQueueMapset

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
