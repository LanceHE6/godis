package integration

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestZAddAndZScore(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z", redis.Z{Score: 1.5, Member: "a"})
	score, err := rdb.ZScore(ctx, "z", "a").Result()
	if err != nil {
		t.Fatalf("ZSCORE failed: %v", err)
	}
	t.Logf("ZSCORE a = %f", score)
	if score != 1.5 {
		t.Errorf("ZSCORE = %f, want 1.5", score)
	}
}

func TestZAddMultiple(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	n, err := rdb.ZAdd(ctx, "z",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
		redis.Z{Score: 3, Member: "c"},
	).Result()
	if err != nil {
		t.Fatalf("ZADD failed: %v", err)
	}
	if n != 3 {
		t.Errorf("ZADD = %d, want 3", n)
	}
}

func TestZCard(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z", redis.Z{Score: 1, Member: "a"})
	n, err := rdb.ZCard(ctx, "z").Result()
	if err != nil {
		t.Fatalf("ZCARD failed: %v", err)
	}
	if n != 1 {
		t.Errorf("ZCARD = %d, want 1", n)
	}
}

func TestZCount(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
		redis.Z{Score: 3, Member: "c"},
	)
	n, err := rdb.ZCount(ctx, "z", "1", "2").Result()
	if err != nil {
		t.Fatalf("ZCOUNT failed: %v", err)
	}
	if n != 2 {
		t.Errorf("ZCOUNT 1-2 = %d, want 2", n)
	}
}

func TestZIncrBy(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z", redis.Z{Score: 1, Member: "a"})
	newScore, err := rdb.ZIncrBy(ctx, "z", 2.5, "a").Result()
	if err != nil {
		t.Fatalf("ZINCRBY failed: %v", err)
	}
	t.Logf("ZINCRBY = %f", newScore)
	if newScore != 3.5 {
		t.Errorf("ZINCRBY = %f, want 3.5", newScore)
	}
}

func TestZRank(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 10, Member: "a"},
		redis.Z{Score: 20, Member: "b"},
	)
	rank, err := rdb.ZRank(ctx, "z", "a").Result()
	if err != nil {
		t.Fatalf("ZRANK failed: %v", err)
	}
	if rank != 0 {
		t.Errorf("ZRANK a = %d, want 0", rank)
	}
}

func TestZRevRank(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 10, Member: "a"},
		redis.Z{Score: 20, Member: "b"},
	)
	rank, err := rdb.ZRevRank(ctx, "z", "a").Result()
	if err != nil {
		t.Fatalf("ZREVRANK failed: %v", err)
	}
	if rank != 1 {
		t.Errorf("ZREVRANK a = %d, want 1 (b is higher)", rank)
	}
}

func TestZRem(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
	)
	n, err := rdb.ZRem(ctx, "z", "a").Result()
	if err != nil {
		t.Fatalf("ZREM failed: %v", err)
	}
	if n != 1 {
		t.Errorf("ZREM = %d, want 1", n)
	}
}

func TestZPopMax(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 3, Member: "c"},
		redis.Z{Score: 2, Member: "b"},
	)
	result, err := rdb.ZPopMax(ctx, "z", 1).Result()
	if err != nil {
		t.Fatalf("ZPOPMAX failed: %v", err)
	}
	t.Logf("ZPOPMAX = %v", result)
	if len(result) != 1 || result[0].Member != "c" {
		t.Errorf("ZPOPMAX = %v, want [{c 3}]", result)
	}
}

func TestZPopMin(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 3, Member: "c"},
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
	)
	result, err := rdb.ZPopMin(ctx, "z", 1).Result()
	if err != nil {
		t.Fatalf("ZPOPMIN failed: %v", err)
	}
	t.Logf("ZPOPMIN = %v", result)
	if len(result) != 1 || result[0].Member != "a" {
		t.Errorf("ZPOPMIN = %v, want [{a 1}]", result)
	}
}

func TestZRange(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
		redis.Z{Score: 3, Member: "c"},
	)
	result, err := rdb.ZRange(ctx, "z", 0, -1).Result()
	if err != nil {
		t.Fatalf("ZRANGE failed: %v", err)
	}
	t.Logf("ZRANGE = %v", result)
	if len(result) != 3 || result[0] != "a" || result[2] != "c" {
		t.Errorf("ZRANGE = %v, want [a b c]", result)
	}
}

func TestZRangeWithScores(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
	)
	result, err := rdb.ZRangeWithScores(ctx, "z", 0, -1).Result()
	if err != nil {
		t.Fatalf("ZRANGE WITHSCORES failed: %v", err)
	}
	t.Logf("ZRANGE WITHSCORES = %v", result)
	if len(result) != 2 || result[0].Member != "a" || result[0].Score != 1 {
		t.Errorf("ZRANGE WITHSCORES = %v, want [{a 1} {b 2}]", result)
	}
}

func TestZRangeRev(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
	)
	result, err := rdb.ZRevRange(ctx, "z", 0, -1).Result()
	if err != nil {
		t.Fatalf("ZREVRANGE failed: %v", err)
	}
	t.Logf("ZREVRANGE = %v", result)
	if len(result) != 2 || result[0] != "b" {
		t.Errorf("ZREVRANGE = %v, want [b a]", result)
	}
}

