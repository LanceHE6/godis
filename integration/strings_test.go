package integration

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
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

func TestBitCount_All(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// "a" = 0b01100001 = 3 个 1
	rdb.Set(ctx, "k", "a", 0)
	t.Log("SET k 'a' (01100001, 3 bits set)")

	n, err := rdb.BitCount(ctx, "k", nil).Result()
	if err != nil {
		t.Fatalf("BITCOUNT failed: %v", err)
	}
	t.Logf("BITCOUNT k = %d", n)
	if n != 3 {
		t.Errorf("BITCOUNT k = %d, want 3", n)
	}
}

func TestBitCount_Range(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// "ab" = [0x61, 0x62] — 'a' 有 3 个 1, 'b' 有 3 个 1
	rdb.Set(ctx, "k", "ab", 0)
	t.Log("SET k 'ab', then BITCOUNT k 0 0 (only first byte)")

	n, err := rdb.BitCount(ctx, "k", &redis.BitCount{Start: 0, End: 0}).Result()
	if err != nil {
		t.Fatalf("BITCOUNT range failed: %v", err)
	}
	t.Logf("BITCOUNT k 0 0 = %d (expect 3, only 'a')", n)
	if n != 3 {
		t.Errorf("BITCOUNT k 0 0 = %d, want 3", n)
	}

	// 整个字符串
	n, err = rdb.BitCount(ctx, "k", &redis.BitCount{Start: 0, End: 1}).Result()
	if err != nil {
		t.Fatalf("BITCOUNT all failed: %v", err)
	}
	t.Logf("BITCOUNT k 0 1 = %d (expect 6, 'a' + 'b')", n)
	if n != 6 {
		t.Errorf("BITCOUNT k 0 1 = %d, want 6", n)
	}
}

func TestBitCount_NonExistent(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("BITCOUNT nokey (non-existent)")
	n, err := rdb.BitCount(ctx, "nokey", nil).Result()
	if err != nil {
		t.Fatalf("BITCOUNT nokey failed: %v", err)
	}
	t.Logf("BITCOUNT nokey = %d (expect 0)", n)
	if n != 0 {
		t.Errorf("BITCOUNT nokey = %d, want 0", n)
	}
}

func TestDecr(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("DECR nokey (key does not exist, should return -1)")
	n, err := rdb.Decr(ctx, "nokey").Result()
	if err != nil {
		t.Fatalf("DECR nokey failed: %v", err)
	}
	t.Logf("DECR nokey = %d", n)
	if n != -1 {
		t.Errorf("DECR nokey = %d, want -1", n)
	}

	rdb.Set(ctx, "k", "5", 0)
	t.Log("SET k 5, then DECR k")
	n, err = rdb.Decr(ctx, "k").Result()
	if err != nil {
		t.Fatalf("DECR k failed: %v", err)
	}
	t.Logf("DECR k = %d", n)
	if n != 4 {
		t.Errorf("DECR k = %d, want 4", n)
	}
}

func TestDecrBy(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "k", "10", 0)
	t.Log("SET k 10, then DECRBY k 3")
	n, err := rdb.DecrBy(ctx, "k", 3).Result()
	if err != nil {
		t.Fatalf("DECRBY k 3 failed: %v", err)
	}
	t.Logf("DECRBY k 3 = %d", n)
	if n != 7 {
		t.Errorf("DECRBY k 3 = %d, want 7", n)
	}

	t.Log("DECRBY nokey 5 (key does not exist)")
	n, err = rdb.DecrBy(ctx, "nokey", 5).Result()
	if err != nil {
		t.Fatalf("DECRBY nokey 5 failed: %v", err)
	}
	t.Logf("DECRBY nokey 5 = %d (expect -5)", n)
	if n != -5 {
		t.Errorf("DECRBY nokey 5 = %d, want -5", n)
	}
}

func TestIncr(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("INCR nokey (key does not exist, should return 1)")
	n, err := rdb.Incr(ctx, "nokey").Result()
	if err != nil {
		t.Fatalf("INCR nokey failed: %v", err)
	}
	t.Logf("INCR nokey = %d", n)
	if n != 1 {
		t.Errorf("INCR nokey = %d, want 1", n)
	}

	rdb.Set(ctx, "k", "5", 0)
	t.Log("SET k 5, then INCR k")
	n, err = rdb.Incr(ctx, "k").Result()
	if err != nil {
		t.Fatalf("INCR k failed: %v", err)
	}
	t.Logf("INCR k = %d", n)
	if n != 6 {
		t.Errorf("INCR k = %d, want 6", n)
	}
}

func TestIncrBy(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "k", "10", 0)
	t.Log("SET k 10, then INCRBY k 7")
	n, err := rdb.IncrBy(ctx, "k", 7).Result()
	if err != nil {
		t.Fatalf("INCRBY k 7 failed: %v", err)
	}
	t.Logf("INCRBY k 7 = %d", n)
	if n != 17 {
		t.Errorf("INCRBY k 7 = %d, want 17", n)
	}
}

