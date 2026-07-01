package integration

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestHSetAndHGet(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("HSET h f1 v1 f2 v2")
	n, err := rdb.HSet(ctx, "h", "f1", "v1", "f2", "v2").Result()
	if err != nil {
		t.Fatalf("HSET failed: %v", err)
	}
	t.Logf("HSET returned %d", n)
	if n != 2 {
		t.Errorf("HSET = %d, want 2", n)
	}

	val, err := rdb.HGet(ctx, "h", "f1").Result()
	if err != nil {
		t.Fatalf("HGET failed: %v", err)
	}
	t.Logf("HGET h f1 = %q", val)
	if val != "v1" {
		t.Errorf("HGET = %q, want v1", val)
	}
}

func TestHGet_NonExistent(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	_, err := rdb.HGet(ctx, "h", "f").Result()
	if err != redis.Nil {
		t.Fatalf("HGET missing = %v, want redis.Nil", err)
	}
}

func TestHGetAll(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.HSet(ctx, "h", "a", "1", "b", "2")
	t.Log("HGETALL h")
	result, err := rdb.HGetAll(ctx, "h").Result()
	if err != nil {
		t.Fatalf("HGETALL failed: %v", err)
	}
	t.Logf("HGETALL = %v", result)
	if len(result) != 2 || result["a"] != "1" || result["b"] != "2" {
		t.Errorf("HGETALL = %v, want map[a:1 b:2]", result)
	}
}

