package db

import "context"

// ClearScoreboardCache deletes the single hash containing a map's scoreboards.
func ClearScoreboardCache(ctx context.Context, md5 string) (int64, error) {
	return Redis.Del(ctx, scoreboardRedisKey(md5)).Result()
}
