package commands

import (
	"testing"

	"godis/datastore"
	"godis/types"
)

// ──── HSET ────

func TestHSet_NewKey(t *testing.T) {
	db := mkDB()
	reply := handleHSet(mkCtx(db, "HSET", "h", "f1", "v1"))
	if reply != ":1\r\n" {
		t.Fatalf("HSET new = %q, want :1", reply)
	}
}

func TestHSet_Multiple(t *testing.T) {
	db := mkDB()
	reply := handleHSet(mkCtx(db, "HSET", "h", "f1", "v1", "f2", "v2"))
	if reply != ":2\r\n" {
		t.Fatalf("HSET multi = %q, want :2", reply)
	}
}

func TestHSet_UpdateExisting(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "f", "v1"))
	reply := handleHSet(mkCtx(db, "HSET", "h", "f", "v2"))
	if reply != ":0\r\n" {
		t.Fatalf("HSET update = %q, want :0 (no new fields)", reply)
	}
}

func TestHSet_WrongType(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "h", "str"))
	reply := handleHSet(mkCtx(db, "HSET", "h", "f", "v"))
	expected := "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
	if reply != expected {
		t.Fatalf("HSET wrong type\nexpected: %q\n     got: %q", expected, reply)
	}
}

// ──── HMSET ────

func TestHMSet(t *testing.T) {
	db := mkDB()
	reply := handleHMSet(mkCtx(db, "HMSET", "h", "f1", "v1", "f2", "v2"))
	if reply != "+OK\r\n" {
		t.Fatalf("HMSET = %q, want +OK", reply)
	}
	hv, _ := getHashItemRead(mkCtx(db, "HGET", "h"), "h")
	if hv == nil || hv.Len() != 2 {
		t.Fatalf("HMSET: expected 2 fields, got %d", hv.Len())
	}
}

// ──── HGET ────

func TestHGet(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "f", "hello"))
	reply := handleHGet(mkCtx(db, "HGET", "h", "f"))
	if reply != "$5\r\nhello\r\n" {
		t.Fatalf("HGET = %q, want $5\\r\\nhello", reply)
	}
}

func TestHGet_NonExistentKey(t *testing.T) {
	db := mkDB()
	reply := handleHGet(mkCtx(db, "HGET", "h", "f"))
	if reply != "$-1\r\n" {
		t.Fatalf("HGET missing key = %q, want nil", reply)
	}
}

func TestHGet_NonExistentField(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "f", "v"))
	reply := handleHGet(mkCtx(db, "HGET", "h", "nofield"))
	if reply != "$-1\r\n" {
		t.Fatalf("HGET missing field = %q, want nil", reply)
	}
}

func TestHGet_WrongType(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "h", "str"))
	reply := handleHGet(mkCtx(db, "HGET", "h", "f"))
	expected := "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
	if reply != expected {
		t.Fatalf("HGET wrong type\nexpected: %q\n     got: %q", expected, reply)
	}
}

// ──── HGETALL ────

func TestHGetAll(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "a", "1", "b", "2"))
	reply := handleHGetAll(mkCtx(db, "HGETALL", "h"))
	expected := "*4\r\n$1\r\na\r\n$1\r\n1\r\n$1\r\nb\r\n$1\r\n2\r\n"
	if reply != expected {
		t.Fatalf("HGETALL\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestHGetAll_Empty(t *testing.T) {
	db := mkDB()
	reply := handleHGetAll(mkCtx(db, "HGETALL", "h"))
	if reply != "*0\r\n" {
		t.Fatalf("HGETALL empty = %q, want *0", reply)
	}
}

// ──── HKEYS ────

func TestHKeys(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "a", "1", "b", "2", "c", "3"))
	reply := handleHKeys(mkCtx(db, "HKEYS", "h"))
	expected := "*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"
	if reply != expected {
		t.Fatalf("HKEYS\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestHKeys_Empty(t *testing.T) {
	db := mkDB()
	reply := handleHKeys(mkCtx(db, "HKEYS", "h"))
	if reply != "*0\r\n" {
		t.Fatalf("HKEYS empty = %q, want *0", reply)
	}
}

// ──── HVALS ────

func TestHVals(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "a", "1", "b", "2"))
	reply := handleHVals(mkCtx(db, "HVALS", "h"))
	expected := "*2\r\n$1\r\n1\r\n$1\r\n2\r\n"
	if reply != expected {
		t.Fatalf("HVALS\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestHVals_Empty(t *testing.T) {
	db := mkDB()
	reply := handleHVals(mkCtx(db, "HVALS", "h"))
	if reply != "*0\r\n" {
		t.Fatalf("HVALS empty = %q, want *0", reply)
	}
}

// ──── HLEN ────

func TestHLen(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "a", "1", "b", "2"))
	reply := handleHLen(mkCtx(db, "HLEN", "h"))
	if reply != ":2\r\n" {
		t.Fatalf("HLEN = %q, want :2", reply)
	}
}

