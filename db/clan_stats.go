package db

import (
	"github.com/Quaver/api2/enums"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strconv"
)

type ClanStats struct {
	ClanId                   int            `gorm:"column:clan_id" json:"-"`
	Mode                     enums.GameMode `gorm:"column:mode" json:"mode"`
	Rank                     int            `gorm:"-:all" json:"rank"`
	OverallAccuracy          float64        `gorm:"column:overall_accuracy" json:"overall_accuracy"`
	OverallPerformanceRating float64        `gorm:"column:overall_performance_rating" json:"overall_performance_rating"`
	TotalMarv                int            `gorm:"column:total_marv" json:"total_marv"`
	TotalPerf                int            `gorm:"column:total_perf" json:"total_perf"`
	TotalGreat               int            `gorm:"column:total_great" json:"total_great"`
	TotalGood                int            `gorm:"column:total_good" json:"total_good"`
	TotalOkay                int            `gorm:"column:total_okay" json:"total_okay"`
	TotalMiss                int            `gorm:"column:total_miss" json:"total_miss"`
}

func (*ClanStats) TableName() string {
	return "clan_stats"
}

func (cs *ClanStats) AfterFind(*gorm.DB) error {
	result, err := Redis.ZRevRank(RedisCtx, ClanLeaderboardKey(cs.Mode), strconv.Itoa(cs.ClanId)).Result()

	if err != nil {
		if err != redis.Nil {
			logrus.Error("Error fetching clan rank", err)
		}

		cs.Rank = -1
		return nil
	}

	cs.Rank = int(result) + 1
	return nil
}

func (cs *ClanStats) Save() error {
	return SQL.Model(&ClanStats{}).
		Where("clan_id = ? AND mode = ?", cs.ClanId, cs.Mode).
		Update("overall_accuracy", cs.OverallAccuracy).
		Update("overall_performance_rating", cs.OverallPerformanceRating).
		Update("total_marv", cs.TotalMarv).
		Update("total_perf", cs.TotalPerf).
		Update("total_great", cs.TotalGreat).
		Update("total_good", cs.TotalGood).
		Update("total_okay", cs.TotalOkay).
		Update("total_miss", cs.TotalMiss).Error
}

// GetClanStatsByMode Retrieves clan stats by its game mode
func GetClanStatsByMode(id int, mode enums.GameMode) (*ClanStats, error) {
	var stats ClanStats

	result := SQL.
		Where("clan_id = ? AND mode = ?", id, mode).
		First(&stats)

	if result.Error != nil {
		return nil, result.Error
	}

	return &stats, nil
}

// PerformFullClanRecalculation Recalculates all of a clan's scores + stats
func PerformFullClanRecalculation(clan *Clan) error {
	for i := 1; i <= 2; i++ {
		mode := enums.GameMode(i)

		clanScores, err := GetClanScoresForMode(clan.Id, mode)

		if err != nil {
			return err
		}

		for _, clanScore := range clanScores {
			newScore, err := CalculateClanScore(clanScore.MapMD5, clan.Id, clanScore.Mode)

			if err != nil {
				return err
			}

			newScore.Id = clanScore.Id

			if err := SQL.Save(&newScore).Error; err != nil {
				return err
			}
		}

		if err := RecalculateClanStats(clan.Id, mode); err != nil {
			return err
		}

	}

	return nil
}

// RecalculateClanStats Recalculates a clan stats for a given mode.
func RecalculateClanStats(id int, mode enums.GameMode, newScore ...*RedisScore) error {
	stats, err := GetClanStatsByMode(id, mode)

	if err != nil {
		return err
	}

	if len(newScore) > 0 {
		stats.TotalMarv += newScore[0].Score.CountMarvelous
		stats.TotalPerf += newScore[0].Score.CountPerfect
		stats.TotalGreat += newScore[0].Score.CountGreat
		stats.TotalGood += newScore[0].Score.CountGood
		stats.TotalOkay += newScore[0].Score.CountOkay
		stats.TotalMiss += newScore[0].Score.CountMiss
	}

	clanScores, err := GetClanScoresForMode(id, mode)

	if err != nil {
		return err
	}

	convertedScores := make([]*Score, 0)

	for _, clanScore := range clanScores {
		convertedScores = append(convertedScores, clanScore.ToScore())
	}

	stats.OverallPerformanceRating = CalculateOverallRating(convertedScores)
	stats.OverallAccuracy = CalculateOverallAccuracy(convertedScores)

	return stats.Save()
}
