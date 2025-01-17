package db

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/enums"
	"gorm.io/gorm"
	"strings"
	"time"
)

type MapQua struct {
	Id                   int                `gorm:"column:id; PRIMARY_KEY" json:"id"`
	MapsetId             int                `gorm:"column:mapset_id" json:"mapset_id"`
	MD5                  string             `gorm:"column:md5" json:"md5"`
	AlternativeMD5       string             `gorm:"column:alternative_md5" json:"alternative_md5"`
	CreatorId            int                `gorm:"column:creator_id" json:"creator_id"`
	CreatorUsername      string             `gorm:"column:creator_username" json:"creator_username"`
	GameMode             enums.GameMode     `gorm:"column:game_mode" json:"game_mode"`
	RankedStatus         enums.RankedStatus `gorm:"column:ranked_status" json:"ranked_status"`
	Artist               string             `gorm:"column:artist" json:"artist"`
	Title                string             `gorm:"column:title" json:"title"`
	Source               string             `gorm:"column:source" json:"source"`
	Tags                 string             `gorm:"column:tags" json:"tags"`
	Description          string             `gorm:"column:description" json:"description"`
	DifficultyName       string             `gorm:"column:difficulty_name" json:"difficulty_name"`
	Length               int                `gorm:"column:length" json:"length"`
	BPM                  float32            `gorm:"column:bpm" json:"bpm"`
	DifficultyRating     float64            `gorm:"column:difficulty_rating" json:"difficulty_rating"`
	CountHitObjectNormal int                `gorm:"column:count_hitobject_normal" json:"count_hitobject_normal"`
	CountHitObjectLong   int                `gorm:"column:count_hitobject_long" json:"count_hitobject_long"`
	LongNotePercentage   float32            `gorm:"-:all" json:"long_note_percentage"`
	MaxCombo             int                `gorm:"-:all" json:"max_combo"`
	PlayCount            int                `gorm:"column:play_count" json:"play_count"`
	FailCount            int                `gorm:"column:fail_count" json:"fail_count"`
	PlayAttempts         int                `gorm:"-:all" json:"play_attempts"`
	ModsPending          int                `gorm:"column:mods_pending" json:"mods_pending"`
	ModsAccepted         int                `gorm:"column:mods_accepted" json:"mods_accepted"`
	ModsDenied           int                `gorm:"column:mods_denied" json:"mods_denied"`
	ModsIgnored          int                `gorm:"column:mods_ignored" json:"mods_ignored"`
	OnlineOffset         int                `gorm:"column:online_offset" json:"online_offset"`
	IsClanRanked         bool               `gorm:"column:clan_ranked" json:"is_clan_ranked"`
	DateClanRanked       *int64             `gorm:"column:date_clan_ranked" json:"-"`
	DateClanRankedJSON   *time.Time         `gorm:"column:date_clan_ranked" json:"date_clan_ranked"`
}

func (m *MapQua) TableName() string {
	return "maps"
}

func (m *MapQua) AfterFind(*gorm.DB) error {
	m.MaxCombo = m.CountHitObjectLong*2 + m.CountHitObjectNormal
	m.PlayAttempts = m.PlayCount + m.FailCount

	if m.CountHitObjectLong != 0 {
		m.LongNotePercentage = float32(m.CountHitObjectLong) /
			float32(m.CountHitObjectNormal+m.CountHitObjectLong) * 100
	}

	if m.DateClanRanked != nil {
		t := time.UnixMilli(*m.DateClanRanked)
		m.DateClanRankedJSON = &t
	}

	return nil
}

func (m *MapQua) String() string {
	return fmt.Sprintf("%v - %v [%v]", m.Artist, m.Title, m.DifficultyName)
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

// GetMapByMD5AndAlternative Retrieves a map from the database by either its md5 or alternative md5
func GetMapByMD5AndAlternative(md5 string, alternativeMd5 string) (*MapQua, error) {
	var qua *MapQua

	result := SQL.
		Where("md5 = ? OR alternative_md5 = ?", md5, alternativeMd5).
		First(&qua)

	if result.Error != nil {
		return nil, result.Error
	}

	return qua, nil
}

// GetMapsInMapset Retrieves all maps in a given set
func GetMapsInMapset(id int) ([]*MapQua, error) {
	maps := make([]*MapQua, 0)

	result := SQL.Where("mapset_id = ?", id).Find(&maps)

	if result.Error != nil {
		return nil, result.Error
	}

	return maps, nil
}

// InsertMap Inserts a map into the database
func InsertMap(m *MapQua) error {
	if err := SQL.Create(&m).Error; err != nil {
		return err
	}

	return nil
}

// UpdateMapMD5 Updates the md5 hash of a map
func UpdateMapMD5(id int, md5 string) error {
	result := SQL.Model(&MapQua{}).
		Where("id = ?", id).
		Update("md5", md5)

	return result.Error
}

// DeleteMap Deletes a map from the DB
func DeleteMap(id int) error {
	result := SQL.Delete(&MapQua{}, "id = ?", id)
	return result.Error
}

// UpdateMapDifficultyRating Updates the difficulty rating of a map
func UpdateMapDifficultyRating(id int, difficultyRating float64) error {
	result := SQL.Model(&MapQua{}).
		Where("id = ?", id).
		Update("difficulty_rating", difficultyRating)

	return result.Error
}

// UpdateMapClanRanked Updates the clan ranked status of a map
func UpdateMapClanRanked(id int, clanRanked bool) error {
	result := SQL.Model(&MapQua{}).
		Where("id = ?", id).
		Update("clan_ranked", clanRanked).
		Update("date_clan_ranked", time.Now().UnixMilli())

	return result.Error
}

func GetBundledMapMd5s() ([]string, error) {
	md5s := make([]string, 0)
	bundledStringSlice := make([]string, len(config.Instance.BundledMapsets))

	for i, num := range config.Instance.BundledMapsets {
		bundledStringSlice[i] = fmt.Sprintf("%d", num)
	}

	bundled := fmt.Sprintf("(%v)", strings.Join(bundledStringSlice, ", "))

	result := SQL.Raw("SELECT md5 FROM maps " +
		"INNER JOIN mapsets ON maps.mapset_id = mapsets.id " +
		"WHERE maps.mapset_id IN " + bundled).Scan(&md5s)

	if result.Error != nil {
		return nil, result.Error
	}

	return md5s, nil
}
