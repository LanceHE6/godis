package integration

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestLPushAndLRange(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.LPush(ctx, "k", "c", "b", "a")
	vals, err := rdb.LRange(ctx, "k", 0, -1).Result()
	if err != nil {
		t.Fatalf("LRANGE failed: %v", err)
	}
	t.Logf("LRANGE = %v", vals)
	if len(vals) != 3 || vals[0] != "a" || vals[2] != "c" {
		t.Errorf("LPush order wrong: %v, want [a b c]", vals)
	}
}

func TestRPush(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.RPush(ctx, "k", "a", "b", "c")
	n, err := rdb.LLen(ctx, "k").Result()
	if err != nil {
		t.Fatalf("LLEN failed: %v", err)
	}
	if n != 3 {
		t.Errorf("LLEN = %d, want 3", n)
	}
}

func TestLPushX(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	n, err := rdb.LPushX(ctx, "k", "a").Result()
	if err != nil {
		t.Fatalf("LPUSHX failed: %v", err)
	}
	t.Logf("LPUSHX on missing key = %d", n)
	if n != 0 {
		t.Errorf("LPUSHX missing = %d, want 0", n)
	}

	rdb.LPush(ctx, "k", "x")
	n, err = rdb.LPushX(ctx, "k", "y").Result()
	if err != nil {
		t.Fatalf("LPUSHX on existing key failed: %v", err)
	}
	if n != 2 {
		t.Errorf("LPUSHX existing = %d, want 2", n)
	}
}

func TestLPop(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.RPush(ctx, "k", "a", "b", "c")
	val, err := rdb.LPop(ctx, "k").Result()
	if err != nil {
		t.Fatalf("LPOP failed: %v", err)
	}
	t.Logf("LPOP = %q", val)
	if val != "a" {
		t.Errorf("LPOP = %q, want a", val)
	}
}

func TestRPop(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.RPush(ctx, "k", "a", "b")
	val, err := rdb.RPop(ctx, "k").Result()
	if err != nil {
		t.Fatalf("RPOP failed: %v", err)
	}
	t.Logf("RPOP = %q", val)
	if val != "b" {
		t.Errorf("RPOP = %q, want b", val)
	}
}

func TestLIndex(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.RPush(ctx, "k", "a", "b", "c")
	val, err := rdb.LIndex(ctx, "k", 1).Result()
	if err != nil {
		t.Fatalf("LINDEX failed: %v", err)
	}
	t.Logf("LINDEX 1 = %q", val)
	if val != "b" {
		t.Errorf("LINDEX = %q, want b", val)
	}
}

func TestLInsert(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.RPush(ctx, "k", "a", "c")
	n, err := rdb.LInsert(ctx, "k", "BEFORE", "c", "b").Result()
	if err != nil {
		t.Fatalf("LINSERT failed: %v", err)
	}
	t.Logf("LINSERT = %d", n)
	if n != 3 {
		t.Errorf("LINSERT = %d, want 3", n)
	}
}

func TestLRem(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.RPush(ctx, "k", "a", "b", "a", "c")
	n, err := rdb.LRem(ctx, "k", 1, "a").Result()
	if err != nil {
		t.Fatalf("LREM failed: %v", err)
	}
	t.Logf("LREM = %d", n)
	if n != 1 {
		t.Errorf("LREM = %d, want 1", n)
	}
}

func TestLSet(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.RPush(ctx, "k", "a", "b", "c")
	err := rdb.LSet(ctx, "k", 1, "X").Err()
	if err != nil {
		t.Fatalf("LSET failed: %v", err)
	}
	val, _ := rdb.LIndex(ctx, "k", 1).Result()
	if val != "X" {
		t.Errorf("LSET result = %q, want X", val)
	}
}

func TestLPos(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.RPush(ctx, "k", "a", "b", "a")
	pos, err := rdb.LPos(ctx, "k", "a", redis.LPosArgs{Rank: 2}).Result()
	if err != nil {
		t.Fatalf("LPOS failed: %v", err)
	}
	t.Logf("LPOS a RANK 2 = %d", pos)
	if pos != 2 {
		t.Errorf("LPOS = %d, want 2", pos)
	}
}

func TestLMove(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.RPush(ctx, "src", "a", "b", "c")
	val, err := rdb.LMove(ctx, "src", "dst", "LEFT", "RIGHT").Result()
	if err != nil {
		t.Fatalf("LMOVE failed: %v", err)
	}
	t.Logf("LMOVE = %q", val)
	if val != "a" {
		t.Errorf("LMOVE = %q, want a", val)
	}
	n, _ := rdb.LLen(ctx, "dst").Result()
	if n != 1 {
		t.Errorf("LMOVE dst len = %d, want 1", n)
	}
}

func TestBLMove(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.RPush(ctx, "src", "x")
	val, err := rdb.BLMove(ctx, "src", "dst", "LEFT", "RIGHT", 0).Result()
	if err != nil {
		t.Fatalf("BLMOVE failed: %v", err)
	}
	t.Logf("BLMOVE = %q", val)
	if val != "x" {
		t.Errorf("BLMOVE = %q, want x", val)
	}
}

func TestBLPop(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.RPush(ctx, "k", "a")
	vals, err := rdb.BLPop(ctx, 0, "k").Result()
	if err != nil {
		t.Fatalf("BLPOP failed: %v", err)
	}
	t.Logf("BLPOP = %v", vals)
	if len(vals) != 2 || vals[0] != "k" || vals[1] != "a" {
		t.Errorf("BLPOP = %v, want [k a]", vals)
	}
}

func TestBLPop_Empty(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	_, err := rdb.BLPop(ctx, 0, "k").Result()
	if err != redis.Nil {
		t.Errorf("BLPOP empty should return nil, got: %v", err)
	}
}

func TestBRPop(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.RPush(ctx, "k", "a", "b")
	vals, err := rdb.BRPop(ctx, 0, "k").Result()
	if err != nil {
		t.Fatalf("BRPOP failed: %v", err)
	}
	t.Logf("BRPOP = %v", vals)
	if vals[1] != "b" {
		t.Errorf("BRPOP = %v, want [k b]", vals)
	}
}
