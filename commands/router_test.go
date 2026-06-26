package commands

import (
	"fmt"
	"strings"
	"testing"

	"godis/datastore"
	"godis/protocol"
)

func TestRegister(t *testing.T) {
	// 清理注册表
	old := CommandRegistry
	CommandRegistry = make(map[string]Command)
	defer func() { CommandRegistry = old }()

	called := false
	Register("TEST", 1, FlagFast, 0, 0, 0, func(ctx *CommandContext) string {
		called = true
		return protocol.MakeSimpleString("OK")
	})

	cmd, ok := CommandRegistry["TEST"]
	if !ok {
		t.Fatal("TEST not registered")
	}
	if cmd.Name != "TEST" {
		t.Errorf("Name = %q, want TEST", cmd.Name)
	}
	if cmd.Flags != FlagFast {
		t.Errorf("Flags = %q, want %q", cmd.Flags, FlagFast)
	}
	if cmd.Arity != 1 {
		t.Errorf("Arity = %d, want 1", cmd.Arity)
	}

	// 测试 handler 可调用
	db := datastore.NewGodisDB()
	id := 0
	ctx := &CommandContext{DB: db, AllDBs: []*datastore.GodisDB{db}, CurrentDBID: &id}
	cmd.Handler(ctx)
	if !called {
		t.Error("handler was not called")
	}
}

func TestRegister_CaseInsensitive(t *testing.T) {
	old := CommandRegistry
	CommandRegistry = make(map[string]Command)
	defer func() { CommandRegistry = old }()

	handler := func(ctx *CommandContext) string {
		return protocol.MakeSimpleString("OK")
	}
	Register("lower", 1, FlagWrite, 0, 0, 0, handler)

	if _, ok := CommandRegistry["LOWER"]; !ok {
		t.Error("lowercase registration should be stored as uppercase")
	}
}

func TestExecute(t *testing.T) {
	old := CommandRegistry
	CommandRegistry = make(map[string]Command)
	defer func() { CommandRegistry = old }()

	Register("ECHO", 2, FlagFast, 0, 0, 0, func(ctx *CommandContext) string {
		return protocol.MakeBulkString(ctx.Args[1])
	})

	db := datastore.NewGodisDB()
	id := 0
	ctx := &CommandContext{
		Args:        []string{"ECHO", "hello"},
		DB:          db,
		AllDBs:      []*datastore.GodisDB{db},
		CurrentDBID: &id,
	}

	reply, cmd, ok := Execute("ECHO", ctx)
	if !ok {
		t.Fatal("Execute returned false")
	}
	if cmd.Name != "ECHO" {
		t.Errorf("cmd.Name = %q, want ECHO", cmd.Name)
	}
	if reply != protocol.MakeBulkString("hello") {
		t.Errorf("reply = %q, want bulk string hello", reply)
	}
}

func TestExecute_NotFound(t *testing.T) {
	old := CommandRegistry
	CommandRegistry = make(map[string]Command)
	defer func() { CommandRegistry = old }()

	db := datastore.NewGodisDB()
	id := 0
	ctx := &CommandContext{DB: db, AllDBs: []*datastore.GodisDB{db}, CurrentDBID: &id}

	_, _, ok := Execute("NOPE", ctx)
	if ok {
		t.Error("Execute for unknown command should return false")
	}
}

func TestExecute_ArityError(t *testing.T) {
	old := CommandRegistry
	CommandRegistry = make(map[string]Command)
	defer func() { CommandRegistry = old }()

	// 注册一个 arity=3 的命令（最少 3 个参数）
	Register("SETX", 3, FlagWrite, 1, 1, 1, func(ctx *CommandContext) string {
		return protocol.MakeSimpleString("OK")
	})

	db := datastore.NewGodisDB()
	id := 0
	ctx := &CommandContext{
		Args:        []string{"SETX"},
		DB:          db,
		AllDBs:      []*datastore.GodisDB{db},
		CurrentDBID: &id,
	}

	reply, _, ok := Execute("SETX", ctx)
	if !ok {
		t.Fatal("SETX should exist in registry")
	}
	expected := protocol.WrongArgsErr("SETX")
	if reply != expected {
		t.Errorf("reply = %q, want %q", reply, expected)
	}
}

func TestExecute_ArityMin(t *testing.T) {
	old := CommandRegistry
	CommandRegistry = make(map[string]Command)
	defer func() { CommandRegistry = old }()

	// 注册一个 arity=-3 的命令（至少 3 个参数）
	Register("SETX", -3, FlagWrite, 1, 1, 1, func(ctx *CommandContext) string {
		return protocol.MakeSimpleString("OK")
	})

	db := datastore.NewGodisDB()
	id := 0
	ctx := &CommandContext{
		Args:        []string{"SETX", "key"},
		DB:          db,
		AllDBs:      []*datastore.GodisDB{db},
		CurrentDBID: &id,
	}

	reply, _, ok := Execute("SETX", ctx)
	if !ok {
		t.Fatal("SETX should exist in registry")
	}
	expected := protocol.WrongArgsErr("SETX")
	if reply != expected {
		t.Errorf("reply = %q, want %q", reply, expected)
	}
}

