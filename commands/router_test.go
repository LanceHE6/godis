package commands

import (
	"godis/datastore"
	"godis/protocol"
	"testing"
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

	Register("lower", 1, FlagWrite, 0, 0, 0, UnimplementedHandlerFunc)

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

func TestUnimplementedHandler(t *testing.T) {
	db := datastore.NewGodisDB()
	id := 0
	ctx := &CommandContext{
		Args:        []string{"DEL", "key"},
		DB:          db,
		AllDBs:      []*datastore.GodisDB{db},
		CurrentDBID: &id,
	}

	reply := UnimplementedHandlerFunc(ctx)
	expected := protocol.MakeError("DEL is not supported")
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
