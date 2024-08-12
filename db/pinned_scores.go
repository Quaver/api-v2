package db

import "github.com/Quaver/api2/enums"

type PinnedScore struct {
	UserId    int            `gorm:"column:user_id" json:"user_id"`
	GameMode  enums.GameMode `gorm:"column:game_mode" json:"game_mode"`
	ScoreId   int            `gorm:"column:score_id" json:"score_id"`
	SortOrder int            `gorm:"column:sort_order" json:"sort_order"`
	Score     *Score         `gorm:"foreignKey:ScoreId; references:Id" json:"score"`
}

func (*PinnedScore) TableName() string {
	return "pinned_scores"
}

func (ps *PinnedScore) Insert() error {
	return SQL.Create(&ps).Error
}

func (ps *PinnedScore) SortID() int {
	return ps.ScoreId
}

// GetUserPinnedScores Retrieves a user's pinned scores
func GetUserPinnedScores(userId int, mode enums.GameMode) ([]*PinnedScore, error) {
	scores := make([]*PinnedScore, 0)

	result := SQL.
		Preload("Score").
		Preload("Score.Map").
		Where("user_id = ? AND game_mode = ?", userId, mode).
		Order("sort_order ASC").
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

// UpdatePinnedScoreSortOrder Updates the sort order of a given pinned score.
func UpdatePinnedScoreSortOrder(userId int, scoreId int, sortOrder int) error {
	return SQL.Model(&PinnedScore{}).
		Where("user_id = ? AND score_id = ?", userId, scoreId).
		Update("sort_order", sortOrder).Error
}

// SyncPinnedScoreSortOrder Resets the order of pinned scores to keep things in sync
func SyncPinnedScoreSortOrder(userId int, mode enums.GameMode) error {
	pinnedScores, err := GetUserPinnedScores(userId, mode)

	if err != nil {
		return err
	}

	return SyncSortOrder(pinnedScores, func(score *PinnedScore, sortOrder int) error {
		return UpdatePinnedScoreSortOrder(score.UserId, score.ScoreId, sortOrder)
	})
}
