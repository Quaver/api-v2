package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/downloadlimit"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestEnforceMapsetDownloadLimitUsesFileSize(t *testing.T) {
	_, client := useTestDownloadRedis(t)
	path := writeTestMapset(t, []byte("mapset-data"))
	ctx := newDownloadTestContext()

	apiErr := enforceMapsetDownloadLimit(ctx, 101, path)

	if apiErr != nil {
		t.Fatalf("API error = %#v", apiErr)
	}

	keys, err := client.Keys(context.Background(), "quaver:download_rate_limit:*").Result()

	if err != nil {
		t.Fatal(err)
	}

	if len(keys) != 1 {
		t.Fatalf("quota keys = %v, want exactly one", keys)
	}

	actual, err := client.Get(context.Background(), keys[0]).Int64()

	if err != nil {
		t.Fatal(err)
	}

	if expected := int64(len("mapset-data")); actual != expected {
		t.Fatalf("counted bytes = %d, want %d", actual, expected)
	}
}

func TestEnforceMapsetDownloadLimitReturnsTooManyRequests(t *testing.T) {
	_, client := useTestDownloadRedis(t)
	ctx := newDownloadTestContext()
	allowed, err := downloadlimit.TryConsume(
		ctx.Request.Context(),
		client,
		202,
		downloadlimit.DailyLimitBytes,
	)

	if err != nil || !allowed {
		t.Fatalf("initial usage: allowed=%v err=%v", allowed, err)
	}

	apiErr := enforceMapsetDownloadLimit(ctx, 202, writeTestMapset(t, []byte{1}))

	if apiErr == nil || apiErr.Status != http.StatusTooManyRequests {
		t.Fatalf("API error = %#v, want status %d", apiErr, http.StatusTooManyRequests)
	}

	if apiErr.Message != "Download rate limit has been reached" {
		t.Fatalf("message = %q", apiErr.Message)
	}

	keys, err := client.Keys(context.Background(), "quaver:download_rate_limit:*").Result()

	if err != nil {
		t.Fatal(err)
	}

	if len(keys) != 1 {
		t.Fatalf("quota keys = %v, want exactly one", keys)
	}

	actual, err := client.Get(context.Background(), keys[0]).Int64()

	if err != nil {
		t.Fatal(err)
	}

	if actual != downloadlimit.DailyLimitBytes {
		t.Fatalf("counter = %d, want %d", actual, downloadlimit.DailyLimitBytes)
	}
}

func TestEnforceMapsetDownloadLimitReturnsServerError(t *testing.T) {
	server, _ := useTestDownloadRedis(t)
	server.Close()
	ctx := newDownloadTestContext()

	apiErr := enforceMapsetDownloadLimit(ctx, 303, writeTestMapset(t, []byte{1}))

	if apiErr == nil || apiErr.Status != http.StatusInternalServerError {
		t.Fatalf("API error = %#v, want status %d", apiErr, http.StatusInternalServerError)
	}
}

func TestSetFileContentLengthDoesNotCountQuota(t *testing.T) {
	_, client := useTestDownloadRedis(t)
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

	keys, err := client.Keys(context.Background(), "quaver:download_rate_limit:*").Result()

	if err != nil {
		t.Fatal(err)
	}

	if len(keys) != 0 {
		t.Fatalf("HEAD request created quota keys: %v", keys)
	}
}

func useTestDownloadRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr:         server.Addr(),
		MaxRetries:   -1,
		DialTimeout:  100 * time.Millisecond,
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
	})
	previousClient := db.Redis
	db.Redis = client

	t.Cleanup(func() {
		db.Redis = previousClient
		_ = client.Close()
	})

	return server, client
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
