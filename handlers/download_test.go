package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/downloadlimit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var downloadTestUserSequence atomic.Int64

func TestEnforceMapsetDownloadLimitUsesFileSize(t *testing.T) {
	client, userID := useTestDownloadRedis(t)
	path := writeTestMapset(t, []byte("mapset-data"))
	ctx := newDownloadTestContext()

	if apiErr := enforceMapsetDownloadLimit(ctx, userID, path); apiErr != nil {
		t.Fatalf("API error = %#v", apiErr)
	}

	actual, err := client.Get(context.Background(), downloadLimitKey(userID)).Int64()
	if err != nil {
		t.Fatal(err)
	}

	if expected := int64(len("mapset-data")); actual != expected {
		t.Fatalf("counted bytes = %d, want %d", actual, expected)
	}
}

func TestEnforceMapsetDownloadLimitReturnsTooManyRequests(t *testing.T) {
	client, userID := useTestDownloadRedis(t)
	ctx := newDownloadTestContext()
	allowed, err := downloadlimit.TryConsume(
		ctx.Request.Context(),
		client,
		userID,
		downloadlimit.DailyLimitBytes,
	)

	if err != nil || !allowed {
		t.Fatalf("initial usage: allowed=%v err=%v", allowed, err)
	}

	apiErr := enforceMapsetDownloadLimit(ctx, userID, writeTestMapset(t, []byte{1}))

	if apiErr == nil || apiErr.Status != http.StatusTooManyRequests {
		t.Fatalf("API error = %#v, want status %d", apiErr, http.StatusTooManyRequests)
	}

	if apiErr.Message != "Download rate limit has been reached" {
		t.Fatalf("message = %q", apiErr.Message)
	}

	actual, err := client.Get(context.Background(), downloadLimitKey(userID)).Int64()
	if err != nil {
		t.Fatal(err)
	}

	if actual != downloadlimit.DailyLimitBytes {
		t.Fatalf("counter = %d, want %d", actual, downloadlimit.DailyLimitBytes)
	}
}

func TestEnforceMapsetDownloadLimitReturnsServerError(t *testing.T) {
	previousClient := db.Redis
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:0",
		MaxRetries:   -1,
		DialTimeout:  10 * time.Millisecond,
		ReadTimeout:  10 * time.Millisecond,
		WriteTimeout: 10 * time.Millisecond,
	})
	db.Redis = client
	t.Cleanup(func() {
		db.Redis = previousClient
		_ = client.Close()
	})

	apiErr := enforceMapsetDownloadLimit(newDownloadTestContext(), 303, writeTestMapset(t, []byte{1}))

	if apiErr == nil || apiErr.Status != http.StatusInternalServerError {
		t.Fatalf("API error = %#v, want status %d", apiErr, http.StatusInternalServerError)
	}
}

func TestSetFileContentLength(t *testing.T) {
	path := writeTestMapset(t, []byte("head-response"))
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodHead, "/v2/download/mapset/1", nil)

	if apiErr := setFileContentLength(ctx, path); apiErr != nil {
		t.Fatalf("API error = %#v", apiErr)
	}

	if got := response.Header().Get("Content-Length"); got != strconv.Itoa(len("head-response")) {
		t.Fatalf("Content-Length = %q, want %d", got, len("head-response"))
	}
}

func useTestDownloadRedis(t *testing.T) (*redis.Client, int) {
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

	userID := int(time.Now().UnixNano() + downloadTestUserSequence.Add(1))
	key := downloadLimitKey(userID)
	t.Cleanup(func() { _ = db.Redis.Del(context.Background(), key).Err() })

	return db.Redis, userID
}

func downloadLimitKey(userID int) string {
	return fmt.Sprintf(
		"quaver:download_rate_limit:%s:%d",
		time.Now().In(time.Local).Format("2006-01-02"),
		userID,
	)
}

func newDownloadTestContext() *gin.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v2/download/mapset/1", nil)
	return ctx
}

func writeTestMapset(t *testing.T, contents []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "mapset.qp")

	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}
