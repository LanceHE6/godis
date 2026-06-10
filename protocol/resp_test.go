package protocol

import (
	"bufio"
	"strings"
	"testing"
)

func parseRESP(input string) ([]string, error) {
	return ParseRESP(bufio.NewReader(strings.NewReader(input)))
}

func TestParseRESP_SingleCommand(t *testing.T) {
	args, err := parseRESP("*1\r\n$4\r\nPING\r\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) != 1 || args[0] != "PING" {
		t.Errorf("got %v, want [PING]", args)
	}
}

func TestParseRESP_MultipleArgs(t *testing.T) {
	args, err := parseRESP("*3\r\n$3\r\nSET\r\n$5\r\nhello\r\n$5\r\nworld\r\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"SET", "hello", "world"}
	if len(args) != len(expected) {
		t.Fatalf("got %d args, want %d", len(args), len(expected))
	}
	for i, v := range expected {
		if args[i] != v {
			t.Errorf("args[%d] = %q, want %q", i, args[i], v)
		}
	}
}

func TestParseRESP_InlineCommand(t *testing.T) {
	args, err := parseRESP("PING\r\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) != 1 || args[0] != "PING" {
		t.Errorf("got %v, want [PING]", args)
	}
}

func TestParseRESP_InlineMultipleArgs(t *testing.T) {
	args, err := parseRESP("SET key value\r\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"SET", "key", "value"}
	if len(args) != len(expected) {
		t.Fatalf("got %d args, want %d", len(args), len(expected))
	}
	for i, v := range expected {
		if args[i] != v {
			t.Errorf("args[%d] = %q, want %q", i, args[i], v)
		}
	}
}

func TestParseRESP_EmptyLine(t *testing.T) {
	_, err := parseRESP("\r\n")
	if err == nil {
		t.Error("expected error for empty line")
	}
}

func TestParseRESP_InlineAsFallback(t *testing.T) {
	// 非 * 开头的行会走 inline 解析
	args, err := parseRESP("+OK\r\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) != 1 || args[0] != "+OK" {
		t.Errorf("got %v, want [+OK]", args)
	}
}

func TestParseRESP_InvalidArrayLen(t *testing.T) {
	_, err := parseRESP("*0\r\n")
	if err == nil {
		t.Error("expected error for zero array length")
	}
}

func TestParseRESP_EOF(t *testing.T) {
	_, err := parseRESP("")
	if err == nil {
		t.Error("expected error for EOF")
	}
}
