package db

import (
	"context"
	"encoding/json"
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

// Helper function to cache json data in redis. The data parameter should be a pointer to the object
// that you're populating
func cacheJsonInRedis(key string, data interface{}, duration time.Duration, fetch func() error) error {
	result, err := Redis.Get(RedisCtx, key).Result()

	if err != nil && err != redis.Nil {
		return err
	}

	// Get cached version
	if result != "" {
		if err := json.Unmarshal([]byte(result), &data); err == nil {
			return nil
		}
	}

	// Call function to fetch the data which sets data
	if err := fetch(); err != nil {
		return err
	}

	// Cache in Redis
	if mapsJson, err := json.Marshal(data); err == nil {
		if err := Redis.Set(RedisCtx, key, mapsJson, duration).Err(); err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}
