package db

import (
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"strconv"
)

const (
	onlineUsersRedisKey  string = "quaver:server:online_users"
	totalUsersRedisKey   string = "quaver:total_users"
	totalMapsetsRedisKey string = "quaver:total_mapsets"
	totalScoresRedisKey  string = "quaver:total_scores"
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
