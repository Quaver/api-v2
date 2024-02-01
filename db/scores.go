package db

import (
	"gorm.io/gorm"
	"time"
)

type Score struct {
	Id                          int       `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId                      int       `gorm:"column:user_id" json:"user_id"`
	MapMD5                      string    `gorm:"column:map_md5" json:"map_md5"`
	ReplayMD5                   string    `gorm:"column:replay_md5" json:"replay_md5"`
	Timestamp                   int64     `gorm:"column:timestamp" json:"-"`
	TimestampJSON               time.Time `gorm:"-:all" json:"timestamp"`
	IsPersonalBest              bool      `gorm:"column:personal_best" json:"is_personal_best"`
	PerformanceRating           float64   `gorm:"column:performance_rating" json:"performance_rating"`
	Modifiers                   int64     `gorm:"column:mods" json:"modifiers"`
	Failed                      bool      `gorm:"column:failed" json:"failed"`
	TotalScore                  int       `gorm:"column:total_score" json:"total_score"`
	Accuracy                    float64   `gorm:"column:accuracy" json:"accuracy"`
	MaxCombo                    int       `gorm:"column:max_combo" json:"max_combo"`
	CountMarvelous              int       `gorm:"column:count_marv" json:"count_marvelous"`
	CountPerfect                int       `gorm:"column:count_perf" json:"count_perfect"`
	CountGreat                  int       `gorm:"column:count_great" json:"count_great"`
	CountGood                   int       `gorm:"column:count_good" json:"count_good"`
	CountOkay                   int       `gorm:"column:count_okay" json:"count_okay"`
	CountMiss                   int       `gorm:"column:count_miss" json:"count_miss"`
	Grade                       string    `gorm:"column:grade" json:"grade"`
	ScrollSpeed                 int       `gorm:"column:scroll_speed" json:"scroll_speed"`
	TimePlayStart               int64     `gorm:"column:time_play_start" json:"-"`
	TimePlayEnd                 int64     `gorm:"column:time_play_end" json:"-"`
	IP                          string    `gorm:"column:ip" json:"-"`
	ExecutingAssembly           string    `gorm:"column:executing_assembly" json:"-"`
	EntryAssembly               string    `gorm:"column:entry_assembly" json:"-"`
	QuaverVersion               string    `gorm:"column:quaver_version" json:"-"`
	PauseCount                  int       `gorm:"column:pause_count" json:"-"`
	PerformanceProcessorVersion string    `gorm:"column:performance_processor_version" json:"-"`
	DifficultyProcessorVersion  string    `gorm:"column:difficulty_processor_version" json:"-"`
	IsDonatorScore              bool      `gorm:"column:is_donator_score" json:"is_donator_score"`
	TournamentGameId            *int      `gorm:"column:tournament_game_id" json:"tournament_game_id"`
	ClanId                      *int      `gorm:"column:clan_id" json:"clan_id"`
}

func (s *Score) TableName() string {
	return "scores"
}

func (s *Score) BeforeCreate(*gorm.DB) (err error) {
	s.TimestampJSON = time.Now()
	return nil
}

func (s *Score) AfterFind(*gorm.DB) (err error) {
	s.TimestampJSON = time.UnixMilli(s.Timestamp)
	return nil
}
