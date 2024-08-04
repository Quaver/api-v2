package db

import (
	"github.com/Quaver/api2/enums"
	"time"
)

type ClanScore struct {
	Id              int            `gorm:"column:id; PRIMARY_KEY"`
	ClanId          int            `gorm:"column:clan_id"`
	MapMD5          string         `gorm:"column:map_md5"`
	Mode            enums.GameMode `gorm:"column:mode"`
	OverallRating   float64        `gorm:"column:overall_rating"`
	OverallAccuracy float64        `gorm:"column:overall_accuracy"`
	Timestamp       int64          `gorm:"column:timestamp"`
}

func (*ClanScore) TableName() string {
	return "clan_scores"
}

// ToScore Converts a clan score to a traditional score object, so we can reuse functions to calculate rating/acc
func (cs *ClanScore) ToScore() *Score {
	return &Score{
		ClanId:            &cs.ClanId,
		MapMD5:            cs.MapMD5,
		PerformanceRating: cs.OverallRating,
		Accuracy:          cs.OverallAccuracy,
	}
}

// GetClanScore Retrieves an existing clan score
func GetClanScore(md5 string, clanId int) (*ClanScore, error) {
	var score ClanScore

	result := SQL.
		Where("map_md5 = ? AND clan_id = ?", md5, clanId).
		First(&score)

	if result.Error != nil {
		return nil, result.Error
	}

	return &score, nil
}

// GetClanScoresForMode Retrieves all clan scores for a given mode
func GetClanScoresForMode(clanId int, mode enums.GameMode) ([]*ClanScore, error) {
	clanScores := make([]*ClanScore, 0)

	result := SQL.
		Where("clan_id = ? AND mode = ?", clanId, mode).
		Find(&clanScores)

	if result.Error != nil {
		return nil, result.Error
	}

	return clanScores, nil
}

// CalculateClanScore Calculates a clan score for a given map
func CalculateClanScore(md5 string, clanId int, mode enums.GameMode) (*ClanScore, error) {
	scores, err := GetClanPlayerScoresOnMap(md5, clanId)

	if err != nil {
		return nil, err
	}

	score := &ClanScore{
		ClanId:          clanId,
		MapMD5:          md5,
		Mode:            mode,
		OverallRating:   CalculateOverallRating(scores),
		OverallAccuracy: CalculateOverallAccuracy(scores),
		Timestamp:       time.Now().UnixMilli(),
	}

	return score, nil
}
