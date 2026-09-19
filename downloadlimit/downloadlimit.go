package downloadlimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// DailyLimitBytes is the maximum number of mapset bytes a user can download
// during a single server-local calendar day.
const DailyLimitBytes int64 = 1 << 30 // 1GB max: 1 << 30, easier test with 50MB max: 50 * 1024 * 1024

// downloadLimitMu keeps the Redis increment, rollback, and expiry operations from
// interleaving.
var downloadLimitMu sync.Mutex

// TryConsume adds byteCount to a user's daily download total. It returns false
// and restores the counter when the new total exceeds DailyLimitBytes.
func TryConsume(ctx context.Context, redisClient *redis.Client, userID int, byteCount int64) (bool, error) {
	return tryConsumeAt(ctx, redisClient, userID, byteCount, time.Now().In(time.Local))
}

func tryConsumeAt(
	ctx context.Context,
	redisClient *redis.Client,
	userID int,
	byteCount int64,
	now time.Time,
) (bool, error) {
	downloadLimitMu.Lock()
	defer downloadLimitMu.Unlock()

	key, expiresAt := quotaWindow(userID, now)
	used, err := redisClient.IncrBy(ctx, key, byteCount).Result()

	if err != nil {
		return false, fmt.Errorf("increment daily download bytes: %w", err)
	}

	if used > DailyLimitBytes {
		if err := redisClient.DecrBy(ctx, key, byteCount).Err(); err != nil {
			return false, fmt.Errorf("restore daily download bytes: %w", err)
		}
	}

	if err := redisClient.ExpireAt(ctx, key, expiresAt).Err(); err != nil {
		return false, fmt.Errorf("expire daily download counter: %w", err)
	}

	return used <= DailyLimitBytes, nil
}

func quotaWindow(userID int, now time.Time) (string, time.Time) {
	expiresAt := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	key := fmt.Sprintf("quaver:download_rate_limit:%s:%d", now.Format("2006-01-02"), userID)
	return key, expiresAt
}
