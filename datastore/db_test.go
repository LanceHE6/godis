package datastore

import (
	"godis/types"
	"testing"
	"time"
)

func TestNewGodisDB(t *testing.T) {
	db := NewGodisDB()
	if db == nil {
		t.Fatal("NewGodisDB returned nil")
	}
	if len(db.Keys()) != 0 {
		t.Errorf("new db Keys() len = %d, want 0", len(db.Keys()))
	}
}

func TestGodisDB_SetGet(t *testing.T) {
	db := NewGodisDB()
	db.Set("key", "value", 0)

	val, ok := db.Get("key")
	if !ok || val != "value" {
		t.Errorf("Get(key) = (%q, %v), want (value, true)", val, ok)
	}
}

func TestGodisDB_GetNotExist(t *testing.T) {
	db := NewGodisDB()
	val, ok := db.Get("missing")
	if ok || val != "" {
		t.Errorf("Get(missing) = (%q, %v), want (, false)", val, ok)
	}
}

func TestGodisDB_SetOverwrite(t *testing.T) {
	db := NewGodisDB()
	db.Set("k", "v1", 0)
	db.Set("k", "v2", 0)

	val, _ := db.Get("k")
	if val != "v2" {
		t.Errorf("Get(k) = %q, want v2", val)
	}
}

func TestGodisDB_SetWithTTL(t *testing.T) {
	db := NewGodisDB()
	db.Set("k", "v", 1) // 1秒过期

	val, ok := db.Get("k")
	if !ok || val != "v" {
		t.Errorf("Get(k) before expire = (%q, %v), want (v, true)", val, ok)
	}

	// 模拟过期（直接修改 ExpiresAt）
	db.mu.Lock()
	item := db.data["k"]
	item.ExpiresAt = time.Now().Add(-1 * time.Second)
	db.data["k"] = item
	db.mu.Unlock()

	val, ok = db.Get("k")
	if ok {
		t.Errorf("Get(k) after expire = (%q, %v), want (, false)", val, ok)
	}
}

func TestGodisDB_Del(t *testing.T) {
	db := NewGodisDB()
	db.Set("a", "1", 0)
	db.Set("b", "2", 0)

	deleted := db.Del("a", "c")
	if deleted != 1 {
		t.Errorf("Del(a,c) = %d, want 1", deleted)
	}
	if _, ok := db.Get("a"); ok {
		t.Error("a should be deleted")
	}
}

func TestGodisDB_TypeOf(t *testing.T) {
	db := NewGodisDB()
	db.Set("str", "val", 0)

	if db.TypeOf("str") != types.TypeString {
		t.Errorf("TypeOf(str) = %d, want %d", db.TypeOf("str"), types.TypeString)
	}
	if db.TypeOf("missing") != -1 {
		t.Errorf("TypeOf(missing) = %d, want -1", db.TypeOf("missing"))
	}
}

func TestGodisDB_Stats(t *testing.T) {
	db := NewGodisDB()
	db.Set("a", "1", 0)    // 永不过期
	db.Set("b", "2", 3600) // 有过期时间

	stats := db.Stats()
	if stats.Keys != 2 {
		t.Errorf("Stats().Keys = %d, want 2", stats.Keys)
	}
	if stats.Expires != 1 {
		t.Errorf("Stats().Expires = %d, want 1", stats.Expires)
	}
}

func TestGodisDB_Keys(t *testing.T) {
	db := NewGodisDB()
	db.Set("a", "1", 0)
	db.Set("b", "2", 0)

	keys := db.Keys()
	if len(keys) != 2 {
		t.Errorf("Keys() len = %d, want 2", len(keys))
	}
}

func TestGodisDB_GetItem(t *testing.T) {
	db := NewGodisDB()
	db.Set("k", "v", 0)

	item, ok := db.GetItem("k")
	if !ok {
		t.Fatal("GetItem(k) = nil, want item")
	}
	if item.Type != types.TypeString {
		t.Errorf("item.Type = %d, want %d", item.Type, types.TypeString)
	}
	if !item.IsNeverDie {
		t.Error("item.IsNeverDie = false, want true")
	}
}
