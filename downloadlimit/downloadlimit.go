package downloadlimit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// DailyLimitBytes is the maximum number of mapset bytes a user can download
	// during a single server-local calendar day.
	DailyLimitBytes int64 = 1 << 30 // 1GB, easier test with 50MB max: 50 * 1024 * 1024

	maxTransactionRetries = 32
	maxRetryBackoff       = 10 * time.Millisecond
)

var (
	errLimitExceeded               = errors.New("daily download limit exceeded")
	errTransactionRetriesExhausted = errors.New("download limit transaction retries exhausted")
)

// Reservation identifies bytes that were added to a daily download counter.
// Its fields are intentionally private so callers can only release reservations
// that were returned by TryReserve.
type Reservation struct {
	key       string
	byteCount int64
	expiresAt time.Time
}

type transactionHook func(attempt int) error

// TryReserve atomically reserves byteCount bytes from a user's server-local
// daily download allowance. The returned boolean is false when the reservation
// would exceed DailyLimitBytes.
func TryReserve(ctx context.Context, redisClient *redis.Client, userID int, byteCount int64) (*Reservation, bool, error) {
	return tryReserveAt(ctx, redisClient, userID, byteCount, time.Now().In(time.Local), nil)
}

func tryReserveAt(
	ctx context.Context,
	redisClient *redis.Client,
	userID int,
	byteCount int64,
	now time.Time,
	beforeCommit transactionHook,
) (*Reservation, bool, error) {
	if redisClient == nil {
		return nil, false, errors.New("redis client is nil")
	}

	if userID <= 0 {
		return nil, false, fmt.Errorf("invalid user id: %d", userID)
	}

	if byteCount < 0 {
		return nil, false, fmt.Errorf("invalid download byte count: %d", byteCount)
	}

	key, expiresAt := quotaWindow(userID, now)
	reservation := &Reservation{
		key:       key,
		byteCount: byteCount,
		expiresAt: expiresAt,
	}

	err := watchWithRetry(ctx, redisClient, key, func(tx *redis.Tx, attempt int) error {
		current, err := tx.Get(ctx, key).Int64()

		if err == redis.Nil {
			current = 0
		} else if err != nil {
			return fmt.Errorf("read daily download counter: %w", err)
		}

		if current < 0 {
			return fmt.Errorf("daily download counter cannot be negative: %d", current)
		}

		if beforeCommit != nil {
			if err := beforeCommit(attempt); err != nil {
				return err
			}
		}

		if current > DailyLimitBytes || byteCount > DailyLimitBytes-current {
			// Execute a read-only transaction so WATCH can detect a concurrent
			// release before we reject based on the value read above.
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Exists(ctx, key)
				return nil
			})

			if err != nil {
				return err
			}

			return errLimitExceeded
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.IncrBy(ctx, key, byteCount)
			pipe.ExpireAt(ctx, key, expiresAt)
			return nil
		})

		return err
	})

	if errors.Is(err, errLimitExceeded) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("reserve daily download bytes: %w", err)
	}

	return reservation, true, nil
}

// Release removes a previously reserved byte count. It is safe to call after
// the reservation's key has expired; in that case it is a no-op.
func Release(ctx context.Context, redisClient *redis.Client, reservation *Reservation) error {
	return release(ctx, redisClient, reservation, nil)
}

func release(
	ctx context.Context,
	redisClient *redis.Client,
	reservation *Reservation,
	beforeCommit transactionHook,
) error {
	if redisClient == nil {
		return errors.New("redis client is nil")
	}

	if reservation == nil {
		return errors.New("download reservation is nil")
	}

	err := watchWithRetry(ctx, redisClient, reservation.key, func(tx *redis.Tx, attempt int) error {
		current, err := tx.Get(ctx, reservation.key).Int64()

		if err == redis.Nil {
			return nil
		}

		if err != nil {
			return fmt.Errorf("read daily download counter: %w", err)
		}

		if current < 0 {
			return fmt.Errorf("daily download counter cannot be negative: %d", current)
		}

		if current < reservation.byteCount {
			return fmt.Errorf(
				"daily download counter underflow: current=%d reservation=%d",
				current,
				reservation.byteCount,
			)
		}

		if beforeCommit != nil {
			if err := beforeCommit(attempt); err != nil {
				return err
			}
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			if current == reservation.byteCount {
				pipe.Del(ctx, reservation.key)
				return nil
			}

			pipe.DecrBy(ctx, reservation.key, reservation.byteCount)
			pipe.ExpireAt(ctx, reservation.key, reservation.expiresAt)
			return nil
		})

		return err
	})

	if err != nil {
		return fmt.Errorf("release daily download bytes: %w", err)
	}

	return nil
}

func watchWithRetry(
	ctx context.Context,
	redisClient *redis.Client,
	key string,
	operation func(tx *redis.Tx, attempt int) error,
) error {
	for attempt := 0; attempt < maxTransactionRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := redisClient.Watch(ctx, func(tx *redis.Tx) error {
			return operation(tx, attempt)
		}, key)

		if !errors.Is(err, redis.TxFailedErr) {
			return err
		}

		if attempt == maxTransactionRetries-1 {
			break
		}

		if err := waitForRetry(ctx, attempt); err != nil {
			return err
		}
	}

	return fmt.Errorf("%w for key %q", errTransactionRetriesExhausted, key)
}

func waitForRetry(ctx context.Context, attempt int) error {
	delay := time.Duration(attempt+1) * time.Millisecond

	if delay > maxRetryBackoff {
		delay = maxRetryBackoff
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func quotaWindow(userID int, now time.Time) (string, time.Time) {
	expiresAt := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	key := fmt.Sprintf("quaver:download_rate_limit:%s:%d", now.Format("2006-01-02"), userID)
	return key, expiresAt
}
