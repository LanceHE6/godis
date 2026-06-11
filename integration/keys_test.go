package integration

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestDel(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "a", "1", 0)
	rdb.Set(ctx, "b", "2", 0)
	rdb.Set(ctx, "c", "3", 0)

	t.Log("DEL a (single key)")
	n, err := rdb.Del(ctx, "a").Result()
	if err != nil {
		t.Fatalf("DEL a failed: %v", err)
	}
	t.Logf("DEL a returned %d", n)
	if n != 1 {
		t.Errorf("DEL a = %d, want 1", n)
	}

	t.Log("DEL b, c, nokey (multiple keys)")
	n, err = rdb.Del(ctx, "b", "c", "nokey").Result()
	if err != nil {
		t.Fatalf("DEL b,c failed: %v", err)
	}
	t.Logf("DEL b,c,nokey returned %d", n)
	if n != 2 {
		t.Errorf("DEL b,c,nokey = %d, want 2", n)
	}
}

func TestExists(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "x", "1", 0)
	rdb.Set(ctx, "y", "2", 0)

	t.Log("EXISTS x, y, z")
	n, err := rdb.Exists(ctx, "x", "y", "z").Result()
	if err != nil {
		t.Fatalf("EXISTS failed: %v", err)
	}
	t.Logf("EXISTS x,y,z = %d", n)
	if n != 2 {
		t.Errorf("EXISTS x,y,z = %d, want 2", n)
	}
}

func TestExpireTTL(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "k", "v", 0)
	t.Log("EXPIRE k 10")

	ok, err := rdb.Expire(ctx, "k", 10*time.Second).Result()
	if err != nil {
		t.Fatalf("EXPIRE failed: %v", err)
	}
	t.Logf("EXPIRE returned %v", ok)
	if !ok {
		t.Error("EXPIRE should return true for existing key")
	}

	ttl, err := rdb.TTL(ctx, "k").Result()
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	t.Logf("TTL k = %v", ttl)
	if ttl <= 0 || ttl > 10*time.Second {
		t.Errorf("TTL k = %v, want > 0 and <= 10s", ttl)
	}
}

func TestPExpirePTTL(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "k", "v", 0)
	t.Log("PEXPIRE k 5000")

	ok, err := rdb.Do(ctx, "PEXPIRE", "k", "5000").Bool()
	if err != nil {
		t.Fatalf("PEXPIRE failed: %v", err)
	}
	t.Logf("PEXPIRE returned %v", ok)
	if !ok {
		t.Error("PEXPIRE should return true")
	}

	pttl, err := rdb.Do(ctx, "PTTL", "k").Int64()
	if err != nil {
		t.Fatalf("PTTL failed: %v", err)
	}
	t.Logf("PTTL k = %dms", pttl)
	if pttl <= 0 || pttl > 5000 {
		t.Errorf("PTTL k = %d, want > 0 and <= 5000", pttl)
	}
}

func TestPersist(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "k", "v", 10*time.Second)
	t.Log("SET k v EX 10, then PERSIST k")

	ok, err := rdb.Persist(ctx, "k").Result()
	if err != nil {
		t.Fatalf("PERSIST failed: %v", err)
	}
	t.Logf("PERSIST returned %v", ok)
	if !ok {
		t.Error("PERSIST should return true")
	}

	ttl, err := rdb.TTL(ctx, "k").Result()
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	t.Logf("TTL after PERSIST = %d", ttl)
	if ttl != -1 {
		t.Errorf("TTL after PERSIST = %d, want -1 (no expiry)", ttl)
	}
}

func TestMove(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "moveme", "val", 0)
	t.Log("SET moveme val, then MOVE moveme 1")

	n, err := rdb.Do(ctx, "MOVE", "moveme", "1").Int64()
	if err != nil {
		t.Fatalf("MOVE failed: %v", err)
	}
	t.Logf("MOVE returned %d", n)
	if n != 1 {
		t.Errorf("MOVE = %d, want 1", n)
	}

	exists, _ := rdb.Exists(ctx, "moveme").Result()
	t.Logf("EXISTS moveme in db0 = %d (expect 0)", exists)
	if exists != 0 {
		t.Error("moveme should not exist in db0 after MOVE")
	}

	rdb.Do(ctx, "SELECT", "1")
	val, err := rdb.Get(ctx, "moveme").Result()
	if err != nil {
		t.Fatalf("GET moveme in db1 failed: %v", err)
	}
	t.Logf("GET moveme in db1 = %q", val)
	if val != "val" {
		t.Errorf("GET moveme in db1 = %q, want val", val)
	}
	rdb.Do(ctx, "SELECT", "0")
}

func TestType(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "str", "hello", 0)

	t.Log("TYPE str")
	typ, err := rdb.Type(ctx, "str").Result()
	if err != nil {
		t.Fatalf("TYPE failed: %v", err)
	}
	t.Logf("TYPE str = %q", typ)
	if typ != "string" {
		t.Errorf("TYPE str = %q, want string", typ)
	}

	t.Log("TYPE nokey (non-existent)")
	typ, err = rdb.Type(ctx, "nokey").Result()
	if err != nil {
		t.Fatalf("TYPE nokey failed: %v", err)
	}
	t.Logf("TYPE nokey = %q", typ)
	if typ != "none" {
		t.Errorf("TYPE nokey = %q, want none", typ)
	}
}

