package db

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ScoreFirstPlace struct {
	MD5               string  `gorm:"column:md5" json:"md5"`
	UserId            int     `gorm:"column:user_id" json:"user_id"`
	ScoreId           int     `gorm:"column:score_id" json:"score_id"`
	PerformanceRating float64 `gorm:"column:performance_rating" json:"performance_rating"`
}

func (s *ScoreFirstPlace) TableName() string {
	return "scores_first_place"
}

// GetUserFirstPlaces Retrieves all of a user's first place scores
func GetUserFirstPlaces(userId int) ([]*ScoreFirstPlace, error) {
	var firstPlaces = make([]*ScoreFirstPlace, 0)

	result := SQL.
		Where("user_id = ?", userId).
		Find(&firstPlaces)

	if result.Error != nil {
		return nil, result.Error
	}

	return firstPlaces, nil
}

func UpdateFirstPlace(md5 string, score *Score) error {
	result := SQL.Model(&ScoreFirstPlace{}).
		Where("md5 = ?", md5).
		Updates(map[string]interface{}{
			"user_id":            score.UserId,
			"score_id":           score.Id,
			"performance_rating": score.PerformanceRating,
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// ReplaceUserFirstPlaces Goes through a user's first place scores and replaces
// it with the next unbanned player. This is used specifically when a player is banned.
func ReplaceUserFirstPlaces(userId int) error {
	firstPlaces, err := GetUserFirstPlaces(userId)

	if err != nil {
		return err
	}

	for index, firstPlace := range firstPlaces {
		score, err := getFirstPlaceScoreOnMap(firstPlace.MD5)

		switch err {
		case nil:
			break
		case gorm.ErrRecordNotFound:
			continue
		default:
			return err
		}

		if err := UpdateFirstPlace(firstPlace.MD5, score); err != nil {
			return err
		}

		logrus.Infof("[First Place Updated %v/%v] "+
			"User #%v -> #%v (Map: %v)", index+1, len(firstPlaces), userId, score.UserId, score.MapMD5)
	}

	return nil
}

// Retrieves the new first place score on a map
func getFirstPlaceScoreOnMap(md5 string) (*Score, error) {
	var score *Score

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.personal_best = 1 "+
			"AND user.allowed = 1", md5).
		Order("scores.performance_rating DESC").
		First(&score)

	if result.Error != nil {
		return nil, result.Error
	}

	return score, nil
}