func TestHLen_Empty(t *testing.T) {
	db := mkDB()
	reply := handleHLen(mkCtx(db, "HLEN", "h"))
	if reply != ":0\r\n" {
		t.Fatalf("HLEN empty = %q, want :0", reply)
	}
}

func TestHLen_WrongType(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "h", "str"))
	reply := handleHLen(mkCtx(db, "HLEN", "h"))
	expected := "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
	if reply != expected {
		t.Fatalf("HLEN wrong type\nexpected: %q\n     got: %q", expected, reply)
	}
}

// ──── HDEL ────

func TestHDel(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "a", "1", "b", "2", "c", "3"))
	reply := handleHDel(mkCtx(db, "HDEL", "h", "a", "b"))
	if reply != ":2\r\n" {
		t.Fatalf("HDEL = %q, want :2", reply)
	}
	if hv, _ := getHashItemRead(mkCtx(db, "HGET", "h"), "h"); hv == nil || hv.Len() != 1 {
		t.Fatalf("HDEL: expected 1 field left, got %d", hv.Len())
	}
}

func TestHDel_NonExistentKey(t *testing.T) {
	db := mkDB()
	reply := handleHDel(mkCtx(db, "HDEL", "h", "a"))
	if reply != ":0\r\n" {
		t.Fatalf("HDEL missing key = %q, want :0", reply)
	}
}

func TestHDel_NonExistentField(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "a", "1"))
	reply := handleHDel(mkCtx(db, "HDEL", "h", "nonexist"))
	if reply != ":0\r\n" {
		t.Fatalf("HDEL missing field = %q, want :0", reply)
	}
}

// ──── HEXISTS ────

func TestHExists_Exists(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "f", "v"))
	reply := handleHExists(mkCtx(db, "HEXISTS", "h", "f"))
	if reply != ":1\r\n" {
		t.Fatalf("HEXISTS exists = %q, want :1", reply)
	}
}

func TestHExists_NotExists(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "f", "v"))
	reply := handleHExists(mkCtx(db, "HEXISTS", "h", "nofield"))
	if reply != ":0\r\n" {
		t.Fatalf("HEXISTS missing = %q, want :0", reply)
	}
}

func TestHExists_NonExistentKey(t *testing.T) {
	db := mkDB()
	reply := handleHExists(mkCtx(db, "HEXISTS", "h", "f"))
	if reply != ":0\r\n" {
		t.Fatalf("HEXISTS no key = %q, want :0", reply)
	}
}

// ──── HMGET ────

func TestHMGet(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "a", "1", "b", "2"))
	reply := handleHMGet(mkCtx(db, "HMGET", "h", "a", "b", "c"))
	expected := "*3\r\n$1\r\n1\r\n$1\r\n2\r\n$-1\r\n"
	if reply != expected {
		t.Fatalf("HMGET\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestHMGet_NonExistentKey(t *testing.T) {
	db := mkDB()
	reply := handleHMGet(mkCtx(db, "HMGET", "h", "a"))
	expected := "*1\r\n$-1\r\n"
	if reply != expected {
		t.Fatalf("HMGET no key = %q, want nil", reply)
	}
}

// ──── HSTRLEN ────

func TestHStrLen(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "f", "hello"))
	reply := handleHStrLen(mkCtx(db, "HSTRLEN", "h", "f"))
	if reply != ":5\r\n" {
		t.Fatalf("HSTRLEN = %q, want :5", reply)
	}
}

