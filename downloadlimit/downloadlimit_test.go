package downloadlimit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestTryConsumeAccumulatesAndEnforcesLimit(t *testing.T) {
	server, client := newTestRedis(t)
	ctx := context.Background()
	now := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)
	server.SetTime(now)

	allowed, err := tryConsumeAt(ctx, client, 42, 128, now)
	if err != nil || !allowed {
		t.Fatalf("first download: allowed=%v err=%v", allowed, err)
	}

	allowed, err = tryConsumeAt(ctx, client, 42, DailyLimitBytes-128, now)
	if err != nil || !allowed {
		t.Fatalf("download reaching limit: allowed=%v err=%v", allowed, err)
	}

	allowed, err = tryConsumeAt(ctx, client, 42, 1, now)
	if err != nil {
		t.Fatal(err)
	}

	if allowed {
		t.Fatal("expected download above the limit to be rejected")
	}

	key, expiresAt := quotaWindow(42, now)
	assertCounterValue(t, ctx, client, key, DailyLimitBytes)

	if ttl := server.TTL(key); ttl != expiresAt.Sub(now) {
		t.Fatalf("TTL = %v, want %v", ttl, expiresAt.Sub(now))
	}
}

func TestTryConsumeConcurrentRequestsDoNotExceedLimit(t *testing.T) {
	server, client := newTestRedis(t)
	ctx := context.Background()
	now := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)
	server.SetTime(now)
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

			allowed, err := tryConsumeAt(ctx, client, 99, chunkSize, now)

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

	key, _ := quotaWindow(99, now)
	assertCounterValue(t, ctx, client, key, DailyLimitBytes)
}

func TestTryConsumeExpiresAtNextLocalMidnightAcrossDST(t *testing.T) {
	server, client := newTestRedis(t)
	ctx := context.Background()
	location, err := time.LoadLocation("Europe/Sofia")

	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, time.March, 29, 0, 30, 0, 0, location)
	server.SetTime(now)
	allowed, err := tryConsumeAt(ctx, client, 15, 1, now)

	if err != nil || !allowed {
		t.Fatalf("download: allowed=%v err=%v", allowed, err)
	}

	key, expiresAt := quotaWindow(15, now)
	expectedTTL := 22*time.Hour + 30*time.Minute

	if expiresAt.Sub(now) != expectedTTL {
		t.Fatalf("midnight duration = %v, want %v", expiresAt.Sub(now), expectedTTL)
	}

	if ttl := server.TTL(key); ttl != expectedTTL {
		t.Fatalf("TTL = %v, want %v", ttl, expectedTTL)
	}
}

func TestTryConsumeReturnsRedisErrors(t *testing.T) {
	server, client := newTestRedis(t)
	ctx := context.Background()
	now := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)
	key, _ := quotaWindow(55, now)
	server.Set(key, "not-an-integer")

	if _, err := tryConsumeAt(ctx, client, 55, 1, now); err == nil {
		t.Fatal("expected malformed counter to return an error")
	}

	server.Close()

	if _, err := tryConsumeAt(ctx, client, 56, 1, now); err == nil {
		t.Fatal("expected unavailable redis to return an error")
	}
}

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr:         server.Addr(),
		MaxRetries:   -1,
		DialTimeout:  100 * time.Millisecond,
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
	})
	t.Cleanup(func() { _ = client.Close() })

	return server, client
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
