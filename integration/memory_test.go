package integration

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestMemoryStats(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("MEMORY STATS")
	res, err := rdb.Do(ctx, "MEMORY", "STATS").Slice()
	if err != nil {
		t.Fatalf("MEMORY STATS failed: %v", err)
	}
	t.Logf("MEMORY STATS returned %d fields", len(res))
	if len(res) == 0 {
		t.Error("MEMORY STATS should return non-empty result")
	}
}

func TestMemoryUsage(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "k", "hello", 0)
	t.Log("MEMORY USAGE k")

	usage, err := rdb.Do(ctx, "MEMORY", "USAGE", "k").Int64()
	if err != nil {
		t.Fatalf("MEMORY USAGE k failed: %v", err)
	}
	t.Logf("MEMORY USAGE k = %d bytes", usage)
	if usage <= 0 {
		t.Errorf("MEMORY USAGE k = %d, want > 0", usage)
	}

	t.Log("MEMORY USAGE nokey (non-existent)")
	_, err = rdb.Do(ctx, "MEMORY", "USAGE", "nokey").Result()
	if err != redis.Nil {
		t.Errorf("MEMORY USAGE nokey should return nil, got: %v", err)
	}
	t.Logf("MEMORY USAGE nokey returned nil (expected)")
}