func TestTouch(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "a", "1", 0)
	rdb.Set(ctx, "b", "2", 0)

	t.Log("TOUCH a, b, c (c does not exist)")
	n, err := rdb.Do(ctx, "TOUCH", "a", "b", "c").Int64()
	if err != nil {
		t.Fatalf("TOUCH failed: %v", err)
	}
	t.Logf("TOUCH returned %d", n)
	if n != 2 {
		t.Errorf("TOUCH a,b,c = %d, want 2", n)
	}
}

func TestUnlink(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "a", "1", 0)
	rdb.Set(ctx, "b", "2", 0)

	t.Log("UNLINK a, b, c (c does not exist)")
	n, err := rdb.Do(ctx, "UNLINK", "a", "b", "c").Int64()
	if err != nil {
		t.Fatalf("UNLINK failed: %v", err)
	}
	t.Logf("UNLINK returned %d", n)
	if n != 2 {
		t.Errorf("UNLINK a,b,c = %d, want 2", n)
	}

	exists, _ := rdb.Exists(ctx, "a", "b").Result()
	t.Logf("EXISTS a,b after UNLINK = %d (expect 0)", exists)
	if exists != 0 {
		t.Error("a and b should not exist after UNLINK")
	}
}

func TestDBSize(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "a", "1", 0)
	rdb.Set(ctx, "b", "2", 0)

	t.Log("DBSIZE after SET a, b")
	n, err := rdb.Do(ctx, "DBSIZE").Int64()
	if err != nil {
		t.Fatalf("DBSIZE failed: %v", err)
	}
	t.Logf("DBSIZE = %d", n)
	if n != 2 {
		t.Errorf("DBSIZE = %d, want 2", n)
	}
}

func TestFlushDB(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "a", "1", 0)
	rdb.Set(ctx, "b", "2", 0)
	t.Log("SET a, b, then FLUSHDB")
	rdb.Do(ctx, "FLUSHDB")

	n, _ := rdb.Do(ctx, "DBSIZE").Int64()
	t.Logf("DBSIZE after FLUSHDB = %d", n)
	if n != 0 {
		t.Errorf("DBSIZE after FLUSHDB = %d, want 0", n)
	}
}

func TestFlushAll(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "a", "1", 0)
	rdb.Do(ctx, "SELECT", "1")
	rdb.Set(ctx, "b", "2", 0)
	rdb.Do(ctx, "SELECT", "0")
	t.Log("SET a in db0, SET b in db1, then FLUSHALL")
	rdb.Do(ctx, "FLUSHALL")

	n0, _ := rdb.Do(ctx, "DBSIZE").Int64()
	rdb.Do(ctx, "SELECT", "1")
	n1, _ := rdb.Do(ctx, "DBSIZE").Int64()
	rdb.Do(ctx, "SELECT", "0")
	t.Logf("DBSIZE after FLUSHALL: db0=%d, db1=%d", n0, n1)
	if n0 != 0 || n1 != 0 {
		t.Errorf("DBSIZE after FLUSHALL: db0=%d, db1=%d, want 0,0", n0, n1)
	}
}

func TestSort(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("SORT nokey (non-existent)")
	res, err := rdb.Do(ctx, "SORT", "nokey").StringSlice()
	if err != nil {
		t.Fatalf("SORT nokey failed: %v", err)
	}
	t.Logf("SORT nokey = %v (expect empty)", res)
	if len(res) != 0 {
		t.Errorf("SORT nokey = %v, want empty", res)
	}

	rdb.Set(ctx, "str", "val", 0)
	t.Log("SORT str (string type, expect WRONGTYPE)")
	_, err = rdb.Do(ctx, "SORT", "str").Result()
	if err == nil {
		t.Error("SORT on string key should return WRONGTYPE error")
	}
	t.Logf("SORT str returned error (expected): %v", err)
}

func TestScan(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("creating 20 keys for SCAN test")
	for i := 0; i < 20; i++ {
		rdb.Set(ctx, fmt.Sprintf("scankey%d", i), "v", 0)
	}

	var allKeys []string
	var cursor uint64
	round := 0
	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, "", 5).Result()
		if err != nil {
			t.Fatalf("SCAN failed: %v", err)
		}
		round++
		t.Logf("SCAN round %d: cursor=%d, got %d keys: %v", round, cursor, len(keys), keys)
		allKeys = append(allKeys, keys...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	t.Logf("SCAN completed: %d rounds, total %d keys", round, len(allKeys))
	if len(allKeys) < 20 {
		t.Errorf("SCAN collected %d keys, want at least 20", len(allKeys))
	}
}

func TestScan_Match(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "abc", "1", 0)
	rdb.Set(ctx, "abd", "2", 0)
	rdb.Set(ctx, "xyz", "3", 0)
	t.Log("SET abc, abd, xyz; SCAN MATCH ab*")

	var allKeys []string
	var cursor uint64
	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, "ab*", 100).Result()
		if err != nil {
			t.Fatalf("SCAN MATCH failed: %v", err)
		}
		allKeys = append(allKeys, keys...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	t.Logf("SCAN MATCH ab* returned: %v", allKeys)
	found := map[string]bool{}
	for _, k := range allKeys {
		found[k] = true
	}

	if !found["abc"] || !found["abd"] {
		t.Errorf("SCAN MATCH ab* should include abc and abd, got %v", allKeys)
	}
	if found["xyz"] {
		t.Error("SCAN MATCH ab* should not include xyz")
	}
}
