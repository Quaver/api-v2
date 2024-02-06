package db

type UserStats struct {
	UserId                   int        `gorm:"column:user_id" json:"-"`
	Ranks                    *UserRanks `gorm:"-:all" json:"ranks"`
	TotalScore               int64      `gorm:"column:total_score" json:"total_score"`
	RankedScore              int64      `gorm:"column:ranked_score" json:"ranked_score"`
	OverallAccuracy          float64    `gorm:"column:overall_accuracy" json:"overall_accuracy"`
	OverallPerformanceRating float64    `gorm:"column:overall_performance_rating" json:"overall_performance_rating"`
	PlayCount                int        `gorm:"column:play_count" json:"play_count"`
	FailCount                int        `gorm:"column:fail_count" json:"fail_count"`
	MaxCombo                 int        `gorm:"column:max_combo" json:"max_combo"`
	TotalMarvelous           int        `gorm:"column:total_marv" json:"total_marvelous"`
	TotalPerfect             int        `gorm:"column:total_perf" json:"total_perfect"`
	TotalGreat               int        `gorm:"column:total_great" json:"total_great"`
	TotalGood                int        `gorm:"column:total_good" json:"total_good"`
	TotalOkay                int        `gorm:"column:total_okay" json:"total_okay"`
	TotalMiss                int        `gorm:"column:total_miss" json:"total_miss"`
	CountGradeX              int        `gorm:"column:count_grade_x" json:"count_grade_x"`
	CountGradeSS             int        `gorm:"column:count_grade_ss" json:"count_grade_ss"`
	CountGradeS              int        `gorm:"column:count_grade_s" json:"count_grade_s"`
	CountGradeA              int        `gorm:"column:count_grade_a" json:"count_grade_a"`
	CountGradeB              int        `gorm:"column:count_grade_b" json:"count_grade_b"`
	CountGradeC              int        `gorm:"column:count_grade_c" json:"count_grade_c"`
	CountGradeD              int        `gorm:"column:count_grade_d" json:"count_grade_d"`
}

type UserRanks struct {
	Global  int `json:"global"`
	Country int `json:"country"`
}

type UserStatsKeys4 UserStats
type UserStatsKeys7 UserStats

func (*UserStatsKeys4) TableName() string {
	return "user_stats_keys4"
}

func (*UserStatsKeys7) TableName() string {
	return "user_stats_keys7"
}