func TestHGetAll_Empty(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	result, err := rdb.HGetAll(ctx, "h").Result()
	if err != nil {
		t.Fatalf("HGETALL empty failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("HGETALL empty = %v, want empty map", result)
	}
}

func TestHKeys(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.HSet(ctx, "h", "a", "1", "b", "2")
	keys, err := rdb.HKeys(ctx, "h").Result()
	if err != nil {
		t.Fatalf("HKEYS failed: %v", err)
	}
	t.Logf("HKEYS = %v", keys)
	if len(keys) != 2 {
		t.Errorf("HKEYS = %v, want [a b]", keys)
	}
}

func TestHVals(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.HSet(ctx, "h", "a", "1", "b", "2")
	vals, err := rdb.HVals(ctx, "h").Result()
	if err != nil {
		t.Fatalf("HVALS failed: %v", err)
	}
	t.Logf("HVALS = %v", vals)
	if len(vals) != 2 {
		t.Errorf("HVALS = %v, want [1 2]", vals)
	}
}

func TestHLen(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.HSet(ctx, "h", "a", "1", "b", "2")
	n, err := rdb.HLen(ctx, "h").Result()
	if err != nil {
		t.Fatalf("HLEN failed: %v", err)
	}
	t.Logf("HLEN = %d", n)
	if n != 2 {
		t.Errorf("HLEN = %d, want 2", n)
	}
}

func TestHLen_Empty(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	n, err := rdb.HLen(ctx, "h").Result()
	if err != nil {
		t.Fatalf("HLEN empty failed: %v", err)
	}
	if n != 0 {
		t.Errorf("HLEN empty = %d, want 0", n)
	}
}

func TestHDel(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.HSet(ctx, "h", "a", "1", "b", "2", "c", "3")
	n, err := rdb.HDel(ctx, "h", "a", "b").Result()
	if err != nil {
		t.Fatalf("HDEL failed: %v", err)
	}
	t.Logf("HDEL = %d", n)
	if n != 2 {
		t.Errorf("HDEL = %d, want 2", n)
	}
	remaining, _ := rdb.HLen(ctx, "h").Result()
	if remaining != 1 {
		t.Errorf("HLEN after HDEL = %d, want 1", remaining)
	}
}

func TestHDel_NonExistent(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	n, err := rdb.HDel(ctx, "h", "a").Result()
	if err != nil {
		t.Fatalf("HDEL missing failed: %v", err)
	}
	if n != 0 {
		t.Errorf("HDEL missing = %d, want 0", n)
	}
}

func TestHExists(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.HSet(ctx, "h", "f", "v")
	exists, err := rdb.HExists(ctx, "h", "f").Result()
	if err != nil {
		t.Fatalf("HEXISTS failed: %v", err)
	}
	t.Logf("HEXISTS f = %t", exists)
	if !exists {
		t.Errorf("HEXISTS f = false, want true")
	}

	exists, err = rdb.HExists(ctx, "h", "nofield").Result()
	if err != nil {
		t.Fatalf("HEXISTS nofield failed: %v", err)
	}
	t.Logf("HEXISTS nofield = %t", exists)
	if exists {
		t.Errorf("HEXISTS nofield = true, want false")
	}
}

func TestHMGet(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.HSet(ctx, "h", "a", "1", "b", "2")
	vals, err := rdb.HMGet(ctx, "h", "a", "b", "c").Result()
	if err != nil {
		t.Fatalf("HMGET failed: %v", err)
	}
	t.Logf("HMGET = %v", vals)
	if len(vals) != 3 {
		t.Fatalf("HMGET len = %d, want 3", len(vals))
	}
	if v, ok := vals[0].(string); !ok || v != "1" {
		t.Errorf("HMGET[0] = %v, want 1", vals[0])
	}
	if v, ok := vals[1].(string); !ok || v != "2" {
		t.Errorf("HMGET[1] = %v, want 2", vals[1])
	}
	if vals[2] != nil {
		t.Errorf("HMGET[2] = %v, want nil", vals[2])
	}
}

func TestHMSet(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	err := rdb.HMSet(ctx, "h", map[string]interface{}{
		"a": "1",
		"b": "2",
	}).Err()
	if err != nil {
		t.Fatalf("HMSET failed: %v", err)
	}
	t.Log("HMSET h a 1 b 2")

	val, _ := rdb.HGet(ctx, "h", "a").Result()
	if val != "1" {
		t.Errorf("HGET a = %q, want 1", val)
	}
}

func TestHStrLen(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.HSet(ctx, "h", "f", "hello")
	n, err := rdb.HStrLen(ctx, "h", "f").Result()
	if err != nil {
		t.Fatalf("HSTRLEN failed: %v", err)
	}
	t.Logf("HSTRLEN = %d", n)
	if n != 5 {
		t.Errorf("HSTRLEN = %d, want 5", n)
	}
}

func TestHStrLen_NonExistent(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	n, err := rdb.HStrLen(ctx, "h", "f").Result()
	if err != nil {
		t.Fatalf("HSTRLEN missing failed: %v", err)
	}
	if n != 0 {
		t.Errorf("HSTRLEN missing = %d, want 0", n)
	}
}

func TestHScan(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.HSet(ctx, "h", "a", "1", "b", "2", "c", "3")
	keys, cursor, err := rdb.HScan(ctx, "h", 0, "", 0).Result()
	if err != nil {
		t.Fatalf("HSCAN failed: %v", err)
	}
	t.Logf("HSCAN cursor=%d keys=%v", cursor, keys)
	if cursor != 0 {
		t.Errorf("HSCAN cursor = %d, want 0", cursor)
	}
	if len(keys) != 3 {
		t.Errorf("HSCAN keys = %v, want [a b c]", keys)
	}
}

func TestHScan_WithMatch(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.HSet(ctx, "h", "abc", "1", "bcd", "2", "cde", "3")
	keys, _, err := rdb.HScan(ctx, "h", 0, "a*", 10).Result()
	if err != nil {
		t.Fatalf("HSCAN MATCH failed: %v", err)
	}
	t.Logf("HSCAN MATCH a* = %v", keys)
	if len(keys) != 1 || keys[0] != "abc" {
		t.Errorf("HSCAN MATCH a* = %v, want [abc]", keys)
	}
}

func TestHSet_UpdateExistingReturnsZero(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.HSet(ctx, "h", "f", "v1")
	n, err := rdb.HSet(ctx, "h", "f", "v2").Result()
	if err != nil {
		t.Fatalf("HSET update failed: %v", err)
	}
	t.Logf("HSET update = %d", n)
	if n != 0 {
		t.Errorf("HSET update = %d, want 0 (no new fields)", n)
	}
	val, _ := rdb.HGet(ctx, "h", "f").Result()
	if val != "v2" {
		t.Errorf("HGET after update = %q, want v2", val)
	}
}

func TestHSet_WrongType(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "s", "string", 0)
	err := rdb.HSet(ctx, "s", "f", "v").Err()
	t.Logf("HSET on string key error: %v", err)
	if err == nil || err.Error() != "WRONGTYPE Operation against a key holding the wrong kind of value" {
		t.Errorf("HSET wrong type = %v, want WRONGTYPE error", err)
	}
}

func TestHMSet_WrongType(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "s", "string", 0)
	err := rdb.HMSet(ctx, "s", map[string]interface{}{"f": "v"}).Err()
	t.Logf("HMSET on string key error: %v", err)
	if err == nil || err.Error() != "WRONGTYPE Operation against a key holding the wrong kind of value" {
		t.Errorf("HMSET wrong type = %v, want WRONGTYPE error", err)
	}
}
