package db

import "fmt"

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

	return maps, nil
}
