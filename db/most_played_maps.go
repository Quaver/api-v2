package db

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"strings"
	"time"
)

type UserMostPlayedMap struct {
	Id              int    `gorm:"column:id" json:"id"`
	CreatorId       int    `gorm:"column:creator_id" json:"creator_id"`
	CreatorUsername string `gorm:"column:creator_username" json:"creator_username"`
	Artist          string `gorm:"column:artist" json:"artist"`
	Title           string `gorm:"column:title" json:"title"`
	DifficultyName  string `gorm:"column:difficulty_name" json:"difficulty_name"`
	Count           int    `gorm:"column:COUNT(*)" json:"count"`
}

// GetUserMostPlayedMaps Returns a user's most played maps
func GetUserMostPlayedMaps(id int, limit int, page int) ([]*UserMostPlayedMap, error) {
	var maps []*UserMostPlayedMap
	redisKey := fmt.Sprintf("quaver:most_played:%v:%v:%v", id, limit, page)

	if err := cacheJsonInRedis(redisKey, &maps, time.Hour*24, func() error {
		return SQL.Raw("SELECT "+
			"maps.id, maps.creator_id, maps.creator_username, maps.artist, maps.title, maps.difficulty_name, COUNT(*) "+
			"FROM scores s "+
			"INNER JOIN "+
			"maps ON maps.md5 = s.map_md5 "+
			"WHERE "+
			"s.user_id = ? AND maps.mapset_id != -1 AND maps.ranked_status = 2 "+
			"GROUP BY "+
			"maps.id "+
			"ORDER BY COUNT(*) DESC "+
			fmt.Sprintf("LIMIT %v OFFSET %v", limit, page*limit), id).
			Scan(&maps).Error
	}); err != nil {
		return nil, err
	}

	return maps, nil
}

type WeeklyMostPlayedMapsets struct {
	MapsetId        int    `gorm:"column:mapset_id" json:"id"`
	CreatorId       int    `gorm:"column:creator_id" json:"creator_id"`
	CreatorUsername string `gorm:"column:creator_username" json:"creator_username"`
	Artist          string `gorm:"column:artist" json:"artist"`
	Title           string `gorm:"column:title" json:"title"`
	Count           int    `gorm:"column:COUNT(*)" json:"count"`
}

// GetWeeklyMostPlayedMapsets Retrieves the most played mapsets in the past week
func GetWeeklyMostPlayedMapsets() ([]*WeeklyMostPlayedMapsets, error) {
	var mapsets []*WeeklyMostPlayedMapsets
	redisKey := "quaver:weekly_most_played"

	// Convert bundled mapset int slice to string slice, so we can use strings.Join
	bundledStringSlice := make([]string, len(config.Instance.BundledMapsets))

	for i, num := range config.Instance.BundledMapsets {
		bundledStringSlice[i] = fmt.Sprintf("%d", num)
	}

	bundled := fmt.Sprintf("(%v)", strings.Join(bundledStringSlice, ", "))

	if err := cacheJsonInRedis(redisKey, &mapsets, time.Hour*24, func() error {
		return SQL.Raw("SELECT "+
			"maps.mapset_id, maps.creator_id, maps.creator_username, maps.artist, maps.title, COUNT(*) "+
			"FROM scores s "+
			"INNER JOIN "+
			"maps ON maps.md5 = s.map_md5 "+
			"WHERE "+
			"s.timestamp > ? AND maps.mapset_id NOT IN "+bundled+" AND maps.mapset_id != -1 AND maps.ranked_status = 2 "+
			"GROUP BY "+
			"maps.mapset_id "+
			"ORDER BY COUNT(*) DESC "+
			"LIMIT 10", time.Now().AddDate(0, 0, -7).UnixMilli()).
			Scan(&mapsets).Error
	}); err != nil {
		return nil, err
	}

	return mapsets, nil
}
