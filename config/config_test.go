package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit_GeneratesDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	if err := Init(path); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if Global.Bind != "0.0.0.0" {
		t.Errorf("Bind = %q, want 0.0.0.0", Global.Bind)
	}
	if Global.Port != 6379 {
		t.Errorf("Port = %d, want 6379", Global.Port)
	}
	if Global.Databases != 16 {
		t.Errorf("Databases = %d, want 16", Global.Databases)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Error("config file is empty")
	}
}

func TestInit_LoadsExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	content := `bind: 127.0.0.1
port: 6380
databases: 8
aof_file: /tmp/test.aof
log_file: /tmp/test.log
log_level: debug
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := Init(path); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if Global.Bind != "127.0.0.1" {
		t.Errorf("Bind = %q, want 127.0.0.1", Global.Bind)
	}
	if Global.Port != 6380 {
		t.Errorf("Port = %d, want 6380", Global.Port)
	}
	if Global.Databases != 8 {
		t.Errorf("Databases = %d, want 8", Global.Databases)
	}
	if Global.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug", Global.LogLevel)
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.yaml")

	Init(path)

	Global.Port = 9999
	Global.LogLevel = "warn"
	if err := Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// 重新加载验证
	Global = Config{}
	if err := Init(path); err != nil {
		t.Fatalf("Init after save: %v", err)
	}
	if Global.Port != 9999 {
		t.Errorf("Port after save = %d, want 9999", Global.Port)
	}
	if Global.LogLevel != "warn" {
		t.Errorf("LogLevel after save = %q, want warn", Global.LogLevel)
	}
}
