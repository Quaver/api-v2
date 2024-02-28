package db

import "github.com/sirupsen/logrus"

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
