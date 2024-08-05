package db

import (
	"github.com/Quaver/api2/enums"
	"gorm.io/gorm"
	"time"
)

type ClanScore struct {
	Id              int            `gorm:"column:id; PRIMARY_KEY" json:"id"`
	ClanId          int            `gorm:"column:clan_id" json:"-"`
	MapMD5          string         `gorm:"column:map_md5" json:"-"`
	Mode            enums.GameMode `gorm:"column:mode" json:"-"`
	OverallRating   float64        `gorm:"column:overall_rating" json:"overall_rating"`
	OverallAccuracy float64        `gorm:"column:overall_accuracy" json:"overall_accuracy"`
	Timestamp       int64          `gorm:"column:timestamp" json:"-"`
	TimestampJSON   time.Time      `gorm:"-:all" json:"timestamp"`
	Map             *MapQua        `gorm:"foreignKey:MapMD5; references:MD5" json:"map,omitempty"`
}

func (*ClanScore) TableName() string {
	return "clan_scores"
}

func (cs *ClanScore) AfterFind(*gorm.DB) error {
	cs.TimestampJSON = time.UnixMilli(cs.Timestamp)
	return nil
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

// GetClanScoreById Retrieves a clan score by id
func GetClanScoreById(id int) (*ClanScore, error) {
	var score ClanScore

	result := SQL.
		Where("id = ?", id).
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

// GetClanScoresForModeFull Retrieves clan scores with full data
func GetClanScoresForModeFull(clanId int, mode enums.GameMode, page int) ([]*ClanScore, error) {
	clanScores := make([]*ClanScore, 0)

	const limit int = 50

	result := SQL.
		Preload("Map").
		Where("clan_id = ? AND mode = ?", clanId, mode).
		Order("overall_rating DESC").
		Offset(page * limit).
		Limit(limit).
		Find(&clanScores)

	if result.Error != nil {
		return nil, result.Error
	}

	return clanScores, nil
}

// CalculateClanScore Calculates a clan score for a given map
func CalculateClanScore(md5 string, clanId int, mode enums.GameMode) (*ClanScore, error) {
	scores, err := GetClanPlayerScoresOnMap(md5, clanId, false)

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
