package commands

import (
	"testing"

	"godis/datastore"
	"godis/types"
)

// mkCtx 快速创建命令上下文的辅助函数
func mkCtx(db *datastore.GodisDB, args ...string) *CommandContext {
	id := 0
	return &CommandContext{
		Args:        args,
		DB:          db,
		AllDBs:      []*datastore.GodisDB{db},
		CurrentDBID: &id,
	}
}

// mkDB 创建新数据库的辅助函数
func mkDB() *datastore.GodisDB {
	return datastore.NewGodisDB()
}

// ---- SET ----

func TestSet(t *testing.T) {
	db := mkDB()
	reply := handleSet(mkCtx(db, "SET", "k", "v"))
	if reply != "+OK\r\n" {
		t.Fatalf("SET = %q, want +OK", reply)
	}
	// 验证值已设置
	val, exists := db.Get("k")
	if !exists || val != "v" {
		t.Fatalf("GET k = %q, want v (exists=%v)", val, exists)
	}
}

func TestSet_WithEX(t *testing.T) {
	db := mkDB()
	reply := handleSet(mkCtx(db, "SET", "k", "v", "EX", "5"))
	if reply != "+OK\r\n" {
		t.Fatalf("SET EX = %q, want +OK", reply)
	}
	// 检查 TTL
	item, exists := db.GetItem("k")
	if !exists {
		t.Fatal("key should exist")
	}
	if item.IsNeverDie {
		t.Fatal("key should have TTL")
	}
}

func TestSet_Overwrite(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "old"))
	handleSet(mkCtx(db, "SET", "k", "new"))
	val, exists := db.Get("k")
	if !exists || val != "new" {
		t.Fatalf("GET k = %q, want new", val)
	}
}

// ---- GET ----

func TestGet(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "hello"))
	reply := handleGet(mkCtx(db, "GET", "k"))
	if reply != "$5\r\nhello\r\n" {
		t.Fatalf("GET = %q, want $5\\r\\nhello\\r\\n", reply)
	}
}

func TestGet_NonExistent(t *testing.T) {
	db := mkDB()
	reply := handleGet(mkCtx(db, "GET", "nokey"))
	if reply != "$-1\r\n" {
		t.Fatalf("GET nokey = %q, want $-1", reply)
	}
}

// ---- MSET ----

