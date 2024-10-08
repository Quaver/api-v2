package db

import (
	"fmt"
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
	var maps = make([]*UserMostPlayedMap, 0)
	redisKey := fmt.Sprintf("quaver:most_played:%v:%v:%v", id, limit, page)

	if err := CacheJsonInRedis(redisKey, &maps, time.Hour*24, false, func() error {
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
func GetWeeklyMostPlayedMapsets(ignoreCache bool) ([]*WeeklyMostPlayedMapsets, error) {
	var mapsets = make([]*WeeklyMostPlayedMapsets, 0)
	redisKey := "quaver:weekly_most_played"

	bundledMd5s, err := GetBundledMapMd5s()

	if err != nil {
		return nil, err
	}

	bundled := "("

	for index, md5 := range bundledMd5s {
		bundled += fmt.Sprintf("'%v'", md5)

		if index != len(bundledMd5s)-1 {
			bundled += ","
		}
	}

	bundled += ")"

	lastScoreId, err := GetLastScoreId()

	if err != nil {
		return nil, err
	}

	if err := CacheJsonInRedis(redisKey, &mapsets, time.Hour*24, false, func() error {
		return SQL.Raw("SELECT "+
			"maps.mapset_id, maps.creator_id, maps.creator_username, maps.artist, maps.title, COUNT(*) "+
			"FROM scores s "+
			"INNER JOIN "+
			"maps ON maps.md5 = s.map_md5 "+
			"WHERE "+
			"s.id > ? - 300000 AND s.map_md5 NOT IN "+bundled+" AND s.is_donator_score = 0 "+
			"GROUP BY "+
			"maps.mapset_id "+
			"ORDER BY COUNT(*) DESC "+
			"LIMIT 10", lastScoreId).
			Scan(&mapsets).Error
	}); err != nil {
		return nil, err
	}

	return mapsets, nil
}
