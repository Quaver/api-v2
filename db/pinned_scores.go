package db

import "github.com/Quaver/api2/enums"

type PinnedScore struct {
	UserId   int            `gorm:"column:user_id" json:"user_id"`
	GameMode enums.GameMode `gorm:"column:game_mode" json:"game_mode"`
	ScoreId  int            `gorm:"column:score_id" json:"score_id"`
	Score    *Score         `gorm:"foreignKey:ScoreId; references:Id" json:"score"`
}

func (*PinnedScore) TableName() string {
	return "pinned_scores"
}

func (ps *PinnedScore) Insert() error {
	return SQL.Create(&ps).Error
}

// GetUserPinnedScores Retrieves a user's pinned scores
func GetUserPinnedScores(userId int, mode enums.GameMode) ([]*PinnedScore, error) {
	scores := make([]*PinnedScore, 0)

	result := SQL.
		Preload("Score").
		Preload("Score.Map").
		Where("user_id = ? AND game_mode = ?", userId, mode).
		Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	return scores, nil
}

// DeletePinnedScore Deletes a pinned score
func DeletePinnedScore(userId int, scoreId int) error {
	return SQL.Delete(&PinnedScore{}, "user_id = ? AND score_id = ?", userId, scoreId).Error
}
