package db

import (
	"fmt"
	"github.com/Quaver/api2/enums"
	"time"
)

type UserRank struct {
	UserId                   int       `gorm:"column:user_id" json:"-"`
	Rank                     int       `gorm:"column:rank" json:"rank"`
	OverallPerformanceRating float64   `gorm:"column:overall_performance_rating" json:"overall_performance_rating"`
	Timestamp                time.Time `gorm:"column:timestamp" json:"timestamp"`
}

type UserRankKeys4 UserRank

func (ur *UserRankKeys4) TableName() string {
	return "user_rank_keys4"
}

type UserRankKeys7 UserRank

func (ur *UserRankKeys7) TableName() string {
	return "user_rank_keys7"
}

// GetUserRankStatisticsForMode Retrieves a users rank statistics for a given game mode
func GetUserRankStatisticsForMode(id int, mode enums.GameMode) ([]*UserRank, error) {
	var ranks = make([]*UserRank, 0)

	// Invalid mode
	if mode < enums.GameModeKeys4 || mode > enums.GameModeKeys7 {
		return ranks, nil
	}

	modeStr := fmt.Sprintf("user_rank_%v", enums.GetGameModeString(mode))

	result := SQL.
		Where("user_id = ?", id).
		Table(modeStr).
		Find(&ranks)

	if result.Error != nil {
		return nil, result.Error
	}

	return ranks, nil
}
