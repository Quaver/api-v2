package db

import (
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
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
	result, err := Redis.Get(RedisCtx, redisKey).Result()

	if err != nil && err != redis.Nil {
		return nil, err
	}

	// Get cached version
	if result != "" {
		if err := json.Unmarshal([]byte(result), &maps); err == nil {
			return maps, nil
		}
	}

	if err := SQL.Raw("SELECT "+
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
		Scan(&maps).Error; err != nil {
		return nil, err
	}

	// Cache in Redis
	if mapsJson, err := json.Marshal(maps); err == nil {
		if err := Redis.Set(RedisCtx, redisKey, mapsJson, time.Hour*24).Err(); err != nil {
			return nil, err
		}
	} else {
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

	if err := SQL.Raw("SELECT "+
		"maps.mapset_id, maps.creator_id, maps.creator_username, maps.artist, maps.title, COUNT(*) "+
		"FROM scores s "+
		"INNER JOIN "+
		"maps ON maps.md5 = s.map_md5 "+
		"WHERE "+
		"s.timestamp > ? AND maps.mapset_id NOT IN (919, 536, 563, 523, 922, 919, 9, 923, 994, 954, 822, 21846) AND maps.mapset_id != -1 AND maps.ranked_status = 2 "+
		"GROUP BY "+
		"maps.mapset_id "+
		"ORDER BY COUNT(*) DESC "+
		"LIMIT 10", time.Now().AddDate(0, 0, -7).UnixMilli()).
		Scan(&mapsets).Error; err != nil {
		return nil, err
	}

	return mapsets, nil
}
