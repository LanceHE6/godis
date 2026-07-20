package commands

import (
	"testing"

	"godis/types"
)

// ---- LPUSH ----

func TestLPush(t *testing.T) {
	db := mkDB()
	reply := handleLPush(mkCtx(db, "LPUSH", "k", "a", "b", "c"))
	if reply != ":3\r\n" {
		t.Fatalf("LPUSH = %q, want :3", reply)
	}
	// 验证顺序：c, b, a
	lv, _ := getListItemRead(mkCtx(db, "LINDEX", "k"), "k")
	vals := lv.Data()
	if len(vals) != 3 || vals[0] != "c" || vals[1] != "b" || vals[2] != "a" {
		t.Fatalf("LPUSH order: %v, want [c b a]", vals)
	}
}

func TestLPush_WrongType(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "k", "s"))
	reply := handleLPush(mkCtx(db, "LPUSH", "k", "v"))
	expected := "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
	if reply != expected {
		t.Fatalf("LPUSH wrong type\nexpected: %q\n     got: %q", expected, reply)
	}
}

// ---- LPUSHX ----

func TestLPushX_Existing(t *testing.T) {
	db := mkDB()
	handleLPush(mkCtx(db, "LPUSH", "k", "a"))
	reply := handleLPushX(mkCtx(db, "LPUSHX", "k", "b"))
	if reply != ":2\r\n" {
		t.Fatalf("LPUSHX = %q, want :2", reply)
	}
}

func TestLPushX_NonExistent(t *testing.T) {
	db := mkDB()
	reply := handleLPushX(mkCtx(db, "LPUSHX", "k", "a"))
	if reply != ":0\r\n" {
		t.Fatalf("LPUSHX missing key = %q, want :0", reply)
	}
}

// ---- RPUSH ----

func TestRPush(t *testing.T) {
	db := mkDB()
	reply := handleRPush(mkCtx(db, "RPUSH", "k", "a", "b", "c"))
	if reply != ":3\r\n" {
		t.Fatalf("RPUSH = %q, want :3", reply)
	}
	lv, _ := getListItemRead(mkCtx(db, "LINDEX", "k"), "k")
	vals := lv.Data()
	if len(vals) != 3 || vals[0] != "a" || vals[2] != "c" {
		t.Fatalf("RPUSH order: %v, want [a b c]", vals)
	}
}

// ---- RPUSHX ----

func TestRPushX_NonExistent(t *testing.T) {
	db := mkDB()
	reply := handleRPushX(mkCtx(db, "RPUSHX", "k", "a"))
	if reply != ":0\r\n" {
		t.Fatalf("RPUSHX missing key = %q, want :0", reply)
	}
}

// ---- LPOP ----

func TestLPop(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "b", "c"))
	reply := handleLPop(mkCtx(db, "LPOP", "k"))
	if reply != "$1\r\na\r\n" {
		t.Fatalf("LPOP = %q, want a", reply)
	}
}

func TestLPop_Empty(t *testing.T) {
	db := mkDB()
	reply := handleLPop(mkCtx(db, "LPOP", "k"))
	if reply != "$-1\r\n" {
		t.Fatalf("LPOP empty = %q, want nil", reply)
	}
}

// ---- RPOP ----

func TestRPop(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "b", "c"))
	reply := handleRPop(mkCtx(db, "RPOP", "k"))
	if reply != "$1\r\nc\r\n" {
		t.Fatalf("RPOP = %q, want c", reply)
	}
}

// ---- LLEN ----

func TestLLen(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "b"))
	reply := handleLLen(mkCtx(db, "LLEN", "k"))
	if reply != ":2\r\n" {
		t.Fatalf("LLEN = %q, want :2", reply)
	}
}

func TestLLen_Empty(t *testing.T) {
	db := mkDB()
	reply := handleLLen(mkCtx(db, "LLEN", "k"))
	if reply != ":0\r\n" {
		t.Fatalf("LLEN empty = %q, want :0", reply)
	}
}

// ---- LINDEX ----

func TestLIndex(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "b", "c"))
	reply := handleLIndex(mkCtx(db, "LINDEX", "k", "1"))
	if reply != "$1\r\nb\r\n" {
		t.Fatalf("LINDEX = %q, want b", reply)
	}
}

