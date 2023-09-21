package db

type ClanScore struct {
	Id              int     `gorm:"column:id; PRIMARY_KEY"`
	ClanId          int     `gorm:"column:clan_id"`
	MapMD5          string  `gorm:"column:map_md5"`
	OverallRating   float64 `gorm:"column:overall_rating"`
	OverallAccuracy float64 `gorm:"column:overall_accuracy"`
}

func (*ClanScore) TableName() string {
	return "clan_scores"
}