func TestGetRange(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "k", "Hello世界", 0)
	t.Log("SET k 'Hello世界', then GETRANGE k 0 4")
	s, err := rdb.GetRange(ctx, "k", 0, 4).Result()
	if err != nil {
		t.Fatalf("GETRANGE failed: %v", err)
	}
	t.Logf("GETRANGE k 0 4 = %q", s)
	if s != "Hello" {
		t.Errorf("GETRANGE k 0 4 = %q, want Hello", s)
	}

	// 负索引（Redis 用字节索引，"Hello世界" 是 11 字节，-3 即第 8 个字节 = "界"）
	t.Log("GETRANGE k -3 -1 (last 3 bytes, expect 界)")
	s, err = rdb.GetRange(ctx, "k", -3, -1).Result()
	if err != nil {
		t.Fatalf("GETRANGE neg failed: %v", err)
	}
	t.Logf("GETRANGE k -3 -1 = %q", s)
	if s != "界" {
		t.Errorf("GETRANGE k -3 -1 = %q, want 界", s)
	}

	// 不存在 key
	t.Log("GETRANGE nokey 0 1")
	s, err = rdb.GetRange(ctx, "nokey", 0, 1).Result()
	if err != nil {
		t.Fatalf("GETRANGE nokey failed: %v", err)
	}
	t.Logf("GETRANGE nokey = %q (expect empty)", s)
	if s != "" {
		t.Errorf("GETRANGE nokey = %q, want \"\"", s)
	}
}

func TestGetSet(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// key 不存在，返回 nil
	t.Log("GETSET nokey hello (key does not exist)")
	old, err := rdb.GetSet(ctx, "nokey", "hello").Result()
	if err != nil && err != redis.Nil {
		t.Fatalf("GETSET nokey failed: %v", err)
	}
	t.Logf("GETSET nokey old = %q (expect nil)", old)
	if err != redis.Nil {
		t.Errorf("GETSET nokey should return nil")
	}

	val, _ := rdb.Get(ctx, "nokey").Result()
	t.Logf("GET nokey = %q", val)
	if val != "hello" {
		t.Errorf("GET nokey = %q, want hello", val)
	}

	// key 已存在
	t.Log("GETSET nokey world")
	old, err = rdb.GetSet(ctx, "nokey", "world").Result()
	if err != nil {
		t.Fatalf("GETSET nokey failed: %v", err)
	}
	t.Logf("GETSET nokey old = %q (expect hello)", old)
	if old != "hello" {
		t.Errorf("GETSET nokey = %q, want hello", old)
	}

	val, _ = rdb.Get(ctx, "nokey").Result()
	t.Logf("GET nokey = %q (expect world)", val)
	if val != "world" {
		t.Errorf("GET nokey = %q, want world", val)
	}
}

func TestMSetAndMGet(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("MSET k1 v1 k2 v2 k3 v3")
	err := rdb.MSet(ctx, "k1", "v1", "k2", "v2", "k3", "v3").Err()
	if err != nil {
		t.Fatalf("MSET failed: %v", err)
	}

	t.Log("MGET k1 k2 k3 nonexistent")
	vals, err := rdb.MGet(ctx, "k1", "k2", "k3", "nonexistent").Result()
	if err != nil {
		t.Fatalf("MGET failed: %v", err)
	}
	t.Logf("MGET results: %v", vals)

	if len(vals) != 4 {
		t.Fatalf("MGET returned %d results, want 4", len(vals))
	}
	if v, ok := vals[0].(string); !ok || v != "v1" {
		t.Errorf("MGET[0] = %v, want v1", vals[0])
	}
	if v, ok := vals[1].(string); !ok || v != "v2" {
		t.Errorf("MGET[1] = %v, want v2", vals[1])
	}
	if v, ok := vals[2].(string); !ok || v != "v3" {
		t.Errorf("MGET[2] = %v, want v3", vals[2])
	}
	if vals[3] != nil {
		t.Errorf("MGET[3] = %v, want nil", vals[3])
	}
}

func TestMSet_Overwrite(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "k", "old", 0)
	t.Log("SET k old, then MSET k new")
	rdb.MSet(ctx, "k", "new")

	val, err := rdb.Get(ctx, "k").Result()
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	t.Logf("GET k = %q", val)
	if val != "new" {
		t.Errorf("GET k = %q, want new", val)
	}
}

func TestMGet_Empty(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("MGET with no keys")
	_, err := rdb.MGet(ctx).Result()
	t.Logf("MGET empty error (expected): %v", err)
	if err == nil {
		t.Error("MGET empty should return error")
	}
}

func TestStrLen(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// 不存在 key 返回 0
	n, err := rdb.StrLen(ctx, "nokey").Result()
	if err != nil {
		t.Fatalf("STRLEN nokey failed: %v", err)
	}
	t.Logf("STRLEN nokey = %d", n)
	if n != 0 {
		t.Errorf("STRLEN nokey = %d, want 0", n)
	}

	// 字符串长度
	rdb.Set(ctx, "k", "hello", 0)
	n, err = rdb.StrLen(ctx, "k").Result()
	if err != nil {
		t.Fatalf("STRLEN k failed: %v", err)
	}
	t.Logf("STRLEN k = %d", n)
	if n != 5 {
		t.Errorf("STRLEN k = %d, want 5", n)
	}

	// 空字符串
	rdb.Set(ctx, "empty", "", 0)
	n, err = rdb.StrLen(ctx, "empty").Result()
	if err != nil {
		t.Fatalf("STRLEN empty failed: %v", err)
	}
	t.Logf("STRLEN empty = %d", n)
	if n != 0 {
		t.Errorf("STRLEN empty = %d, want 0", n)
	}

	// 中文 UTF-8 字节数
	rdb.Set(ctx, "zh", "世界", 0)
	n, err = rdb.StrLen(ctx, "zh").Result()
	if err != nil {
		t.Fatalf("STRLEN zh failed: %v", err)
	}
	t.Logf("STRLEN zh = %d", n)
	if n != 6 {
		t.Errorf("STRLEN zh = %d, want 6 (UTF-8 bytes)", n)
	}
}
