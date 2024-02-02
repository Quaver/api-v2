package db

type ScoreFirstPlace struct {
	MD5               string  `gorm:"column:md5" json:"md5"`
	UserId            int     `gorm:"column:user_id" json:"user_id"`
	ScoreId           int     `gorm:"column:score_id" json:"score_id"`
	PerformanceRating float64 `gorm:"column:performance_rating" json:"performance_rating"`
}

func (s *ScoreFirstPlace) TableName() string {
	return "scores_first_place"
}