func TestZRangeByScore(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
		redis.Z{Score: 3, Member: "c"},
	)
	result, err := rdb.ZRangeByScore(ctx, "z", &redis.ZRangeBy{
		Min: "1", Max: "2",
	}).Result()
	if err != nil {
		t.Fatalf("ZRANGEBYSCORE failed: %v", err)
	}
	t.Logf("ZRANGEBYSCORE = %v", result)
	if len(result) != 2 || result[0] != "a" {
		t.Errorf("ZRANGEBYSCORE = %v, want [a b]", result)
	}
}

func TestZRangeByLex(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 0, Member: "a"},
		redis.Z{Score: 0, Member: "b"},
		redis.Z{Score: 0, Member: "c"},
	)
	result, err := rdb.ZRangeByLex(ctx, "z", &redis.ZRangeBy{
		Min: "-", Max: "+",
	}).Result()
	if err != nil {
		t.Fatalf("ZRANGEBYLEX failed: %v", err)
	}
	t.Logf("ZRANGEBYLEX = %v", result)
	if len(result) != 3 {
		t.Errorf("ZRANGEBYLEX = %v, want [a b c]", result)
	}
}

func TestZRemRangeByScore(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
		redis.Z{Score: 3, Member: "c"},
	)
	n, err := rdb.ZRemRangeByScore(ctx, "z", "1", "2").Result()
	if err != nil {
		t.Fatalf("ZREMRANGEBYSCORE failed: %v", err)
	}
	if n != 2 {
		t.Errorf("ZREMRANGEBYSCORE = %d, want 2", n)
	}
}

func TestZRemRangeByRank(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
		redis.Z{Score: 3, Member: "c"},
	)
	n, err := rdb.ZRemRangeByRank(ctx, "z", 0, 1).Result()
	if err != nil {
		t.Fatalf("ZREMRANGEBYRANK failed: %v", err)
	}
	if n != 2 {
		t.Errorf("ZREMRANGEBYRANK = %d, want 2", n)
	}
}

func TestZRemRangeByLex(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 0, Member: "a"},
		redis.Z{Score: 0, Member: "b"},
		redis.Z{Score: 0, Member: "c"},
	)
	n, err := rdb.ZRemRangeByLex(ctx, "z", "[a", "[b").Result()
	if err != nil {
		t.Fatalf("ZREMRANGEBYLEX failed: %v", err)
	}
	if n != 2 {
		t.Errorf("ZREMRANGEBYLEX = %d, want 2", n)
	}
}

func TestZMScore(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
	)
	scores, err := rdb.ZMScore(ctx, "z", "a", "b", "missing").Result()
	if err != nil {
		t.Fatalf("ZMSCORE failed: %v", err)
	}
	t.Logf("ZMSCORE = %v", scores)
	if len(scores) != 3 || scores[0] != 1 || scores[1] != 2 {
		t.Errorf("ZMSCORE = %v, want [1 2 <nil>]", scores)
	}
	// missing key returns NaN in go-redis
	if !isNaN(scores[2]) {
		t.Logf("ZMSCORE missing = %f (expected NaN)", scores[2])
	}
}

func isNaN(f float64) bool { return f != f }

func TestZUnion(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z1", redis.Z{Score: 1, Member: "a"}, redis.Z{Score: 2, Member: "b"})
	rdb.ZAdd(ctx, "z2", redis.Z{Score: 3, Member: "b"}, redis.Z{Score: 4, Member: "c"})
	result, err := rdb.ZUnion(ctx, redis.ZStore{Keys: []string{"z1", "z2"}}).Result()
	if err != nil {
		t.Fatalf("ZUNION failed: %v", err)
	}
	t.Logf("ZUNION = %v", result)
	if len(result) != 3 {
		t.Errorf("ZUNION = %v, want 3 members", result)
	}
}

func TestZInter(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z1", redis.Z{Score: 1, Member: "a"}, redis.Z{Score: 2, Member: "b"})
	rdb.ZAdd(ctx, "z2", redis.Z{Score: 3, Member: "b"}, redis.Z{Score: 4, Member: "c"})
	result, err := rdb.ZInter(ctx, &redis.ZStore{Keys: []string{"z1", "z2"}}).Result()
	if err != nil {
		t.Fatalf("ZINTER failed: %v", err)
	}
	t.Logf("ZINTER = %v", result)
	if len(result) != 1 || result[0] != "b" {
		t.Errorf("ZINTER = %v, want [b]", result)
	}
}

func TestZScan(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z", redis.Z{Score: 1, Member: "a"}, redis.Z{Score: 2, Member: "b"})
	keys, cursor, err := rdb.ZScan(ctx, "z", 0, "", 0).Result()
	if err != nil {
		t.Fatalf("ZSCAN failed: %v", err)
	}
	t.Logf("ZSCAN cursor=%d keys=%v", cursor, keys)
	if cursor != 0 {
		t.Errorf("ZSCAN cursor = %d, want 0", cursor)
	}
	if len(keys) != 2 {
		t.Errorf("ZSCAN keys = %v, want [a b]", keys)
	}
}

func TestZLexCount(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.ZAdd(ctx, "z",
		redis.Z{Score: 0, Member: "a"},
		redis.Z{Score: 0, Member: "b"},
		redis.Z{Score: 0, Member: "c"},
	)
	n, err := rdb.ZLexCount(ctx, "z", "-", "+").Result()
	if err != nil {
		t.Fatalf("ZLEXCOUNT failed: %v", err)
	}
	if n != 3 {
		t.Errorf("ZLEXCOUNT = %d, want 3", n)
	}
}
