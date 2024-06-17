package db

type MultiplayerMatchScore struct {
	Id                int     `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId            int     `gorm:"column:user_id" json:"user_id"`
	MatchId           int     `gorm:"column:match_id" json:"match_id"`
	Team              int8    `gorm:"column:team" json:"-"`
	Modifiers         int64   `gorm:"column:mods" json:"modifiers"`
	PerformanceRating float64 `gorm:"column:performance_rating" json:"performance_rating"`
	Score             int     `gorm:"column:score" json:"-"`
	Accuracy          float64 `gorm:"column:accuracy" json:"accuracy"`
	MaxCombo          int     `gorm:"column:max_combo" json:"max_combo"`
	CountMarvelous    int     `gorm:"column:count_marv" json:"count_marvelous"`
	CountPerfect      int     `gorm:"column:count_perf" json:"count_perfect"`
	CountGreat        int     `gorm:"column:count_great" json:"count_great"`
	CountGood         int     `gorm:"column:count_good" json:"count_good"`
	CountOkay         int     `gorm:"column:count_okay" json:"count_okay"`
	CountMiss         int     `gorm:"column:count_miss" json:"count_miss"`
	FullCombo         bool    `gorm:"column:full_combo" json:"-"`
	LivesLeft         int     `gorm:"column:lives_left" json:"-"`
	HasFailed         bool    `gorm:"column:has_failed" json:"-"`
	Won               bool    `gorm:"column:won" json:"won"`
	BattleRoyalePlace int     `gorm:"column:battle_royale_place" json:"-"`
	User              *User   `gorm:"foreignKey:UserId; references:Id" json:"user"`
}

func (*MultiplayerMatchScore) TableName() string {
	return "multiplayer_match_scores"
}
