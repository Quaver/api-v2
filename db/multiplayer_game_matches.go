package db

import (
	"github.com/Quaver/api2/enums"
	"time"
)

type MultiplayerGameMatches struct {
	Id              int                      `gorm:"column:id; PRIMARY_KEY" json:"id"`
	GameId          int                      `gorm:"column:game_id" json:"game_id"`
	TimePlayed      int64                    `gorm:"column:time_played" json:"-"`
	TimePlayedJSON  time.Time                `gorm:"-:all" json:"time_played"`
	MapMD5          string                   `gorm:"column:map_md5" json:"map_md5"`
	MapString       string                   `gorm:"column:map" json:"map_string"`
	HostId          int                      `gorm:"column:host_id" json:"host_id"`
	Ruleset         int8                     `gorm:"column:ruleset" json:"-"`
	GameMode        enums.GameMode           `gorm:"column:game_mode" json:"game_mode"`
	GlobalModifiers int64                    `gorm:"column:global_modifiers" json:"global_modifiers"`
	FreeModType     int8                     `gorm:"column:free_mod_type" json:"free_mod_type"`
	HealthType      int8                     `gorm:"column:health_type" json:"-"`
	Lives           int                      `gorm:"column:lives" json:"-"`
	Aborted         bool                     `gorm:"column:aborted" json:"aborted"`
	Map             *MapQua                  `gorm:"foreignKey:MapMD5; references:MD5" json:"map"`
	Scores          []*MultiplayerMatchScore `gorm:"foreignKey:MatchId;" json:"scores,omitempty"`
}

func (*MultiplayerGameMatches) TableName() string {
	return "multiplayer_game_matches"
}
