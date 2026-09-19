package downloadlimit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/redis/go-redis/v9"
)

var testUserSequence atomic.Int64

func TestTryConsumeAccumulatesAndEnforcesLimit(t *testing.T) {
	client, userID, now := useTestRedis(t)
	ctx := context.Background()

	allowed, err := tryConsumeAt(ctx, client, userID, 128, now)
	if err != nil || !allowed {
		t.Fatalf("first download: allowed=%v err=%v", allowed, err)
	}

	allowed, err = tryConsumeAt(ctx, client, userID, DailyLimitBytes-128, now)
	if err != nil || !allowed {
		t.Fatalf("download reaching limit: allowed=%v err=%v", allowed, err)
	}

	allowed, err = tryConsumeAt(ctx, client, userID, 1, now)
	if err != nil {
		t.Fatal(err)
	}

	if allowed {
		t.Fatal("expected download above the limit to be rejected")
	}

	key, _ := quotaWindow(userID, now)
	assertCounterValue(t, ctx, client, key, DailyLimitBytes)

	if ttl, err := client.TTL(ctx, key).Result(); err != nil || ttl <= 0 {
		t.Fatalf("TTL = %v, err=%v", ttl, err)
	}
}

func TestTryConsumeConcurrentRequestsDoNotExceedLimit(t *testing.T) {
	client, userID, now := useTestRedis(t)
	ctx := context.Background()
	const workerCount = 24
	const allowedCount = 8
	chunkSize := DailyLimitBytes / allowedCount
	start := make(chan struct{})
	errorsChannel := make(chan error, workerCount)
	var successful atomic.Int64
	var waitGroup sync.WaitGroup

	for range workerCount {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()
			<-start

			allowed, err := tryConsumeAt(ctx, client, userID, chunkSize, now)

			if err != nil {
				errorsChannel <- err
				return
			}

			if allowed {
				successful.Add(1)
			}
		}()
	}

	close(start)
	waitGroup.Wait()
	close(errorsChannel)

	for err := range errorsChannel {
		t.Fatal(err)
	}

	if got := successful.Load(); got != allowedCount {
		t.Fatalf("allowed downloads = %d, want %d", got, allowedCount)
	}

	key, _ := quotaWindow(userID, now)
	assertCounterValue(t, ctx, client, key, DailyLimitBytes)
}

func TestQuotaWindowEndsAtNextLocalMidnightAcrossDST(t *testing.T) {
	location, err := time.LoadLocation("Europe/Sofia")

	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, time.March, 29, 0, 30, 0, 0, location)
	_, expiresAt := quotaWindow(15, now)
	expectedTTL := 22*time.Hour + 30*time.Minute

	if expiresAt.Sub(now) != expectedTTL {
		t.Fatalf("midnight duration = %v, want %v", expiresAt.Sub(now), expectedTTL)
	}
}

func TestTryConsumeReturnsRedisErrors(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:0",
		MaxRetries:   -1,
		DialTimeout:  10 * time.Millisecond,
		ReadTimeout:  10 * time.Millisecond,
		WriteTimeout: 10 * time.Millisecond,
	})
	t.Cleanup(func() { _ = client.Close() })

	if _, err := TryConsume(context.Background(), client, 55, 1); err == nil {
		t.Fatal("expected Redis error")
	}
}

func useTestRedis(t *testing.T) (*redis.Client, int, time.Time) {
	t.Helper()

	if config.Instance == nil {
		if err := config.Load("../config.json"); err != nil {
			t.Fatal(err)
		}
	}

	db.InitializeRedis()
	if err := db.Redis.Ping(context.Background()).Err(); err != nil {
		t.Fatal(err)
	}

	now := time.Now().In(time.Local)
	userID := int(now.UnixNano() + testUserSequence.Add(1))
	key, _ := quotaWindow(userID, now)
	t.Cleanup(func() { _ = db.Redis.Del(context.Background(), key).Err() })

	return db.Redis, userID, now
}

func assertCounterValue(t *testing.T, ctx context.Context, client *redis.Client, key string, expected int64) {
	t.Helper()

	actual, err := client.Get(ctx, key).Int64()

	if err != nil {
		t.Fatal(err)
	}

	if actual != expected {
		t.Fatalf("counter = %d, want %d", actual, expected)
	}
}