func TestHStrLen_NonExistentKey(t *testing.T) {
	db := mkDB()
	reply := handleHStrLen(mkCtx(db, "HSTRLEN", "h", "f"))
	if reply != ":0\r\n" {
		t.Fatalf("HSTRLEN no key = %q, want :0", reply)
	}
}

func TestHStrLen_NonExistentField(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "f", "v"))
	reply := handleHStrLen(mkCtx(db, "HSTRLEN", "h", "nofield"))
	if reply != ":0\r\n" {
		t.Fatalf("HSTRLEN missing field = %q, want :0", reply)
	}
}

func TestHStrLen_UTF8(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "f", "世界"))
	reply := handleHStrLen(mkCtx(db, "HSTRLEN", "h", "f"))
	if reply != ":6\r\n" {
		t.Fatalf("HSTRLEN UTF-8 = %q, want :6", reply)
	}
}

// ──── HSCAN ────

func TestHScan_Basic(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "a", "1", "b", "2", "c", "3"))
	reply := handleHScan(mkCtx(db, "HSCAN", "h", "0"))
	// 应该有 a,b,c 三个 fields
	if reply[:6] != "*2\r\n$1" && reply[:6] != "*2\r\n$2" {
		t.Fatalf("HSCAN unexpected start: %q", reply[:20])
	}
}

func TestHScan_Empty(t *testing.T) {
	db := mkDB()
	reply := handleHScan(mkCtx(db, "HSCAN", "h", "0"))
	expected := "*2\r\n$1\r\n0\r\n*0\r\n"
	if reply != expected {
		t.Fatalf("HSCAN empty\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestHScan_WithMatch(t *testing.T) {
	db := mkDB()
	handleHSet(mkCtx(db, "HSET", "h", "abc", "1", "bcd", "2", "cde", "3"))
	reply := handleHScan(mkCtx(db, "HSCAN", "h", "0", "MATCH", "a*"))
	if !stringsContains(reply, "abc") {
		t.Fatalf("HSCAN MATCH a* should contain abc, got: %q", reply)
	}
}

func TestHScan_WrongType(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "h", "str"))
	reply := handleHScan(mkCtx(db, "HSCAN", "h", "0"))
	expected := "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
	if reply != expected {
		t.Fatalf("HSCAN wrong type\nexpected: %q\n     got: %q", expected, reply)
	}
}

// ──── WRONGTYPE 通用测试 ────

// 模拟将 hash key 改为其他类型
func setTypeToHash(db *datastore.GodisDB, key string) {
	db.SetItem(key, datastore.Item{
		Type:       types.TypeHash,
		Value:      types.NewHashValue(),
		IsNeverDie: true,
	})
}

func TestHash_WrongTypeOnStringKey(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "s", "val"))

	commands := []struct {
		name string
		fn   func(*CommandContext) string
		args []string
	}{
		{"HGET", handleHGet, []string{"HGET", "s", "f"}},
		{"HSET", handleHSet, []string{"HSET", "s", "f", "v"}},
		{"HDEL", handleHDel, []string{"HDEL", "s", "f"}},
		{"HLEN", handleHLen, []string{"HLEN", "s"}},
		{"HGETALL", handleHGetAll, []string{"HGETALL", "s"}},
		{"HKEYS", handleHKeys, []string{"HKEYS", "s"}},
		{"HVALS", handleHVals, []string{"HVALS", "s"}},
		{"HEXISTS", handleHExists, []string{"HEXISTS", "s", "f"}},
		{"HMGET", handleHMGet, []string{"HMGET", "s", "f"}},
		{"HMSET", handleHMSet, []string{"HMSET", "s", "f", "v"}},
		{"HSCAN", handleHScan, []string{"HSCAN", "s", "0"}},
		{"HSTRLEN", handleHStrLen, []string{"HSTRLEN", "s", "f"}},
	}

	for _, c := range commands {
		reply := c.fn(mkCtx(db, c.args...))
		expected := "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
		if reply != expected {
			t.Errorf("%s on string key\nexpected: %q\n     got: %q", c.name, expected, reply)
		}
	}
}

// stringsContains 检查字符串是否包含子串
func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
