package integration

import (
	"context"
	"testing"
)

func TestSelect(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("SELECT 1, SET db1key hello")
	rdb.Do(ctx, "SELECT", "1")
	rdb.Set(ctx, "db1key", "hello", 0)

	t.Log("SELECT 0, GET db1key (should not exist)")
	rdb.Do(ctx, "SELECT", "0")
	_, err := rdb.Get(ctx, "db1key").Result()
	if err == nil {
		t.Error("GET db1key in db0 should return error (key not found)")
	}
	t.Logf("GET db1key in db0: not found (expected)")

	t.Log("SELECT 1, GET db1key (should exist)")
	rdb.Do(ctx, "SELECT", "1")
	val, err := rdb.Get(ctx, "db1key").Result()
	if err != nil {
		t.Fatalf("GET db1key in db1 failed: %v", err)
	}
	t.Logf("GET db1key in db1 = %q", val)
	if val != "hello" {
		t.Errorf("GET db1key = %q, want hello", val)
	}

	rdb.Do(ctx, "SELECT", "0")
}

func TestSelect_InvalidIndex(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("SELECT -1 (invalid index)")
	_, err := rdb.Do(ctx, "SELECT", "-1").Result()
	if err == nil {
		t.Error("SELECT -1 should return error")
	}
	t.Logf("SELECT -1 returned error (expected): %v", err)
}
