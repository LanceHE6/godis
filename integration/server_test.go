package integration

import (
	"context"
	"strings"
	"testing"
)

func TestInfo(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("INFO")
	info, err := rdb.Do(ctx, "INFO").Text()
	if err != nil {
		t.Fatalf("INFO failed: %v", err)
	}
	t.Logf("INFO response length: %d bytes", len(info))
	if !strings.Contains(info, "godis_version") {
		t.Error("INFO should contain godis_version")
	}
	if !strings.Contains(info, "# Server") {
		t.Error("INFO should contain # Server section")
	}
}

func TestInfo_Keyspace(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	rdb.Set(ctx, "k1", "v1", 0)
	rdb.Set(ctx, "k2", "v2", 0)
	t.Log("SET k1, k2, then INFO keyspace")

	info, err := rdb.Do(ctx, "INFO", "keyspace").Text()
	if err != nil {
		t.Fatalf("INFO keyspace failed: %v", err)
	}
	t.Logf("INFO keyspace:\n%s", info)
	if !strings.Contains(info, "db0:keys=2") {
		t.Errorf("INFO keyspace should show db0:keys=2, got %q", info)
	}
}

func TestCommand(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("COMMAND")
	res, err := rdb.Do(ctx, "COMMAND").Slice()
	if err != nil {
		t.Fatalf("COMMAND failed: %v", err)
	}
	t.Logf("COMMAND returned %d commands", len(res))
	if len(res) == 0 {
		t.Error("COMMAND should return non-empty list")
	}
}

func TestCommand_Count(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("COMMAND COUNT")
	n, err := rdb.Do(ctx, "COMMAND", "COUNT").Int64()
	if err != nil {
		t.Fatalf("COMMAND COUNT failed: %v", err)
	}
	t.Logf("COMMAND COUNT = %d", n)
	if n < 10 {
		t.Errorf("COMMAND COUNT = %d, want >= 10", n)
	}
}

func TestConfigGet(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("CONFIG GET port")
	res, err := rdb.Do(ctx, "CONFIG", "GET", "port").StringSlice()
	if err != nil {
		t.Fatalf("CONFIG GET port failed: %v", err)
	}
	t.Logf("CONFIG GET port = %v", res)
	if len(res) != 2 || res[0] != "port" {
		t.Errorf("CONFIG GET port = %v, want [port <value>]", res)
	}
}

func TestConfigSet(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	t.Log("CONFIG SET log_level debug")
	_, err := rdb.Do(ctx, "CONFIG", "SET", "log_level", "debug").Result()
	if err != nil {
		t.Fatalf("CONFIG SET log_level failed: %v", err)
	}

	res, err := rdb.Do(ctx, "CONFIG", "GET", "log_level").StringSlice()
	if err != nil {
		t.Fatalf("CONFIG GET log_level failed: %v", err)
	}
	t.Logf("CONFIG GET log_level = %v", res)
	if len(res) != 2 || res[1] != "debug" {
		t.Errorf("CONFIG GET log_level = %v, want [log_level debug]", res)
	}

	rdb.Do(ctx, "CONFIG", "SET", "log_level", "info")
}