func TestMSetAndMGet(t *testing.T) {
	db := mkDB()
	reply := handleMSet(mkCtx(db, "MSET", "k1", "v1", "k2", "v2", "k3", "v3"))
	if reply != "+OK\r\n" {
		t.Fatalf("MSET = %q, want +OK", reply)
	}

	reply = handleMGet(mkCtx(db, "MGET", "k1", "k2", "k3", "nonexistent"))
	expected := "*4\r\n$2\r\nv1\r\n$2\r\nv2\r\n$2\r\nv3\r\n$-1\r\n"
	if reply != expected {
		t.Fatalf("MGET\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestMSet_OddArgs(t *testing.T) {
	db := mkDB()
	reply := handleMSet(mkCtx(db, "MSET", "k1", "v1", "k2"))
	expected := "-ERR wrong number of arguments for MSET\r\n"
	if reply != expected {
		t.Fatalf("MSET odd args\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestMGet_SingleKey(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "v"))
	reply := handleMGet(mkCtx(db, "MGET", "k"))
	if reply != "*1\r\n$1\r\nv\r\n" {
		t.Fatalf("MGET single = %q, want *1\\r\\n$1\\r\\nv\\r\\n", reply)
	}
}

func TestMGet_Empty(t *testing.T) {
	db := mkDB()
	// 直接调用 handler（绕过路由器参数校验），空参数返回空数组
	reply := handleMGet(mkCtx(db, "MGET"))
	if reply != "*0\r\n" {
		t.Fatalf("MGET empty = %q, want *0", reply)
	}
	// 通过路由器执行应返回参数错误（已在 TestExecute_ArityMin 中覆盖）
}

func TestMGet_AllMissing(t *testing.T) {
	db := mkDB()
	reply := handleMGet(mkCtx(db, "MGET", "a", "b"))
	expected := "*2\r\n$-1\r\n$-1\r\n"
	if reply != expected {
		t.Fatalf("MGET all missing\nexpected: %q\n     got: %q", expected, reply)
	}
}

// ---- APPEND ----

func TestAppend_Create(t *testing.T) {
	db := mkDB()
	reply := handleAppend(mkCtx(db, "APPEND", "k", "hello"))
	if reply != ":5\r\n" {
		t.Fatalf("APPEND create = %q, want :5", reply)
	}
	val, exists := db.Get("k")
	if !exists || val != "hello" {
		t.Fatalf("GET k = %q, want hello", val)
	}
}

func TestAppend_Existing(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "hello"))
	reply := handleAppend(mkCtx(db, "APPEND", "k", " world"))
	if reply != ":11\r\n" {
		t.Fatalf("APPEND existing = %q, want :11", reply)
	}
	val, _ := db.Get("k")
	if val != "hello world" {
		t.Fatalf("GET k = %q, want 'hello world'", val)
	}
}

// ---- BITCOUNT ----

func TestBitCount_NonExistent(t *testing.T) {
	db := mkDB()
	reply := handleBitCount(mkCtx(db, "BITCOUNT", "nokey"))
	if reply != ":0\r\n" {
		t.Fatalf("BITCOUNT nokey = %q, want :0", reply)
	}
}

func TestBitCount_All(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "a")) // 'a' = 0x61 = 0b01100001 = 3 bits
	reply := handleBitCount(mkCtx(db, "BITCOUNT", "k"))
	if reply != ":3\r\n" {
		t.Fatalf("BITCOUNT k = %q, want :3", reply)
	}
}

func TestBitCount_Range(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "ab")) // 'a'=3 bits, 'b'=3 bits
	reply := handleBitCount(mkCtx(db, "BITCOUNT", "k", "0", "0"))
	if reply != ":3\r\n" {
		t.Fatalf("BITCOUNT k 0 0 = %q, want :3", reply)
	}
}

func TestBitCount_InvalidRange(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "a"))
	reply := handleBitCount(mkCtx(db, "BITCOUNT", "k", "5", "3"))
	if reply != ":0\r\n" {
		t.Fatalf("BITCOUNT invalid range = %q, want :0", reply)
	}
}

// ---- INCR / INCRBY ----

func TestIncr_Create(t *testing.T) {
	db := mkDB()
	reply := handleIncr(mkCtx(db, "INCR", "k"))
	if reply != ":1\r\n" {
		t.Fatalf("INCR create = %q, want :1", reply)
	}
}

func TestIncr_Existing(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "5"))
	reply := handleIncr(mkCtx(db, "INCR", "k"))
	if reply != ":6\r\n" {
		t.Fatalf("INCR = %q, want :6", reply)
	}
}

func TestIncrBy(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "10"))
	reply := handleIncrBy(mkCtx(db, "INCRBY", "k", "7"))
	if reply != ":17\r\n" {
		t.Fatalf("INCRBY = %q, want :17", reply)
	}
}

// ---- DECR / DECRBY ----

func TestDecr_Create(t *testing.T) {
	db := mkDB()
	reply := handleDecr(mkCtx(db, "DECR", "k"))
	if reply != ":-1\r\n" {
		t.Fatalf("DECR create = %q, want :-1", reply)
	}
}

func TestDecr_Existing(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "5"))
	reply := handleDecr(mkCtx(db, "DECR", "k"))
	if reply != ":4\r\n" {
		t.Fatalf("DECR = %q, want :4", reply)
	}
}

func TestDecrBy(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "10"))
	reply := handleDecrBy(mkCtx(db, "DECRBY", "k", "3"))
	if reply != ":7\r\n" {
		t.Fatalf("DECRBY = %q, want :7", reply)
	}
}

