package logger

import "testing"

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"WARN", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"unknown", LevelInfo}, // 默认
		{"", LevelInfo},        // 空字符串默认
	}
	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.want {
			t.Errorf("ParseLevel(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestNewModuleLogger(t *testing.T) {
	ml := NewModuleLogger("TEST")
	if ml == nil {
		t.Fatal("NewModuleLogger returned nil")
	}
	if ml.moduleName != "TEST" {
		t.Errorf("moduleName = %q, want TEST", ml.moduleName)
	}
}

func TestModuleLogger_LogMethods(t *testing.T) {
	// 测试各日志方法不会 panic
	ml := NewModuleLogger("TEST")
	ml.Debug("test %s", "debug")
	ml.Info("test %s", "info")
	ml.Warn("test %s", "warn")
	ml.Error("test %s", "error")
}

func TestInitAndCloseLogger(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.log"

	if err := InitGlobalLogger(path, LevelDebug); err != nil {
		t.Fatalf("InitGlobalLogger: %v", err)
	}
	CloseLogSystem()
}
