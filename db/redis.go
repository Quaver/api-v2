package db

import (
	"context"
	"github.com/Quaver/api2/config"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

var (
	Redis    *redis.Client
	RedisCtx = context.Background()
)

// InitializeRedis Initializes a Redis client
func InitializeRedis() {
	if Redis != nil {
		return
	}

	Redis = redis.NewClient(&redis.Options{
		Addr:         config.Instance.Redis.Host,
		Password:     config.Instance.Redis.Password,
		DB:           config.Instance.Redis.Database,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	})

	result := Redis.Ping(RedisCtx)

	if result.Err() != nil {
		logrus.Error(result.Err())
	}

	logrus.Info("Successfully connected to redis")
}

// Parses a redis string to an int with a default value if there's an error.
func parseRedisIntWithDefault(str string, defaultVal int) int {
	val, err := strconv.Atoi(str)

	if err != nil {
		return defaultVal
	}

	return val
}
