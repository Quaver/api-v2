package db

import (
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

const (
	onlineUsersRedisKey    string = "quaver:server:online_users"
	totalUsersRedisKey     string = "quaver:total_users"
	totalMapsetsRedisKey   string = "quaver:total_mapsets"
	totalScoresRedisKey    string = "quaver:total_scores"
	countryPlayersRedisKey string = "quaver:country_players"
)

// CacheTotalUsersInRedis Caches the number of total scores in redis
func CacheTotalUsersInRedis() {
	var users int

	if err := SQL.Raw("SELECT COUNT(*) FROM users").Scan(&users).Error; err != nil {
		panic(err)
	}

	if err := Redis.Set(RedisCtx, totalUsersRedisKey, users, 0).Err(); err != nil {
		panic(err)
	}

	logrus.Info("Cached total user count in redis")
}

// CacheTotalMapsetsInRedis Caches the number of total scores in redis
func CacheTotalMapsetsInRedis() {
	var mapsets int

	if err := SQL.Raw("SELECT COUNT(*) FROM mapsets").Scan(&mapsets).Error; err != nil {
		panic(err)
	}

	if err := Redis.Set(RedisCtx, totalMapsetsRedisKey, mapsets, 0).Err(); err != nil {
		panic(err)
	}

	logrus.Info("Cached total mapset count in redis")
}

// CacheTotalScoresInRedis Caches the number of total scores in redis
func CacheTotalScoresInRedis() {
	var scores int

	if err := SQL.Raw("SELECT COUNT(*) FROM scores").Scan(&scores).Error; err != nil {
		panic(err)
	}

	if err := Redis.Set(RedisCtx, totalScoresRedisKey, scores, 0).Err(); err != nil {
		panic(err)
	}

	logrus.Info("Cached total score count in redis")
}

// GetOnlineUserCountFromRedis Retrieves the amount of online users from redis
func GetOnlineUserCountFromRedis() (int, error) {
	return getValueFromRedis(onlineUsersRedisKey)
}

// GetTotalUserCountFromRedis Retrieves the total amount of users from redis
func GetTotalUserCountFromRedis() (int, error) {
	return getValueFromRedis(totalUsersRedisKey)
}

// GetTotalMapsetCountFromRedis Retrieves the total amount of mapsets from redis
func GetTotalMapsetCountFromRedis() (int, error) {
	return getValueFromRedis(totalMapsetsRedisKey)
}

// GetTotalScoreCountFromRedis Retrieves the total amount of scores from redis
func GetTotalScoreCountFromRedis() (int, error) {
	return getValueFromRedis(totalScoresRedisKey)
}

// GetTotalCountryPlayersFromRedis Retrieves the total amount of users on the country leaderboard
func GetTotalCountryPlayersFromRedis() (int, error) {
	result, err := Redis.HGet(RedisCtx, countryPlayersRedisKey, "total").Result()

	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}

		return 0, err
	}

	count, err := strconv.Atoi(result)

	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetCountryPlayerCountFromRedis Gets the total amount of users from a given country in redis
func GetCountryPlayerCountFromRedis(country string) (int, error) {
	result, err := Redis.HGet(RedisCtx, countryPlayersRedisKey, strings.ToLower(country)).Result()

	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}

		return 0, err
	}

	count, err := strconv.Atoi(result)

	if err != nil {
		return 0, err
	}

	return count, nil
}

// CacheCountryPlayersInRedis Caches the amount of country players in redis
func CacheCountryPlayersInRedis() (map[string]string, error) {
	totalUserCount, err := GetTotalUnbannedUserCount()

	if err != nil {
		return nil, err
	}

	countryPlayerCount, err := GetTotalCountryPlayersFromRedis()

	if err != nil {
		return nil, err
	}

	type CountryPlayers struct {
		CountryCode string `gorm:"column:country" json:"country"`
		Total       int    `gorm:"column:total" json:"total"`
	}

	var countries = make([]*CountryPlayers, 0)

	// Total user count does match the amount of country players we have cached, so re-cache.
	if totalUserCount != countryPlayerCount {
		result := SQL.
			Raw("SELECT country, count(country) as total FROM users WHERE allowed = 1 GROUP BY country").
			Scan(&countries)

		if result.Error != nil {
			return nil, result.Error
		}

		totalUsers := 0

		// Cache total players in each country
		for _, country := range countries {
			err := Redis.
				HSet(RedisCtx, countryPlayersRedisKey, strings.ToLower(country.CountryCode), country.Total).
				Err()

			if err != nil {
				return nil, err
			}

			totalUsers += country.Total
		}

		// Cache total country players
		err := Redis.HSet(RedisCtx, countryPlayersRedisKey, "total", totalUsers).Err()

		if err != nil {
			return nil, err
		}
	}

	result, err := Redis.HGetAll(RedisCtx, countryPlayersRedisKey).Result()

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Retrieves a single cached value from redis
func getValueFromRedis(key string) (int, error) {
	result, err := Redis.Get(RedisCtx, key).Result()

	if err != nil && err != redis.Nil {
		return 0, err
	}

	value, err := strconv.Atoi(result)

	if err != nil {
		return 0, err
	}

	return value, nil
}