func TestLIndex_Negative(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "b", "c"))
	reply := handleLIndex(mkCtx(db, "LINDEX", "k", "-1"))
	if reply != "$1\r\nc\r\n" {
		t.Fatalf("LINDEX -1 = %q, want c", reply)
	}
}

func TestLIndex_NonExistent(t *testing.T) {
	db := mkDB()
	reply := handleLIndex(mkCtx(db, "LINDEX", "k", "0"))
	if reply != "$-1\r\n" {
		t.Fatalf("LINDEX empty = %q, want nil", reply)
	}
}

// ---- LINSERT ----

func TestLInsert_Before(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "c"))
	reply := handleLInsert(mkCtx(db, "LINSERT", "k", "BEFORE", "c", "b"))
	if reply != ":3\r\n" {
		t.Fatalf("LINSERT BEFORE = %q, want :3", reply)
	}
	lv, _ := getListItemRead(mkCtx(db, "LINDEX", "k"), "k")
	vals := lv.Data()
	if len(vals) != 3 || vals[1] != "b" {
		t.Fatalf("LINSERT result: %v, want [a b c]", vals)
	}
}

func TestLInsert_After(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "c"))
	reply := handleLInsert(mkCtx(db, "LINSERT", "k", "AFTER", "a", "b"))
	if reply != ":3\r\n" {
		t.Fatalf("LINSERT AFTER = %q, want :3", reply)
	}
}

func TestLInsert_NotFound(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a"))
	reply := handleLInsert(mkCtx(db, "LINSERT", "k", "BEFORE", "x", "y"))
	if reply != ":-1\r\n" {
		t.Fatalf("LINSERT not found = %q, want :-1", reply)
	}
}

// ---- LRANGE ----

func TestLRange(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "b", "c", "d"))
	reply := handleLRange(mkCtx(db, "LRANGE", "k", "1", "2"))
	expected := "*2\r\n$1\r\nb\r\n$1\r\nc\r\n"
	if reply != expected {
		t.Fatalf("LRANGE 1 2\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestLRange_Negative(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "b", "c"))
	reply := handleLRange(mkCtx(db, "LRANGE", "k", "-2", "-1"))
	expected := "*2\r\n$1\r\nb\r\n$1\r\nc\r\n"
	if reply != expected {
		t.Fatalf("LRANGE -2 -1\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestLRange_Empty(t *testing.T) {
	db := mkDB()
	reply := handleLRange(mkCtx(db, "LRANGE", "k", "0", "-1"))
	if reply != "*0\r\n" {
		t.Fatalf("LRANGE empty = %q, want *0", reply)
	}
}

// ---- LREM ----

func TestLRem_PositiveCount(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "b", "a", "a", "c"))
	reply := handleLRem(mkCtx(db, "LREM", "k", "2", "a"))
	if reply != ":2\r\n" {
		t.Fatalf("LREM 2 = %q, want :2", reply)
	}
}

func TestLRem_All(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "a", "b"))
	reply := handleLRem(mkCtx(db, "LREM", "k", "0", "a"))
	if reply != ":2\r\n" {
		t.Fatalf("LREM 0 = %q, want :2", reply)
	}
}

// ---- LSET ----

func TestLSet(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "b", "c"))
	reply := handleLSet(mkCtx(db, "LSET", "k", "1", "X"))
	if reply != "+OK\r\n" {
		t.Fatalf("LSET = %q, want +OK", reply)
	}
	val, _ := db.GetItem("k")
	lv := val.Value.(*types.ListValue)
	v, _ := lv.Index(1)
	if v != "X" {
		t.Fatalf("LSET result = %q, want X", v)
	}
}

func TestLSet_OutOfRange(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a"))
	reply := handleLSet(mkCtx(db, "LSET", "k", "5", "X"))
	expected := "-ERR index out of range\r\n"
	if reply != expected {
		t.Fatalf("LSET OOB\nexpected: %q\n     got: %q", expected, reply)
	}
}

// ---- LPOS ----

func TestLPos(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "b", "a", "c"))
	reply := handleLPos(mkCtx(db, "LPOS", "k", "a"))
	if reply != ":0\r\n" {
		t.Fatalf("LPOS = %q, want :0", reply)
	}
}

