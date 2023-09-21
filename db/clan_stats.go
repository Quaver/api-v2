package db

type ClanStats struct {
	ClanId                   int     `gorm:"column:clan_id" json:"clan_id"`
	Mode                     int     `gorm:"column:mode" json:"mode"`
	OverallAccuracy          float64 `gorm:"column:overall_accuracy" json:"overall_accuracy"`
	OverallPerformanceRating float64 `gorm:"column:overall_performance_rating" json:"overall_performance_rating"`
	TotalMarv                int     `gorm:"column:total_marv" json:"total_marv"`
	TotalPerf                int     `gorm:"column:total_perf" json:"total_perf"`
	TotalGreat               int     `gorm:"column:total_great" json:"total_great"`
	TotalGood                int     `gorm:"column:total_good" json:"total_good"`
	TotalOkay                int     `gorm:"column:total_okay" json:"total_okay"`
	TotalMiss                int     `gorm:"column:total_miss" json:"total_miss"`
}

func (*ClanStats) TableName() string {
	return "clan_stats"
}
