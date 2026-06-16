package integration

import (
	"context"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("SET mykey myval")
	err := rdb.Set(ctx, "mykey", "myval", 0).Err()
	if err != nil {
		t.Fatalf("SET failed: %v", err)
	}

	val, err := rdb.Get(ctx, "mykey").Result()
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	t.Logf("GET mykey = %q", val)
	if val != "myval" {
		t.Errorf("GET mykey = %q, want myval", val)
	}
}

func TestSetGet_Overwrite(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("SET k v1, then SET k v2")
	rdb.Set(ctx, "k", "v1", 0)
	rdb.Set(ctx, "k", "v2", 0)

	val, err := rdb.Get(ctx, "k").Result()
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	t.Logf("GET k = %q (expect overwritten value v2)", val)
	if val != "v2" {
		t.Errorf("GET k = %q, want v2", val)
	}
}

func TestSetGet_WithEX(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("SET ttlkey val EX 2")
	err := rdb.Set(ctx, "ttlkey", "val", 2*time.Second).Err()
	if err != nil {
		t.Fatalf("SET EX failed: %v", err)
	}

	val, err := rdb.Get(ctx, "ttlkey").Result()
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	t.Logf("GET ttlkey = %q", val)
	if val != "val" {
		t.Errorf("GET ttlkey = %q, want val", val)
	}

	ttl := rdb.TTL(ctx, "ttlkey").Val()
	t.Logf("TTL ttlkey = %v", ttl)
	if ttl <= 0 || ttl > 2*time.Second {
		t.Errorf("TTL ttlkey = %v, want > 0 and <= 2s", ttl)
	}
}

func TestGet_NonExistent(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("GET nokey (non-existent)")
	_, err := rdb.Get(ctx, "nokey").Result()
	if err == nil {
		t.Error("GET non-existent key should return error")
	}
	t.Logf("GET nokey returned error (expected): %v", err)
}

func TestAppend_Create(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("APPEND mykey world (key does not exist)")
	n, err := rdb.Append(ctx, "mykey", "world").Result()
	if err != nil {
		t.Fatalf("APPEND failed: %v", err)
	}
	t.Logf("APPEND returned length %d", n)
	if n != 5 {
		t.Errorf("APPEND = %d, want 5", n)
	}

	val, _ := rdb.Get(ctx, "mykey").Result()
	t.Logf("GET mykey = %q", val)
	if val != "world" {
		t.Errorf("GET mykey = %q, want world", val)
	}
}

func TestAppend_Existing(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "greeting", "hello", 0)
	t.Log("SET greeting hello, then APPEND greeting ' world'")

	n, err := rdb.Append(ctx, "greeting", " world").Result()
	if err != nil {
		t.Fatalf("APPEND failed: %v", err)
	}
	t.Logf("APPEND returned length %d", n)
	if n != 11 {
		t.Errorf("APPEND = %d, want 11", n)
	}

	val, _ := rdb.Get(ctx, "greeting").Result()
	t.Logf("GET greeting = %q", val)
	if val != "hello world" {
		t.Errorf("GET greeting = %q, want 'hello world'", val)
	}
}
