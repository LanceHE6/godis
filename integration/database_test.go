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

func TestSelect_MultiDBIsolation(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// DB 0: 写入两个 key
	t.Log("DB 0: SET k1 a, SET k2 b")
	rdb.Set(ctx, "k1", "a", 0)
	rdb.Set(ctx, "k2", "b", 0)

	// 切换到 DB 1
	t.Log("SELECT 1")
	rdb.Do(ctx, "SELECT", "1")

	// DB 1: 写入一个 key
	t.Log("DB 1: SET k3 c")
	rdb.Set(ctx, "k3", "c", 0)

	// 切回 DB 0，验证原来的两个 key 仍然存在
	t.Log("SELECT 0, 验证 k1 和 k2 仍在")
	rdb.Do(ctx, "SELECT", "0")

	v1, err := rdb.Get(ctx, "k1").Result()
	if err != nil {
		t.Fatalf("DB0 GET k1 failed: %v", err)
	}
	if v1 != "a" {
		t.Errorf("DB0 k1 = %q, want a", v1)
	}

	v2, err := rdb.Get(ctx, "k2").Result()
	if err != nil {
		t.Fatalf("DB0 GET k2 failed: %v", err)
	}
	if v2 != "b" {
		t.Errorf("DB0 k2 = %q, want b", v2)
	}

	// 验证 k3 在 DB 0 中不存在（多数据库隔离）
	_, err = rdb.Get(ctx, "k3").Result()
	if err == nil {
		t.Error("DB0 should not have k3 (multi-db isolation)")
	}
	t.Logf("DB0 GET k3: not found (isolation verified)")

	// 切回 DB 1，验证 k3 存在
	t.Log("SELECT 1, 验证 k3")
	rdb.Do(ctx, "SELECT", "1")
	v3, err := rdb.Get(ctx, "k3").Result()
	if err != nil {
		t.Fatalf("DB1 GET k3 failed: %v", err)
	}
	if v3 != "c" {
		t.Errorf("DB1 k3 = %q, want c", v3)
	}

	rdb.Do(ctx, "SELECT", "0")
}
