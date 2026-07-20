package commands

import (
	"testing"

	"godis/datastore"
)

// ---- SADD ----

func TestSAdd(t *testing.T) {
	db := mkDB()
	reply := handleSAdd(mkCtx(db, "SADD", "s", "a", "b", "c"))
	if reply != ":3\r\n" {
		t.Fatalf("SADD = %q, want :3", reply)
	}
}

func TestSAdd_Duplicate(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s", "a"))
	reply := handleSAdd(mkCtx(db, "SADD", "s", "a", "b"))
	if reply != ":1\r\n" {
		t.Fatalf("SADD dup = %q, want :1", reply)
	}
}

// ---- SCARD ----

func TestSCard(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s", "a", "b"))
	reply := handleSCard(mkCtx(db, "SCARD", "s"))
	if reply != ":2\r\n" {
		t.Fatalf("SCARD = %q, want :2", reply)
	}
}

func TestSCard_Empty(t *testing.T) {
	db := mkDB()
	reply := handleSCard(mkCtx(db, "SCARD", "s"))
	if reply != ":0\r\n" {
		t.Fatalf("SCARD empty = %q, want :0", reply)
	}
}

// ---- SISMEMBER ----

func TestSIsMember(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s", "a"))
	reply := handleSIsMember(mkCtx(db, "SISMEMBER", "s", "a"))
	if reply != ":1\r\n" {
		t.Fatalf("SISMEMBER = %q, want :1", reply)
	}
}

func TestSIsMember_NotExists(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s", "a"))
	reply := handleSIsMember(mkCtx(db, "SISMEMBER", "s", "b"))
	if reply != ":0\r\n" {
		t.Fatalf("SISMEMBER = %q, want :0", reply)
	}
}

// ---- SMEMBERS ----