// ---- GETRANGE ----

func TestGetRange(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "Hello"))
	reply := handleGetRange(mkCtx(db, "GETRANGE", "k", "0", "4"))
	if reply != "$5\r\nHello\r\n" {
		t.Fatalf("GETRANGE 0 4 = %q, want Hello", reply)
	}
}

func TestGetRange_Partial(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "Hello"))
	reply := handleGetRange(mkCtx(db, "GETRANGE", "k", "0", "2"))
	if reply != "$3\r\nHel\r\n" {
		t.Fatalf("GETRANGE 0 2 = %q, want Hel", reply)
	}
}

func TestGetRange_Negative(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "Hello"))
	reply := handleGetRange(mkCtx(db, "GETRANGE", "k", "-3", "-1"))
	if reply != "$3\r\nllo\r\n" {
		t.Fatalf("GETRANGE -3 -1 = %q, want llo", reply)
	}
}

func TestGetRange_NonExistent(t *testing.T) {
	db := mkDB()
	reply := handleGetRange(mkCtx(db, "GETRANGE", "nokey", "0", "1"))
	if reply != "$0\r\n\r\n" {
		t.Fatalf("GETRANGE nokey = %q, want empty bulk string", reply)
	}
}

// ---- GETSET ----

func TestGetSet_NonExistent(t *testing.T) {
	db := mkDB()
	reply := handleGetSet(mkCtx(db, "GETSET", "k", "new"))
	if reply != "$-1\r\n" {
		t.Fatalf("GETSET new key = %q, want nil", reply)
	}
	// 验证新值已设置
	val, exists := db.Get("k")
	if !exists || val != "new" {
		t.Fatalf("GET k = %q, want new", val)
	}
}

func TestGetSet_Existing(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "old"))
	reply := handleGetSet(mkCtx(db, "GETSET", "k", "new"))
	if reply != "$3\r\nold\r\n" {
		t.Fatalf("GETSET existing = %q, want old", reply)
	}
	val, _ := db.Get("k")
	if val != "new" {
		t.Fatalf("GET k = %q, want new", val)
	}
}

// ---- STRLEN ----

func TestStrLen_NonExistent(t *testing.T) {
	db := mkDB()
	reply := handleStrLen(mkCtx(db, "STRLEN", "nokey"))
	if reply != ":0\r\n" {
		t.Fatalf("STRLEN nokey = %q, want :0", reply)
	}
}

func TestStrLen_Normal(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "hello"))
	reply := handleStrLen(mkCtx(db, "STRLEN", "k"))
	if reply != ":5\r\n" {
		t.Fatalf("STRLEN = %q, want :5", reply)
	}
}

func TestStrLen_Empty(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", ""))
	reply := handleStrLen(mkCtx(db, "STRLEN", "k"))
	if reply != ":0\r\n" {
		t.Fatalf("STRLEN empty = %q, want :0", reply)
	}
}

func TestStrLen_UTF8(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "世界"))
	reply := handleStrLen(mkCtx(db, "STRLEN", "k"))
	if reply != ":6\r\n" {
		t.Fatalf("STRLEN UTF-8 = %q, want :6 (6 UTF-8 bytes)", reply)
	}
}

func TestStrLen_WrongType(t *testing.T) {
	db := mkDB()
	// 先用 SET 设置后，通过 datastore 手动改成 hash 类型模拟 WRONGTYPE
	handleSet(mkCtx(db, "SET", "k", "val"))
	item, _ := db.GetItem("k")
	item.Type = types.TypeHash
	db.SetItem("k", *item)

	reply := handleStrLen(mkCtx(db, "STRLEN", "k"))
	expected := "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
	if reply != expected {
		t.Fatalf("STRLEN wrong type\nexpected: %q\n     got: %q", expected, reply)
	}
}