func TestLPos_NotFound(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a"))
	reply := handleLPos(mkCtx(db, "LPOS", "k", "x"))
	if reply != "$-1\r\n" {
		t.Fatalf("LPOS not found = %q, want nil", reply)
	}
}

// ---- LMOVE ----

func TestLMove(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "src", "a", "b", "c"))
	reply := handleLMove(mkCtx(db, "LMOVE", "src", "dst", "LEFT", "RIGHT"))
	if reply != "$1\r\na\r\n" {
		t.Fatalf("LMOVE = %q, want a", reply)
	}
	// 验证 src 少了 a
	lv, _ := getListItemRead(mkCtx(db, "LLEN", "src"), "src")
	if lv.Len() != 2 {
		t.Fatalf("LMOVE: src len = %d, want 2", lv.Len())
	}
	// 验证 dst 有 a
	dst, _ := getListItemRead(mkCtx(db, "LLEN", "dst"), "dst")
	vals := dst.Data()
	if len(vals) != 1 || vals[0] != "a" {
		t.Fatalf("LMOVE: dst = %v, want [a]", vals)
	}
}

func TestLMove_EmptySource(t *testing.T) {
	db := mkDB()
	reply := handleLMove(mkCtx(db, "LMOVE", "src", "dst", "LEFT", "RIGHT"))
	if reply != "$-1\r\n" {
		t.Fatalf("LMOVE empty = %q, want nil", reply)
	}
}

// ---- BLPOP / BRPOP ----

func TestBLPop(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "b"))
	reply := handleBLPop(mkCtx(db, "BLPOP", "k", "0"))
	expected := "*2\r\n$1\r\nk\r\n$1\r\na\r\n"
	if reply != expected {
		t.Fatalf("BLPOP\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestBLPop_Empty(t *testing.T) {
	db := mkDB()
	reply := handleBLPop(mkCtx(db, "BLPOP", "k", "0"))
	if reply != "$-1\r\n" {
		t.Fatalf("BLPOP empty = %q, want nil", reply)
	}
}

func TestBRPop(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "k", "a", "b"))
	reply := handleBRPop(mkCtx(db, "BRPOP", "k", "0"))
	expected := "*2\r\n$1\r\nk\r\n$1\r\nb\r\n"
	if reply != expected {
		t.Fatalf("BRPOP\nexpected: %q\n     got: %q", expected, reply)
	}
}

// ---- BLMOVE ----

func TestBLMove(t *testing.T) {
	db := mkDB()
	handleRPush(mkCtx(db, "RPUSH", "src", "a"))
	reply := handleBLMove(mkCtx(db, "BLMOVE", "src", "dst", "LEFT", "RIGHT", "0"))
	if reply != "$1\r\na\r\n" {
		t.Fatalf("BLMOVE = %q, want a", reply)
	}
}

// ---- 不存在的 key 和 WRONGTYPE 边界 ----

func TestList_WrongType(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "s", "val"))

	tests := []struct {
		name string
		args []string
		fn   func(*CommandContext) string
	}{
		{"LINDEX", []string{"LINDEX", "s", "0"}, handleLIndex},
		{"LLEN", []string{"LLEN", "s"}, handleLLen},
		{"LPOP", []string{"LPOP", "s"}, handleLPop},
		{"RPOP", []string{"RPOP", "s"}, handleRPop},
		{"LRANGE", []string{"LRANGE", "s", "0", "-1"}, handleLRange},
		{"LINSERT", []string{"LINSERT", "s", "BEFORE", "x", "y"}, handleLInsert},
		{"LREM", []string{"LREM", "s", "0", "a"}, handleLRem},
		{"LSET", []string{"LSET", "s", "0", "v"}, handleLSet},
		{"LPOS", []string{"LPOS", "s", "a"}, handleLPos},
		{"LPUSH", []string{"LPUSH", "s", "v"}, handleLPush},
		{"RPUSH", []string{"RPUSH", "s", "v"}, handleRPush},
	}

	for _, tc := range tests {
		reply := tc.fn(mkCtx(db, tc.args...))
		expected := "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
		if reply != expected {
			t.Errorf("%s: expected %q, got %q", tc.name, expected, reply)
		}
	}
}