func TestSMembers(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s", "b", "a"))
	reply := handleSMembers(mkCtx(db, "SMEMBERS", "s"))
	expected := "*2\r\n$1\r\na\r\n$1\r\nb\r\n"
	if reply != expected {
		t.Fatalf("SMEMBERS\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestSMembers_Empty(t *testing.T) {
	db := mkDB()
	reply := handleSMembers(mkCtx(db, "SMEMBERS", "s"))
	if reply != "*0\r\n" {
		t.Fatalf("SMEMBERS empty = %q, want *0", reply)
	}
}

// ---- SMISMEMBER ----

func TestSMIsMember(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s", "a", "c"))
	reply := handleSMIsMember(mkCtx(db, "SMISMEMBER", "s", "a", "b", "c"))
	expected := "*3\r\n:1\r\n:0\r\n:1\r\n"
	if reply != expected {
		t.Fatalf("SMISMEMBER\nexpected: %q\n     got: %q", expected, reply)
	}
}

// ---- SREM ----

func TestSRem(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s", "a", "b", "c"))
	reply := handleSRem(mkCtx(db, "SREM", "s", "a", "b", "x"))
	if reply != ":2\r\n" {
		t.Fatalf("SREM = %q, want :2", reply)
	}
}

func TestSRem_Empty(t *testing.T) {
	db := mkDB()
	reply := handleSRem(mkCtx(db, "SREM", "s", "a"))
	if reply != ":0\r\n" {
		t.Fatalf("SREM empty = %q, want :0", reply)
	}
}

// ---- SPOP ----

func TestSPop(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s", "a"))
	reply := handleSPop(mkCtx(db, "SPOP", "s"))
	if reply != "$1\r\na\r\n" {
		t.Fatalf("SPOP = %q, want a", reply)
	}
	if scardReply := handleSCard(mkCtx(db, "SCARD", "s")); scardReply != ":0\r\n" {
		t.Fatalf("SPOP: set should be empty, got %q", scardReply)
	}
}

func TestSPop_Empty(t *testing.T) {
	db := mkDB()
	reply := handleSPop(mkCtx(db, "SPOP", "s"))
	if reply != "$-1\r\n" {
		t.Fatalf("SPOP empty = %q, want nil", reply)
	}
}

// ---- SRANDMEMBER ----

func TestSRandMember(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s", "a", "b"))
	reply := handleSRandMember(mkCtx(db, "SRANDMEMBER", "s"))
	if reply != "$1\r\na\r\n" && reply != "$1\r\nb\r\n" {
		t.Fatalf("SRANDMEMBER = %q, want a or b", reply)
	}
}

func TestSRandMember_Count(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s", "a", "b", "c"))
	reply := handleSRandMember(mkCtx(db, "SRANDMEMBER", "s", "2"))
	// 应该返回数组，有2个元素
	if reply[:6] != "*2\r\n$1" && reply[:6] != "*2\r\n$2" {
		t.Fatalf("SRANDMEMBER 2 unexpected: %q", reply[:20])
	}
}

// ---- SMOVE ----

func TestSMove(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s1", "a"))
	reply := handleSMove(mkCtx(db, "SMOVE", "s1", "s2", "a"))
	if reply != ":1\r\n" {
		t.Fatalf("SMOVE = %q, want :1", reply)
	}
	if !checkSIsMember(t, db, "s2", "a") {
		t.Fatal("SMOVE: s2 should have a")
	}
	if checkSIsMember(t, db, "s1", "a") {
		t.Fatal("SMOVE: s1 should NOT have a")
	}
}

func TestSMove_NotExists(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s1", "a"))
	reply := handleSMove(mkCtx(db, "SMOVE", "s1", "s2", "x"))
	if reply != ":0\r\n" {
		t.Fatalf("SMOVE missing = %q, want :0", reply)
	}
}

func checkSIsMember(t *testing.T, db *datastore.GodisDB, key, member string) bool {
	t.Helper()
	reply := handleSIsMember(mkCtx(db, "SISMEMBER", key, member))
	return reply == ":1\r\n"
}

// ---- SUNION / SUNIONSTORE ----

func TestSUnion(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s1", "a", "b"))
	handleSAdd(mkCtx(db, "SADD", "s2", "b", "c"))
	reply := handleSUnion(mkCtx(db, "SUNION", "s1", "s2"))
	expected := "*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"
	if reply != expected {
		t.Fatalf("SUNION\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestSUnionStore(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s1", "a"))
	handleSAdd(mkCtx(db, "SADD", "s2", "b"))
	reply := handleSUnionStore(mkCtx(db, "SUNIONSTORE", "dst", "s1", "s2"))
	if reply != ":2\r\n" {
		t.Fatalf("SUNIONSTORE = %q, want :2", reply)
	}
}

// ---- SINTER / SINTERSTORE ----

func TestSInter(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s1", "a", "b", "c"))
	handleSAdd(mkCtx(db, "SADD", "s2", "b", "c", "d"))
	reply := handleSInter(mkCtx(db, "SINTER", "s1", "s2"))
	expected := "*2\r\n$1\r\nb\r\n$1\r\nc\r\n"
	if reply != expected {
		t.Fatalf("SINTER\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestSInter_OneMissing(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s1", "a"))
	reply := handleSInter(mkCtx(db, "SINTER", "s1", "s2"))
	if reply != "*0\r\n" {
		t.Fatalf("SINTER missing = %q, want *0", reply)
	}
}

func TestSInterStore(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s1", "a", "b"))
	handleSAdd(mkCtx(db, "SADD", "s2", "b", "c"))
	reply := handleSInterStore(mkCtx(db, "SINTERSTORE", "dst", "s1", "s2"))
	if reply != ":1\r\n" {
		t.Fatalf("SINTERSTORE = %q, want :1", reply)
	}
}

// ---- SINTERCARD ----

func TestSInterCard(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s1", "a", "b", "c"))
	handleSAdd(mkCtx(db, "SADD", "s2", "b", "c", "d"))
	reply := handleSInterCard(mkCtx(db, "SINTERCARD", "2", "s1", "s2"))
	if reply != ":2\r\n" {
		t.Fatalf("SINTERCARD = %q, want :2", reply)
	}
}

// ---- SDIFF / SDIFFSTORE ----

func TestSDiff(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s1", "a", "b", "c"))
	handleSAdd(mkCtx(db, "SADD", "s2", "b", "d"))
	reply := handleSDiff(mkCtx(db, "SDIFF", "s1", "s2"))
	expected := "*2\r\n$1\r\na\r\n$1\r\nc\r\n"
	if reply != expected {
		t.Fatalf("SDIFF\nexpected: %q\n     got: %q", expected, reply)
	}
}

func TestSDiffStore(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s1", "a", "b"))
	handleSAdd(mkCtx(db, "SADD", "s2", "b"))
	reply := handleSDiffStore(mkCtx(db, "SDIFFSTORE", "dst", "s1", "s2"))
	if reply != ":1\r\n" {
		t.Fatalf("SDIFFSTORE = %q, want :1", reply)
	}
}

// ---- SSCAN ----

func TestSScan_Basic(t *testing.T) {
	db := mkDB()
	handleSAdd(mkCtx(db, "SADD", "s", "a", "b", "c"))
	reply := handleSScan(mkCtx(db, "SSCAN", "s", "0"))
	if reply[:6] != "*2\r\n$1" && reply[:6] != "*2\r\n$2" {
		t.Fatalf("SSCAN unexpected: %q", reply[:20])
	}
}

func TestSScan_Empty(t *testing.T) {
	db := mkDB()
	reply := handleSScan(mkCtx(db, "SSCAN", "s", "0"))
	expected := "*2\r\n$1\r\n0\r\n*0\r\n"
	if reply != expected {
		t.Fatalf("SSCAN empty\nexpected: %q\n     got: %q", expected, reply)
	}
}

// ---- WRONGTYPE ----

func TestSet_WrongType(t *testing.T) {
	db := mkDB()
	handleSet(mkCtx(db, "SET", "s", "str"))

	tests := []struct {
		name string
		args []string
		fn   func(*CommandContext) string
	}{
		{"SADD", []string{"SADD", "s", "a"}, handleSAdd},
		{"SCARD", []string{"SCARD", "s"}, handleSCard},
		{"SISMEMBER", []string{"SISMEMBER", "s", "a"}, handleSIsMember},
		{"SMEMBERS", []string{"SMEMBERS", "s"}, handleSMembers},
		{"SMISMEMBER", []string{"SMISMEMBER", "s", "a"}, handleSMIsMember},
		{"SREM", []string{"SREM", "s", "a"}, handleSRem},
		{"SPOP", []string{"SPOP", "s"}, handleSPop},
		{"SRANDMEMBER", []string{"SRANDMEMBER", "s"}, handleSRandMember},
	}

	for _, tc := range tests {
		reply := tc.fn(mkCtx(db, tc.args...))
		expected := "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
		if reply != expected {
			t.Errorf("%s: expected %q, got %q", tc.name, expected, reply)
		}
	}
}
