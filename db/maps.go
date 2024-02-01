package db

import "github.com/Quaver/api2/enums"

type MapQua struct {
	Id                   int            `gorm:"column:id; PRIMARY_KEY" json:"id"`
	MapsetId             int            `gorm:"column:mapset_id" json:"mapset_id"`
	MD5                  string         `gorm:"column:md5" json:"md5"`
	AlternativeMD5       string         `gorm:"column:alternative_md5" json:"alternative_md5"`
	CreatorId            int            `gorm:"column:creator_id" json:"creator_id"`
	CreatorUsername      string         `gorm:"column:creator_username" json:"creator_username"`
	GameMode             enums.GameMode `gorm:"column:game_mode" json:"game_mode"`
	RankedStatus         int8           `gorm:"column:ranked_status" json:"ranked_status"`
	Artist               string         `gorm:"column:artist" json:"artist"`
	Title                string         `gorm:"column:title" json:"title"`
	Source               string         `gorm:"column:source" json:"source"`
	Tags                 string         `gorm:"column:tags" json:"tags"`
	Description          string         `gorm:"column:description" json:"description"`
	DifficultyName       string         `gorm:"column:difficulty_name" json:"difficulty_name"`
	Length               int            `gorm:"column:length" json:"length"`
	BPM                  float32        `gorm:"column:bpm" json:"bpm"`
	DifficultyRating     float64        `gorm:"column:difficulty_rating" json:"difficulty_rating"`
	CountHitObjectNormal int            `gorm:"column:count_hitobject_normal" json:"count_hitobject_normal"`
	CountHitObjectLong   int            `gorm:"column:count_hitobject_long" json:"count_hit_object_long"`
	PlayCount            int            `gorm:"column:play_count" json:"play_count"`
	FailCount            int            `gorm:"column:fail_count" json:"fail_count"`
	ModsPending          int            `gorm:"column:mods_pending" json:"mods_pending"`
	ModsAccepted         int            `gorm:"column:mods_accepted" json:"mods_accepted"`
	ModsDenied           int            `gorm:"column:mods_denied" json:"mods_denied"`
	ModsIgnored          int            `gorm:"column:mods_ignored" json:"mods_ignored"`
	OnlineOffset         int            `gorm:"column:online_offset" json:"online_offset"`
	IsClanRanked         bool           `gorm:"column:clan_ranked" json:"is_clan_ranked"`
}

func (m *MapQua) TableName() string {
	return "maps"
}

// GetMapById Retrieves a map from the database by id
func GetMapById(id int) (*MapQua, error) {
	var qua *MapQua

	result := SQL.
		Where("id = ?", id).
		First(&qua)

	if result.Error != nil {
		return nil, result.Error
	}

	return qua, nil
}

// GetMapByMD5 Retrieves a map from the database by md5
func GetMapByMD5(md5 string) (*MapQua, error) {
	var qua *MapQua

	result := SQL.
		Where("md5 = ? OR alternative_md5 = ?", md5, md5).
		First(&qua)

	if result.Error != nil {
		return nil, result.Error
	}

	return qua, nil
}
