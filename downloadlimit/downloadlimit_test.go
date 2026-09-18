package downloadlimit

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestTryReserveAccumulatesAndEnforcesLimit(t *testing.T) {
	server, client := newTestRedis(t)
	ctx := context.Background()
	now := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)
	server.SetTime(now)

	first, allowed, err := tryReserveAt(ctx, client, 42, 128, now, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !allowed || first == nil {
		t.Fatal("expected first reservation to be allowed")
	}

	second, allowed, err := tryReserveAt(ctx, client, 42, DailyLimitBytes-128, now, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !allowed || second == nil {
		t.Fatal("expected reservation reaching the exact limit to be allowed")
	}

	reservation, allowed, err := tryReserveAt(ctx, client, 42, 1, now, nil)
	if err != nil {
		t.Fatal(err)
	}

	if allowed || reservation != nil {
		t.Fatal("expected reservation above the limit to be rejected")
	}

	key, expiresAt := quotaWindow(42, now)
	assertCounterValue(t, ctx, client, key, DailyLimitBytes)

	if ttl := server.TTL(key); ttl != expiresAt.Sub(now) {
		t.Fatalf("TTL = %v, want %v", ttl, expiresAt.Sub(now))
	}
}

func TestTryReserveRetriesWatchConflict(t *testing.T) {
	server, client := newTestRedis(t)
	otherClient := redis.NewClient(client.Options())
	t.Cleanup(func() { _ = otherClient.Close() })

	ctx := context.Background()
	now := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)
	server.SetTime(now)
	key, _ := quotaWindow(7, now)
	hookCalls := 0

	reservation, allowed, err := tryReserveAt(ctx, client, 7, 11, now, func(attempt int) error {
		hookCalls++

		if attempt == 0 {
			return otherClient.IncrBy(ctx, key, 7).Err()
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if !allowed || reservation == nil {
		t.Fatal("expected reservation to succeed after retry")
	}

	if hookCalls < 2 {
		t.Fatalf("hook called %d times, want at least 2", hookCalls)
	}

	assertCounterValue(t, ctx, client, key, 18)
}

func TestTryReserveRetriesStaleRejection(t *testing.T) {
	server, client := newTestRedis(t)
	otherClient := redis.NewClient(client.Options())
	t.Cleanup(func() { _ = otherClient.Close() })

	ctx := context.Background()
	now := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)
	server.SetTime(now)
	fullReservation, allowed, err := tryReserveAt(ctx, client, 8, DailyLimitBytes, now, nil)

	if err != nil || !allowed {
		t.Fatalf("initial reservation: allowed=%v err=%v", allowed, err)
	}

	hookCalls := 0
	reservation, allowed, err := tryReserveAt(ctx, client, 8, 1, now, func(attempt int) error {
		hookCalls++

		if attempt == 0 {
			return Release(ctx, otherClient, fullReservation)
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if !allowed || reservation == nil {
		t.Fatal("expected reservation to be allowed after the concurrent release")
	}

	if hookCalls < 2 {
		t.Fatalf("hook called %d times, want at least 2", hookCalls)
	}

	key, _ := quotaWindow(8, now)
	assertCounterValue(t, ctx, client, key, 1)
}

func TestTryReserveConcurrentRequestsDoNotExceedLimit(t *testing.T) {
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

			_, allowed, err := tryReserveAt(ctx, client, 99, chunkSize, now, nil)

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
		t.Fatalf("allowed reservations = %d, want %d", got, allowedCount)
	}

	key, _ := quotaWindow(99, now)
	assertCounterValue(t, ctx, client, key, DailyLimitBytes)
}

func TestTryReserveExpiresAtNextLocalMidnightAcrossDST(t *testing.T) {
	server, client := newTestRedis(t)
	ctx := context.Background()
	location, err := time.LoadLocation("Europe/Sofia")

	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, time.March, 29, 0, 30, 0, 0, location)
	server.SetTime(now)
	_, allowed, err := tryReserveAt(ctx, client, 15, 1, now, nil)

	if err != nil {
		t.Fatal(err)
	}

	if !allowed {
		t.Fatal("expected reservation to be allowed")
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

func TestReleaseRestoresCounterAndHandlesExpiredReservation(t *testing.T) {
	server, client := newTestRedis(t)
	ctx := context.Background()
	dayOne := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)
	server.SetTime(dayOne)

	first, allowed, err := tryReserveAt(ctx, client, 21, 100, dayOne, nil)
	if err != nil || !allowed {
		t.Fatalf("first reservation: allowed=%v err=%v", allowed, err)
	}

	_, allowed, err = tryReserveAt(ctx, client, 21, 25, dayOne, nil)
	if err != nil || !allowed {
		t.Fatalf("second reservation: allowed=%v err=%v", allowed, err)
	}

	if err := Release(ctx, client, first); err != nil {
		t.Fatal(err)
	}

	dayOneKey, dayOneExpiry := quotaWindow(21, dayOne)
	assertCounterValue(t, ctx, client, dayOneKey, 25)

	server.FastForward(dayOneExpiry.Sub(dayOne) + time.Second)
	dayTwo := dayOneExpiry.Add(time.Hour)
	server.SetTime(dayTwo)
	_, allowed, err = tryReserveAt(ctx, client, 21, 200, dayTwo, nil)
	if err != nil || !allowed {
		t.Fatalf("day two reservation: allowed=%v err=%v", allowed, err)
	}

	if err := Release(ctx, client, first); err != nil {
		t.Fatal(err)
	}

	dayTwoKey, _ := quotaWindow(21, dayTwo)
	assertCounterValue(t, ctx, client, dayTwoKey, 200)
}

func TestReleaseRetriesConcurrentCounterChange(t *testing.T) {
	server, client := newTestRedis(t)
	otherClient := redis.NewClient(client.Options())
	t.Cleanup(func() { _ = otherClient.Close() })

	ctx := context.Background()
	now := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)
	server.SetTime(now)
	reservation, allowed, err := tryReserveAt(ctx, client, 31, 100, now, nil)

	if err != nil || !allowed {
		t.Fatalf("initial reservation: allowed=%v err=%v", allowed, err)
	}

	hookCalls := 0
	err = release(ctx, client, reservation, func(attempt int) error {
		hookCalls++

		if attempt == 0 {
			_, allowed, err := tryReserveAt(ctx, otherClient, 31, 25, now, nil)

			if err != nil {
				return err
			}

			if !allowed {
				return errors.New("concurrent reservation was unexpectedly rejected")
			}
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if hookCalls < 2 {
		t.Fatalf("hook called %d times, want at least 2", hookCalls)
	}

	key, _ := quotaWindow(31, now)
	assertCounterValue(t, ctx, client, key, 25)
}

func TestTryReserveReturnsRedisAndCounterErrors(t *testing.T) {
	server, client := newTestRedis(t)
	ctx := context.Background()
	now := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)
	key, _ := quotaWindow(55, now)
	server.Set(key, "not-an-integer")

	if _, _, err := tryReserveAt(ctx, client, 55, 1, now, nil); err == nil {
		t.Fatal("expected malformed counter to return an error")
	}

	server.Close()

	if _, _, err := tryReserveAt(ctx, client, 56, 1, now, nil); err == nil {
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
