package db

import (
	"fmt"
	"github.com/Quaver/api2/enums"
	"sort"
	"strings"
)

// GetGlobalLeaderboard Retrieves the global leaderboard for a specific game mode
func GetGlobalLeaderboard(mode enums.GameMode, page int, limit int) ([]*User, error) {
	key := fmt.Sprintf("quaver:leaderboard:%v", mode)

	users, err := getLeaderboardUsers(key, page, limit)

	if err != nil {
		return []*User{}, err
	}

	sort.Slice(users, func(i, j int) bool {
		switch mode {
		case enums.GameModeKeys4:
			return users[i].StatsKeys4.Ranks.Global < users[j].StatsKeys4.Ranks.Global
		case enums.GameModeKeys7:
			return users[i].StatsKeys7.Ranks.Global < users[j].StatsKeys7.Ranks.Global
		default:
			return true
		}
	})

	return users, nil
}

// GetCountryLeaderboard Retrieves the country leaderboard for a given country and mode
func GetCountryLeaderboard(country string, mode enums.GameMode, page int, limit int) ([]*User, error) {
	key := fmt.Sprintf("quaver:country_leaderboard:%v:%v", strings.ToLower(country), mode)

	users, err := getLeaderboardUsers(key, page, limit)

	if err != nil {
		return []*User{}, err
	}

	sort.Slice(users, func(i, j int) bool {
		switch mode {
		case enums.GameModeKeys4:
			return users[i].StatsKeys4.Ranks.Country < users[j].StatsKeys4.Ranks.Country
		case enums.GameModeKeys7:
			return users[i].StatsKeys7.Ranks.Country < users[j].StatsKeys7.Ranks.Country
		default:
			return true
		}
	})

	return users, nil
}

// GetTotalHitsLeaderboard  Retrieves the total hits leaderboard for a specific game mode
func GetTotalHitsLeaderboard(page int, limit int) ([]*User, error) {
	users, err := getLeaderboardUsers("quaver:leaderboard:total_hits_global", page, limit)

	if err != nil {
		return []*User{}, err
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].StatsKeys4.Ranks.TotalHits < users[j].StatsKeys4.Ranks.TotalHits
	})

	return users, nil
}

// Function to get users from a leaderboard
func getLeaderboardUsers(key string, page int, limit int) ([]*User, error) {
	userIds, err := Redis.ZRevRange(RedisCtx, key, int64(page*limit), int64(page*limit+limit-1)).Result()

	if err != nil {
		return nil, err
	}

	if len(userIds) == 0 {
		return []*User{}, nil
	}

	var users = make([]*User, 0)

	result := SQL.
		Joins("StatsKeys4").
		Joins("StatsKeys7").
		Where(fmt.Sprintf("users.id IN (%v) AND allowed = 1", strings.Join(userIds, ","))).
		Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}