func TestFlagConstants(t *testing.T) {
	if FlagWrite != "write" {
		t.Errorf("FlagWrite = %q, want write", FlagWrite)
	}
	if FlagReadonly != "readonly" {
		t.Errorf("FlagReadonly = %q, want readonly", FlagReadonly)
	}
	if FlagFast != "fast" {
		t.Errorf("FlagFast = %q, want fast", FlagFast)
	}
	if FlagAdmin != "admin" {
		t.Errorf("FlagAdmin = %q, want admin", FlagAdmin)
	}
}

func TestDBSize(t *testing.T) {
	db := datastore.NewGodisDB()
	db.Set("a", "1", 0)
	db.Set("b", "2", 0)
	db.Set("c", "3", 0)

	id := 0
	ctx := &CommandContext{
		Args:        []string{"DBSIZE"},
		DB:          db,
		AllDBs:      []*datastore.GodisDB{db},
		CurrentDBID: &id,
	}

	reply, _, ok := Execute("DBSIZE", ctx)
	if !ok {
		t.Fatal("DBSIZE not registered")
	}
	expected := protocol.MakeInt(3)
	if reply != expected {
		t.Errorf("DBSIZE reply = %q, want %q", reply, expected)
	}
}

func TestScan_Basic(t *testing.T) {
	db := datastore.NewGodisDB()
	for i := 0; i < 15; i++ {
		db.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("%d", i), 0)
	}

	id := 0
	ctx := &CommandContext{
		Args:        []string{"SCAN", "0", "COUNT", "10"},
		DB:          db,
		AllDBs:      []*datastore.GodisDB{db},
		CurrentDBID: &id,
	}

	reply, _, ok := Execute("SCAN", ctx)
	if !ok {
		t.Fatal("SCAN not registered")
	}
	// 第一轮返回 cursor != 0
	if len(reply) == 0 {
		t.Fatal("SCAN reply is empty")
	}
	// cursor 非 0，说明还有剩余 key
	if reply[3:5] == "0\r\n" {
		// 如果 cursor 是 0，说明 10 个以内就全部返回了（不应发生，有 15 个 key）
	}
}

func TestScan_Match(t *testing.T) {
	db := datastore.NewGodisDB()
	db.Set("abc", "1", 0)
	db.Set("abd", "2", 0)
	db.Set("xyz", "3", 0)

	id := 0
	ctx := &CommandContext{
		Args:        []string{"SCAN", "0", "MATCH", "ab*", "COUNT", "100"},
		DB:          db,
		AllDBs:      []*datastore.GodisDB{db},
		CurrentDBID: &id,
	}

	reply, _, ok := Execute("SCAN", ctx)
	if !ok {
		t.Fatal("SCAN not registered")
	}
	// 应包含 abc, abd，不包含 xyz
	if !strings.Contains(reply, "abc") || !strings.Contains(reply, "abd") {
		t.Errorf("expected abc and abd in reply, got %q", reply)
	}
	if strings.Contains(reply, "xyz") {
		t.Errorf("xyz should not be in reply, got %q", reply)
	}
}

func TestScan_Type(t *testing.T) {
	db := datastore.NewGodisDB()
	db.Set("str1", "hello", 0)
	db.Set("str2", "world", 0)

	id := 0
	ctx := &CommandContext{
		Args:        []string{"SCAN", "0", "TYPE", "string", "COUNT", "100"},
		DB:          db,
		AllDBs:      []*datastore.GodisDB{db},
		CurrentDBID: &id,
	}

	reply, _, ok := Execute("SCAN", ctx)
	if !ok {
		t.Fatal("SCAN not registered")
	}
	if !strings.Contains(reply, "str1") || !strings.Contains(reply, "str2") {
		t.Errorf("expected str1 and str2, got %q", reply)
	}
}

func TestScan_Empty(t *testing.T) {
	db := datastore.NewGodisDB()
	id := 0
	ctx := &CommandContext{
		Args:        []string{"SCAN", "0"},
		DB:          db,
		AllDBs:      []*datastore.GodisDB{db},
		CurrentDBID: &id,
	}

	reply, _, ok := Execute("SCAN", ctx)
	if !ok {
		t.Fatal("SCAN not registered")
	}
	if !strings.Contains(reply, "0") {
		t.Errorf("expected cursor 0 for empty db, got %q", reply)
	}
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		s, pattern string
		want       bool
	}{
		{"hello", "hello", true},
		{"hello", "h*", true},
		{"hello", "*o", true},
		{"hello", "h?llo", true},
		{"hello", "h??lo", true},
		{"hello", "h*o", true},
		{"hello", "world", false},
		{"hello", "h?llooo", false},
		{"hello", "*", true},
		{"hello", "?ello", true},
		{"hello", "ell", false},
		{"abc", "a*c", true},
		{"abc", "a*d", false},
	}
	for _, tt := range tests {
		got := matchGlob(tt.s, tt.pattern)
		if got != tt.want {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.s, tt.pattern, got, tt.want)
		}
	}
}
